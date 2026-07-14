package handler

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type Renderer struct {
	templateDir string
}

func NewRenderer(templateDir string) *Renderer {
	return &Renderer{templateDir: templateDir}
}

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
