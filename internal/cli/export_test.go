package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ifuryst/llm-wiki/internal/api"
)

func TestRenderObsidianMarkdownIncludesFrontmatterAndBody(t *testing.T) {
	doc := api.DocumentResponse{
		ID:                42,
		FolderID:          9,
		Slug:              "hello-world",
		Title:             "Hello World",
		Content:           "Body text",
		Status:            "active",
		CurrentRevisionNo: 3,
		UpdatedAt:         "2026-04-08T10:00:00Z",
	}

	got := renderObsidianMarkdown(doc, "team-alpha", "product-docs", "Product Docs")

	for _, want := range []string{
		"llm_wiki: true",
		`ns: "team-alpha"`,
		`folder_key: "product-docs"`,
		`folder_name: "Product Docs"`,
		"document_id: 42",
		`slug: "hello-world"`,
		"# Hello World",
		"Body text\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected markdown to contain %q, got:\n%s", want, got)
		}
	}
}

func TestExportNSToObsidianWritesFolderTreeAndManifest(t *testing.T) {
	dir := t.TempDir()
	folders := []api.FolderResponse{
		{ID: 1, Key: "Product Docs", DisplayName: "Product Docs"},
		{ID: 2, Key: "Engineering/Spec", DisplayName: "Engineering Spec"},
	}
	documents := []api.DocumentResponse{
		{ID: 10, FolderID: 1, Slug: "Vision Doc", Title: "Vision", Content: "North star", Status: "active", CurrentRevisionNo: 1, UpdatedAt: "2026-04-08T10:00:00Z"},
		{ID: 11, FolderID: 2, Slug: "API:v1", Title: "API", Content: "Protocol", Status: "active", CurrentRevisionNo: 2, UpdatedAt: "2026-04-08T11:00:00Z"},
	}

	if err := exportNSToObsidian(dir, "team-alpha", folders, documents); err != nil {
		t.Fatalf("exportNSToObsidian returned error: %v", err)
	}

	visionPath := filepath.Join(dir, "product-docs", "vision-doc.md")
	payload, err := os.ReadFile(visionPath)
	if err != nil {
		t.Fatalf("read exported doc: %v", err)
	}
	if !strings.Contains(string(payload), "# Vision") {
		t.Fatalf("expected exported markdown heading, got:\n%s", string(payload))
	}

	apiPath := filepath.Join(dir, "engineering-spec", "api-v1.md")
	if _, err := os.Stat(apiPath); err != nil {
		t.Fatalf("expected sanitized export path %s: %v", apiPath, err)
	}

	manifestPath := filepath.Join(dir, ".llm-wiki-export.json")
	manifest, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	for _, want := range []string{`"ns": "team-alpha"`, `"doc_count": 2`} {
		if !strings.Contains(string(manifest), want) {
			t.Fatalf("expected manifest to contain %q, got:\n%s", want, string(manifest))
		}
	}
}

func TestSanitizePathSegment(t *testing.T) {
	cases := map[string]string{
		" Product Docs ":   "product-docs",
		"API:v1":           "api-v1",
		"engineering/spec": "engineering-spec",
		`name with "bad"`:  "name-with--bad",
		"...already.clean": "already.clean",
		"   ":              "",
	}
	for input, want := range cases {
		if got := sanitizePathSegment(input); got != want {
			t.Fatalf("sanitizePathSegment(%q) = %q, want %q", input, got, want)
		}
	}
}
