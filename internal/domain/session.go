package domain

import "time"

type SessionID string

type Session struct {
	ID        SessionID
	UserID    UserID
	ExpiresAt time.Time
	CreatedAt time.Time
}

// IsExpired checks if the session has passed its expiration time.
func (s *Session) IsExpired() bool {
	if s == nil {
		return true
	}
	return time.Now().After(s.ExpiresAt)
}

type AuthenticatedUser struct {
	User    PublicUser
	Session Session
}
