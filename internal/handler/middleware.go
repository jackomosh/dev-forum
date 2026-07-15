package handler

import (
	"context"
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
		user, err := m.authService.ValidateSession(cookie.Value)
		if err != nil {
			// Clear invalid cookie
			http.SetCookie(w, &http.Cookie{
				Name:   "session_token",
				Value:  "",
				Path:   "/",
				MaxAge: -1,
			})
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Store user in request context
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		r = r.WithContext(ctx)

		// Call next handler
		next(w, r)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

// OptionalAuth checks auth but doesn't require it
func (m *Middleware) OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session token from cookie
		cookie, err := r.Cookie("session_token")
		if err == nil {
			// Validate session
			user, err := m.authService.ValidateSession(cookie.Value)
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
	return user, okpackage handler

import "forum/internal/domain"

type RequestContext struct {
	RequestID string
	User      domain.PublicUser
	SessionID domain.SessionID
	CSRFToken string
}

type FlashMessage struct {
	Kind    string
	Message string
}

}

// GlobalErrorHandler handles all HTTP errors
func GlobalErrorHandler(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	log.Printf("Error: %v - Path: %s - Method: %s", err, r.URL.Path, r.Method)
	http.Error(w, http.StatusText(statusCode), statusCode)
}