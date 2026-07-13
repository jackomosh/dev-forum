package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// Global mock variables to control frontend states during test runs
type PageData struct {
	User  interface{} // Set to a struct value to test logged-in state
	Error string      // Set to test error alert boxes
}

// Helper to safely render templates relative to the project root
func render(w http.ResponseWriter, tmpl string, data PageData) {
	// Look up templates relative to project root
	files := []string{
		filepath.Join("web", "templates", "base.html"),
		filepath.Join("web", "templates", tmpl),
	}

	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Printf("Template compilation error: %v", err)
		http.Error(w, "Failed to compile template files: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// Serve static files (CSS, media) from "web/static" relative to the root directory
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route: Home Page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		// Try setting User: struct{ Username string }{"jacomondi"} to test logged-in view
		render(w, "index.html", PageData{User: nil})
	})

	// Route: Login Page (With centering styled auth-card)
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		// Mock error state if query parameter "?err=true" is passed
		errMsg := ""
		if r.URL.Query().Get("err") == "true" {
			errMsg = "Invalid username or password credentials."
		}
		render(w, "login.html", PageData{Error: errMsg})
	})

	// Route: Registration Page
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		render(w, "register.html", PageData{})
	})

	// Route: Forgot Password Page
	http.HandleFunc("/forgot-password", func(w http.ResponseWriter, r *http.Request) {
		render(w, "forgot-password.html", PageData{})
	})

	log.Println("=== Front-End Mock Dev Server ===")
	log.Println("Server running at: http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}