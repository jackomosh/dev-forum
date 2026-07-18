package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"forum/internal/domain"
)

// AuthServiceImpl implements AuthService interface
type AuthServiceImpl struct {
	userRepo    UserRepository
	sessionRepo SessionRepository
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo UserRepository, sessionRepo SessionRepository) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (s *AuthServiceImpl) RegisterUser(ctx context.Context, username, email, password string) (*domain.User, error) {
	// Check if user exists
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         domain.UserRoleMember,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthServiceImpl) AuthenticateUser(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	if !verifyPassword(user.PasswordHash, password) {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *AuthServiceImpl) CreateSession(ctx context.Context, userID domain.UserID) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}

	session := &domain.Session{
		ID:        domain.SessionID(token),
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthServiceImpl) ValidateSession(ctx context.Context, token string) (*domain.User, error) {
	session, err := s.sessionRepo.GetByID(ctx, domain.SessionID(token))
	if err != nil {
		return nil, errors.New("invalid session")
	}
	if session == nil {
		return nil, errors.New("session not found")
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *AuthServiceImpl) DeleteSession(ctx context.Context, token string) error {
	return s.sessionRepo.Delete(ctx, domain.SessionID(token))
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
