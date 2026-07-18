package handler

import (
	"context"
	"log"
	"net/http"

	"forum/internal/domain"
)

type contextKey string

const UserContextKey contextKey = "user"

// Middleware handles authentication checks
type Middleware struct {
	authService AuthService
}

// NewMiddleware creates a new auth middleware
func NewMiddleware(authService AuthService) *Middleware {
	return &Middleware{
		authService: authService,
	}
}

// RequireAuth ensures the user is authenticated
func (m *Middleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session token from cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Validate session via auth service
		user, err := m.authService.ValidateSession(r.Context(), cookie.Value)
		if err != nil {
			// Clear invalid cookie
			http.SetCookie(w, &http.Cookie{
				Name:   "session_token",
				Value:  "",
				Path:   "/",
				MaxAge: -1,
			})
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Store user in request context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		r = r.WithContext(ctx)

		// Call next handler
		next(w, r)
	}
}

// OptionalAuth checks auth but doesn't require it
func (m *Middleware) OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session token from cookie
		cookie, err := r.Cookie("session_token")
		if err == nil {
			// Validate session
			user, err := m.authService.ValidateSession(r.Context(), cookie.Value)
			if err == nil {
				// Store user in context
				ctx := context.WithValue(r.Context(), UserContextKey, user)
				r = r.WithContext(ctx)
			}
		}

		// Call next handler (even if not authenticated)
		next(w, r)
	}
}

// GetUserFromContext retrieves the user from request context
func GetUserFromContext(ctx context.Context) (*domain.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*domain.User)
	return user, ok
}

// GetPublicUserFromContext retrieves the public user from request context
func GetPublicUserFromContext(ctx context.Context) (*domain.PublicUser, bool) {
	user, ok := ctx.Value(UserContextKey).(*domain.User)
	if !ok || user == nil {
		return nil, false
	}
	return &domain.PublicUser{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}, true
}

// GlobalErrorHandler handles all HTTP errors
func GlobalErrorHandler(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	log.Printf("Error: %v - Path: %s - Method: %s", err, r.URL.Path, r.Method)
	http.Error(w, http.StatusText(statusCode), statusCode)
}
