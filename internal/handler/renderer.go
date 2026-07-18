package handler

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// Renderer handles template rendering
type Renderer struct {
	templateDir string
}

// NewRenderer creates a new template renderer
func NewRenderer(templateDir string) *Renderer {
	return &Renderer{templateDir: templateDir}
}

// Render renders a template with the provided data
func (r *Renderer) Render(w http.ResponseWriter, tmpl string, data interface{}) {
	files := []string{
		filepath.Join(r.templateDir, "base.html"),
		filepath.Join(r.templateDir, tmpl),
	}

	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Printf("template parse error: %v", err)
		http.Error(w, "failed to compile template", http.StatusInternalServerError)
		return
	}

	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("template execute error: %v", err)
		http.Error(w, "failed to render template", http.StatusInternalServerError)
	}
}

// serverError logs and returns a 500 error
func (r *Renderer) serverError(w http.ResponseWriter, err error) {
	log.Printf("server error: %v", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
