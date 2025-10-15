package authentication

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"math/big"
	"strings"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// MFAService handles multi-factor authentication operations
type MFAService struct {
	issuer string
}

// NewMFAService creates a new MFA service
func NewMFAService(issuer string) *MFAService {
	return &MFAService{issuer: issuer}
}

// MFAEnrollment represents MFA enrollment data
type MFAEnrollment struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// GenerateSecret generates a new TOTP secret for a user
func (m *MFAService) GenerateSecret(userID, email string) (*MFAEnrollment, error) {
	// Generate a random secret
	secret, err := totp.Generate(totp.GenerateOpts{
		Issuer:      m.issuer,
		AccountName: email,
		SecretSize:  32,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	// Generate backup codes
	backupCodes, err := m.generateBackupCodes()
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	return &MFAEnrollment{
		Secret:      secret.Secret(),
		QRCodeURL:   secret.URL(),
		BackupCodes: backupCodes,
	}, nil
}

// VerifyTOTP verifies a TOTP code against a secret
func (m *MFAService) VerifyTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

// VerifyBackupCode verifies a backup code
func (m *MFAService) VerifyBackupCode(code string, usedCodes []string) bool {
	// Check if code is in the used codes list
	for _, used := range usedCodes {
		if used == code {
			return false // Code already used
		}
	}

	// Basic validation - backup codes should be 8 characters long
	return len(code) == 8 && m.isValidBackupCode(code)
}

// isValidBackupCode validates backup code format
func (m *MFAService) isValidBackupCode(code string) bool {
	// Backup codes should contain only alphanumeric characters
	for _, char := range code {
		if !((char >= '0' && char <= '9') || (char >= 'A' && char <= 'Z')) {
			return false
		}
	}
	return true
}

// generateBackupCodes generates backup codes for MFA
func (m *MFAService) generateBackupCodes() ([]string, error) {
	const codeCount = 10
	const codeLength = 8

	codes := make([]string, codeCount)

	for i := 0; i < codeCount; i++ {
		code, err := m.generateRandomCode(codeLength)
		if err != nil {
			return nil, err
		}
		codes[i] = code
	}

	return codes, nil
}

// generateRandomCode generates a random alphanumeric code
func (m *MFAService) generateRandomCode(length int) (string, error) {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	code := make([]byte, length)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[num.Int64()]
	}

	return string(code), nil
}

// GenerateQRCodeData generates QR code data for TOTP setup
func (m *MFAService) GenerateQRCodeData(secret, email string) (string, error) {
	key, err := otp.NewKeyFromURL(fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
		m.issuer, email, secret, m.issuer))
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code data: %w", err)
	}

	return key.URL(), nil
}

// ValidateSecret validates a TOTP secret format
func (m *MFAService) ValidateSecret(secret string) error {
	if secret == "" {
		return fmt.Errorf("secret cannot be empty")
	}

	// Remove spaces and convert to uppercase
	secret = strings.ReplaceAll(secret, " ", "")
	secret = strings.ToUpper(secret)

	// Check if it's valid base32
	_, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return fmt.Errorf("invalid secret format: %w", err)
	}

	return nil
}
