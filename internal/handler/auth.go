package handler

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"forum/internal/domain"
	"forum/internal/repository"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	authService AuthService
	templates   *template.Template
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	authService AuthService,
	templates *template.Template,
) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		authService: authService,
		templates:   templates,
	}
}

// AuthPageData contains data for auth page rendering
type AuthPageData struct {
	Error       string
	Success     string
	Username    string
	Email       string
	IsLoggedIn  bool
	CurrentUser *domain.User
}

// HandleRegisterPage displays the registration form
func (h *AuthHandler) HandleRegisterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := AuthPageData{
		Error:       "",
		Success:     "",
		IsLoggedIn:  false,
		CurrentUser: nil,
	}

	h.renderTemplate(w, "register.html", data)
}

// HandleRegister processes registration form submission
func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	// Validate input
	if username == "" || email == "" || password == "" {
		data := AuthPageData{
			Error:    "All fields are required",
			Username: username,
			Email:    email,
		}
		h.renderTemplate(w, "register.html", data)
		return
	}

	if password != confirmPassword {
		data := AuthPageData{
			Error:    "Passwords do not match",
			Username: username,
			Email:    email,
		}
		h.renderTemplate(w, "register.html", data)
		return
	}

	// Call auth service to register user
	user, err := h.authService.RegisterUser(username, email, password)
	if err != nil {
		data := AuthPageData{
			Error:    err.Error(),
			Username: username,
			Email:    email,
		}
		h.renderTemplate(w, "register.html", data)
		return
	}

	// Auto-login after registration
	sessionToken, err := h.authService.CreateSession(user.ID)
	if err != nil {
		log.Printf("Error creating session: %v", err)
		data := AuthPageData{
			Error:    "Registration successful but login failed. Please login manually.",
			Username: username,
			Email:    email,
		}
		h.renderTemplate(w, "register.html", data)
		return
	}

	// Set session cookie
	h.setSessionCookie(w, sessionToken)

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// HandleLoginPage displays the login form
func (h *AuthHandler) HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := AuthPageData{
		Error:       "",
		Success:     "",
		IsLoggedIn:  false,
		CurrentUser: nil,
	}

	h.renderTemplate(w, "login.html", data)
}

// HandleLogin processes login form submission
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		data := AuthPageData{
			Error: "Email and password are required",
			Email: email,
		}
		h.renderTemplate(w, "login.html", data)
		return
	}

	// Call auth service to authenticate user
	user, err := h.authService.AuthenticateUser(email, password)
	if err != nil {
		data := AuthPageData{
			Error: "Invalid email or password",
			Email: email,
		}
		h.renderTemplate(w, "login.html", data)
		return
	}

	// Create session
	sessionToken, err := h.authService.CreateSession(user.ID)
	if err != nil {
		log.Printf("Error creating session: %v", err)
		data := AuthPageData{
			Error: "Failed to create session. Please try again.",
			Email: email,
		}
		h.renderTemplate(w, "login.html", data)
		return
	}

	// Set session cookie
	h.setSessionCookie(w, sessionToken)

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// HandleLogout processes user logout
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	if err == nil {
		// Delete session via auth service
		if err := h.authService.DeleteSession(cookie.Value); err != nil {
			log.Printf("Error deleting session: %v", err)
		}
	}

	// Clear cookie
	h.clearSessionCookie(w)

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// setSessionCookie sets the session cookie
func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   24 * 60 * 60, // 24 hours
	}
	http.SetCookie(w, cookie)
}

// clearSessionCookie clears the session cookie
func (h *AuthHandler) clearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}

// renderTemplate renders an HTML template
func (h *AuthHandler) renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	if err := h.templates.ExecuteTemplate(w, templateName, data); err != nil {
		log.Printf("Error rendering template %s: %v", templateName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}