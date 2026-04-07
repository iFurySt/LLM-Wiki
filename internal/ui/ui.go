package ui

import (
	"embed"
	"html/template"
	"strings"
)

//go:embed templates/*.html
var templateFS embed.FS

func ParseTemplates() (*template.Template, error) {
	return template.New("").Funcs(template.FuncMap{
		"snippet": func(value string, max int) string {
			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				return "No content yet."
			}
			runes := []rune(trimmed)
			if len(runes) <= max {
				return trimmed
			}
			return string(runes[:max]) + "..."
		},
	}).ParseFS(templateFS, "templates/*.html")
}
