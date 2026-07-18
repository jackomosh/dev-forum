package handler

import (
	"net/http"
	"strings"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService AuthService
	renderer    *Renderer
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService AuthService, renderer *Renderer) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		renderer:    renderer,
	}
}

// HandleRegisterPage displays the registration form
func (h *AuthHandler) HandleRegisterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := AuthViewData{
		BaseViewData: BaseViewData{
			Error: "",
		},
		Form: RegisterRequest{},
	}

	h.renderer.Render(w, "register.html", data)
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

	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	// Validate input
	if username == "" || email == "" || password == "" {
		data := AuthViewData{
			BaseViewData: BaseViewData{
				Error: "All fields are required",
			},
			Form: RegisterRequest{
				Username: username,
				Email:    email,
			},
		}
		h.renderer.Render(w, "register.html", data)
		return
	}

	if password != confirmPassword {
		data := AuthViewData{
			BaseViewData: BaseViewData{
				Error: "Passwords do not match",
			},
			Form: RegisterRequest{
				Username: username,
				Email:    email,
			},
		}
		h.renderer.Render(w, "register.html", data)
		return
	}

	// Call auth service to register user
	user, err := h.authService.RegisterUser(r.Context(), username, email, password)
	if err != nil {
		data := AuthViewData{
			BaseViewData: BaseViewData{
				Error: err.Error(),
			},
			Form: RegisterRequest{
				Username: username,
				Email:    email,
			},
		}
		h.renderer.Render(w, "register.html", data)
		return
	}

	// Auto-login after registration
	sessionToken, err := h.authService.CreateSession(r.Context(), user.ID)
	if err != nil {
		data := AuthViewData{
			BaseViewData: BaseViewData{
				Error: "Registration successful but login failed. Please login manually.",
			},
			Form: RegisterRequest{
				Username: username,
				Email:    email,
			},
		}
		h.renderer.Render(w, "register.html", data)
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

	data := AuthViewData{
		BaseViewData: BaseViewData{
			Error: "",
		},
		Form: RegisterRequest{},
	}

	h.renderer.Render(w, "login.html", data)
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

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if email == "" || password == "" {
		data := AuthViewData{
			BaseViewData: BaseViewData{
				Error: "Email and password are required",
			},
			Form: RegisterRequest{Email: email},
		}
		h.renderer.Render(w, "login.html", data)
		return
	}

	// Call auth service to authenticate user
	user, err := h.authService.AuthenticateUser(r.Context(), email, password)
	if err != nil {
		data := AuthViewData{
			BaseViewData: BaseViewData{
				Error: "Invalid email or password",
			},
			Form: RegisterRequest{Email: email},
		}
		h.renderer.Render(w, "login.html", data)
		return
	}

	// Create session
	sessionToken, err := h.authService.CreateSession(r.Context(), user.ID)
	if err != nil {
		data := AuthViewData{
			BaseViewData: BaseViewData{
				Error: "Failed to create session. Please try again.",
			},
			Form: RegisterRequest{Email: email},
		}
		h.renderer.Render(w, "login.html", data)
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
		if err := h.authService.DeleteSession(r.Context(), cookie.Value); err != nil {
			// Log error but continue
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
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   24 * 60 * 60,
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
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}
