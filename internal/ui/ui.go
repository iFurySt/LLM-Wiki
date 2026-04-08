package ui

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	rendererhtml "github.com/yuin/goldmark/renderer/html"
)

//go:embed templates/*.html assets/*
var embeddedFS embed.FS

func ParseTemplates() (*template.Template, error) {
	markdown := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(rendererhtml.WithHardWraps()),
	)

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
		"markdown": func(value string) template.HTML {
			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				return template.HTML("<p>No content yet.</p>")
			}

			var buf bytes.Buffer
			if err := markdown.Convert([]byte(trimmed), &buf); err != nil {
				return template.HTML("<p>Failed to render markdown.</p>")
			}
			return template.HTML(buf.String())
		},
	}).ParseFS(embeddedFS, "templates/*.html")
}

func AssetFS() (http.FileSystem, error) {
	sub, err := fs.Sub(embeddedFS, "assets")
	if err != nil {
		return nil, err
	}
	return http.FS(sub), nil
}
