package http

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type WebHandler struct {
	templates *template.Template
}

func NewWebHandler(templatesDir string) *WebHandler {
	tpl, err := template.ParseGlob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}

	return &WebHandler{
		templates: tpl,
	}
}

func (h *WebHandler) ServeApp(w http.ResponseWriter, r *http.Request) {
	err := h.templates.ExecuteTemplate(w, "app.html", nil)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

