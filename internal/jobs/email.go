package jobs

import (
	"context"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/services"
	"github.com/edsonmichaque/bazaruto/pkg/job"
	"github.com/go-mail/mail/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SendEmailJob represents a job for sending emails
type SendEmailJob struct {
	ID        uuid.UUID `json:"id"`
	To        string    `json:"to"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Template  string    `json:"template"`
	From      string    `json:"from"`
	Attempts  int       `json:"attempts"`
	RunAtTime time.Time `json:"run_at_time"`

	// SMTP Configuration (can be injected or from config)
	SMTPHost     string `json:"smtp_host,omitempty"`
	SMTPPort     int    `json:"smtp_port,omitempty"`
	SMTPUsername string `json:"smtp_username,omitempty"`
	SMTPPassword string `json:"smtp_password,omitempty"`
	SMTPTLS      bool   `json:"smtp_tls,omitempty"`
}

// Perform executes the email sending job
func (j *SendEmailJob) Perform(ctx context.Context) error {
	log := logger.NewLogger("info", "json")

	// Validate email address
	if j.To == "" || !strings.Contains(j.To, "@") {
		return fmt.Errorf("invalid email address: %s", j.To)
	}

	// Set default from address if not provided
	from := j.From
	if from == "" {
		from = "noreply@bazaruto.com"
	}

	// Send email using go-mail
	if err := j.sendEmailWithGoMail(ctx, from, j.To, j.Subject, j.Body); err != nil {
		log.Error("Failed to send email",
			zap.Error(err),
			zap.String("to", j.To),
			zap.String("subject", j.Subject),
			zap.Int("attempt", j.Attempts))
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Info("Email sent successfully",
		zap.String("to", j.To),
		zap.String("subject", j.Subject),
		zap.Int("attempt", j.Attempts))

	return nil
}

// sendEmailWithGoMail sends email using the go-mail library
func (j *SendEmailJob) sendEmailWithGoMail(ctx context.Context, from, to, subject, body string) error {
	// Set default SMTP configuration if not provided
	smtpHost := j.SMTPHost
	if smtpHost == "" {
		smtpHost = "localhost" // Default to localhost for development
	}

	smtpPort := j.SMTPPort
	if smtpPort == 0 {
		smtpPort = 1025 // Default to 1025 for development (MailHog, etc.)
	}

	// Create a new message
	m := mail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Create a new dialer
	d := mail.NewDialer(smtpHost, smtpPort, j.SMTPUsername, j.SMTPPassword)

	// Configure TLS
	if j.SMTPTLS {
		d.StartTLSPolicy = mail.MandatoryStartTLS
	} else {
		d.StartTLSPolicy = mail.NoStartTLS
	}

	// Set timeout for context cancellation
	d.Timeout = 30 * time.Second

	// For development/testing, simulate email sending if no real SMTP is configured
	if smtpHost == "localhost" && j.SMTPUsername == "" {
		// Simulate network delay
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}

		// Simulate occasional failures for testing retry logic
		if j.Attempts == 1 && strings.Contains(to, "fail@example.com") {
			return fmt.Errorf("simulated SMTP failure")
		}

		return nil
	}

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	return nil
}

// Queue returns the queue name for this job
func (j *SendEmailJob) Queue() string {
	return job.QueueMailers
}

// MaxRetries returns the maximum number of retries for this job
func (j *SendEmailJob) MaxRetries() int {
	return 3
}

// RetryBackoff returns the backoff duration for retries
func (j *SendEmailJob) RetryBackoff() time.Duration {
	return time.Duration(j.Attempts) * 30 * time.Second
}

// Priority returns the priority of this job
func (j *SendEmailJob) Priority() int {
	return 5 // Medium priority
}

// WelcomeEmailJob represents a job for sending welcome emails to new users
type WelcomeEmailJob struct {
	ID          uuid.UUID             `json:"id"`
	UserID      uuid.UUID             `json:"user_id"`
	UserService *services.UserService `json:"-"` // Injected dependency
	Attempts    int                   `json:"attempts"`
	RunAtTime   time.Time             `json:"run_at_time"`
}

// Perform executes the welcome email job
func (j *WelcomeEmailJob) Perform(ctx context.Context) error {
	log := logger.NewLogger("info", "json")

	// Fetch user details from database
	user, err := j.UserService.GetUser(ctx, j.UserID)
	if err != nil {
		log.Error("Failed to fetch user for welcome email",
			zap.Error(err),
			zap.String("user_id", j.UserID.String()))
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found: %s", j.UserID.String())
	}

	// Generate personalized welcome email content
	subject := "Welcome to Bazaruto Insurance!"
	body, err := j.generateWelcomeEmailBody(user)
	if err != nil {
		log.Error("Failed to generate welcome email body",
			zap.Error(err),
			zap.String("user_id", j.UserID.String()))
		return fmt.Errorf("failed to generate email body: %w", err)
	}

	// Create and dispatch email job
	emailJob := &SendEmailJob{
		ID:        uuid.New(),
		To:        user.Email,
		Subject:   subject,
		Body:      body,
		From:      "welcome@bazaruto.com",
		Attempts:  j.Attempts,
		RunAtTime: time.Now(),
	}

	// In a real implementation, you would dispatch this job to the email queue
	// For now, we'll execute it directly
	if err := emailJob.Perform(ctx); err != nil {
		log.Error("Failed to send welcome email",
			zap.Error(err),
			zap.String("user_id", j.UserID.String()),
			zap.String("email", user.Email))
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	log.Info("Welcome email sent successfully",
		zap.String("user_id", j.UserID.String()),
		zap.String("email", user.Email))

	return nil
}

// Queue returns the queue name for this job
func (j *WelcomeEmailJob) Queue() string {
	return job.QueueMailers
}

// MaxRetries returns the maximum number of retries for this job
func (j *WelcomeEmailJob) MaxRetries() int {
	return 3
}

// RetryBackoff returns the backoff duration for retries
func (j *WelcomeEmailJob) RetryBackoff() time.Duration {
	return time.Duration(j.Attempts) * 30 * time.Second
}

// Priority returns the priority of this job
func (j *WelcomeEmailJob) Priority() int {
	return 3 // High priority for welcome emails
}

// generateWelcomeEmailBody generates the HTML body for the welcome email
func (j *WelcomeEmailJob) generateWelcomeEmailBody(user *models.User) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome to Bazaruto Insurance</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #2c3e50; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
        .button { display: inline-block; padding: 12px 24px; background-color: #3498db; color: white; text-decoration: none; border-radius: 4px; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to Bazaruto Insurance!</h1>
        </div>
        <div class="content">
            <h2>Hello {{.FullName}}!</h2>
            <p>Welcome to Bazaruto Insurance, your trusted partner for comprehensive insurance solutions.</p>
            <p>We're excited to have you on board and look forward to helping you protect what matters most.</p>
            
            <h3>What's next?</h3>
            <ul>
                <li>Complete your profile setup</li>
                <li>Explore our insurance products</li>
                <li>Get your first quote</li>
                <li>Connect with our support team</li>
            </ul>
            
            <p>
                <a href="https://bazaruto.com/dashboard" class="button">Go to Dashboard</a>
            </p>
            
            <p>If you have any questions, don't hesitate to reach out to our support team.</p>
        </div>
        <div class="footer">
            <p>© 2024 Bazaruto Insurance. All rights reserved.</p>
            <p>This email was sent to {{.Email}}</p>
        </div>
    </div>
</body>
</html>`

	t, err := template.New("welcome").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := t.Execute(&buf, user); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// PasswordResetJob represents a job for sending password reset emails
type PasswordResetJob struct {
	ID          uuid.UUID             `json:"id"`
	UserID      uuid.UUID             `json:"user_id"`
	Token       string                `json:"token"`
	UserService *services.UserService `json:"-"` // Injected dependency
	Attempts    int                   `json:"attempts"`
	RunAtTime   time.Time             `json:"run_at_time"`
}

// Perform executes the password reset email job
func (j *PasswordResetJob) Perform(ctx context.Context) error {
	log := logger.NewLogger("info", "json")

	// Fetch user details from database
	user, err := j.UserService.GetUser(ctx, j.UserID)
	if err != nil {
		log.Error("Failed to fetch user for password reset email",
			zap.Error(err),
			zap.String("user_id", j.UserID.String()))
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found: %s", j.UserID.String())
	}

	// Generate password reset email content
	subject := "Reset Your Bazaruto Insurance Password"
	body, err := j.generatePasswordResetEmailBody(user, j.Token)
	if err != nil {
		log.Error("Failed to generate password reset email body",
			zap.Error(err),
			zap.String("user_id", j.UserID.String()))
		return fmt.Errorf("failed to generate email body: %w", err)
	}

	// Create and dispatch email job
	emailJob := &SendEmailJob{
		To:      user.Email,
		Subject: subject,
		Body:    body,
		From:    "security@bazaruto.com",
	}

	// Execute email job
	if err := emailJob.Perform(ctx); err != nil {
		log.Error("Failed to send password reset email",
			zap.Error(err),
			zap.String("user_id", j.UserID.String()),
			zap.String("email", user.Email))
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	log.Info("Password reset email sent successfully",
		zap.String("user_id", j.UserID.String()),
		zap.String("email", user.Email))

	return nil
}

// generatePasswordResetEmailBody generates the HTML body for the password reset email
func (j *PasswordResetJob) generatePasswordResetEmailBody(user *models.User, token string) (string, error) {
	resetURL := fmt.Sprintf("https://bazaruto.com/reset-password?token=%s", token)

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Your Password - Bazaruto Insurance</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #e74c3c; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
        .button { display: inline-block; padding: 12px 24px; background-color: #e74c3c; color: white; text-decoration: none; border-radius: 4px; margin: 10px 0; }
        .warning { background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; border-radius: 4px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Request</h1>
        </div>
        <div class="content">
            <h2>Hello {{.FullName}}!</h2>
            <p>We received a request to reset your password for your Bazaruto Insurance account.</p>
            
            <p>Click the button below to reset your password:</p>
            <p>
                <a href="{{.ResetURL}}" class="button">Reset Password</a>
            </p>
            
            <div class="warning">
                <strong>Security Notice:</strong>
                <ul>
                    <li>This link will expire in 24 hours</li>
                    <li>If you didn't request this reset, please ignore this email</li>
                    <li>Never share your password with anyone</li>
                </ul>
            </div>
            
            <p>If the button doesn't work, copy and paste this link into your browser:</p>
            <p style="word-break: break-all; background-color: #f0f0f0; padding: 10px; border-radius: 4px;">
                {{.ResetURL}}
            </p>
        </div>
        <div class="footer">
            <p>© 2024 Bazaruto Insurance. All rights reserved.</p>
            <p>This email was sent to {{.Email}}</p>
        </div>
    </div>
</body>
</html>`

	t, err := template.New("password_reset").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := struct {
		*models.User
		ResetURL string
	}{
		User:     user,
		ResetURL: resetURL,
	}

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
