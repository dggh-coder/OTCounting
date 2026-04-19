package web

import (
	"html/template"
	"net/http"
)

type PageHandler struct {
	tpl *template.Template
}

func NewPageHandler(templatePath string) (*PageHandler, error) {
	tpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, err
	}
	return &PageHandler{tpl: tpl}, nil
}

func (h *PageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	_ = h.tpl.Execute(w, nil)
}
