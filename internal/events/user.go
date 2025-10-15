package events

import (
	"time"

	"github.com/google/uuid"
)

// UserRegisteredEvent is published when a new user registers.
type UserRegisteredEvent struct {
	*BaseBusinessEvent
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	FullName string    `json:"full_name"`
	Role     string    `json:"role"`
}

// NewUserRegisteredEvent creates a new user registered event.
func NewUserRegisteredEvent(userID uuid.UUID, email, fullName, role string) *UserRegisteredEvent {
	event := &UserRegisteredEvent{
		BaseBusinessEvent: &BaseBusinessEvent{
			EventID:       uuid.New(),
			EventType:     "user.registered",
			EntityID:      userID,
			EntityType:    "user",
			Timestamp:     time.Now(),
			EventVersion:  "1.0",
			EventMetadata: make(map[string]interface{}),
		},
		UserID:   userID,
		Email:    email,
		FullName: fullName,
		Role:     role,
	}
	return event
}

// UserLoggedInEvent is published when a user logs in.
type UserLoggedInEvent struct {
	*BaseBusinessEvent
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	LoginTime time.Time `json:"login_time"`
	IPAddress string    `json:"ip_address"`
}

// NewUserLoggedInEvent creates a new user logged in event.
func NewUserLoggedInEvent(userID uuid.UUID, email string, loginTime time.Time, ipAddress string) *UserLoggedInEvent {
	event := &UserLoggedInEvent{
		BaseBusinessEvent: &BaseBusinessEvent{
			EventID:       uuid.New(),
			EventType:     "user.logged_in",
			EntityID:      userID,
			EntityType:    "user",
			Timestamp:     time.Now(),
			EventVersion:  "1.0",
			EventMetadata: make(map[string]interface{}),
		},
		UserID:    userID,
		Email:     email,
		LoginTime: loginTime,
		IPAddress: ipAddress,
	}
	return event
}
