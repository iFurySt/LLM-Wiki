package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ifuryst/llm-wiki/internal/api"
)

type obsidianManifest struct {
	ExportedAt string `json:"exported_at"`
	TenantID   string `json:"tenant_id"`
	DocCount   int    `json:"doc_count"`
}

func exportWorkspaceToObsidian(outputDir string, tenantID string, namespaces []api.NamespaceResponse, documents []api.DocumentResponse) error {
	root := strings.TrimSpace(outputDir)
	if root == "" {
		return fmt.Errorf("output directory is required")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}

	namespaceKeys := make(map[int64]string, len(namespaces))
	namespaceNames := make(map[int64]string, len(namespaces))
	for _, item := range namespaces {
		namespaceKeys[item.ID] = sanitizePathSegment(item.Key)
		namespaceNames[item.ID] = item.DisplayName
	}

	for _, item := range documents {
		nsKey := namespaceKeys[item.NamespaceID]
		if nsKey == "" {
			nsKey = "unknown"
		}
		dir := filepath.Join(root, nsKey)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		filename := sanitizePathSegment(item.Slug)
		if filename == "" {
			filename = fmt.Sprintf("document-%d", item.ID)
		}
		body := renderObsidianMarkdown(item, tenantID, nsKey, namespaceNames[item.NamespaceID])
		if err := os.WriteFile(filepath.Join(dir, filename+".md"), []byte(body), 0o644); err != nil {
			return err
		}
	}

	manifest := obsidianManifest{
		ExportedAt: time.Now().Format(time.RFC3339),
		TenantID:   tenantID,
		DocCount:   len(documents),
	}
	payload, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, ".llm-wiki-export.json"), payload, 0o644)
}

func renderObsidianMarkdown(item api.DocumentResponse, tenantID string, namespaceKey string, namespaceName string) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("llm_wiki: true\n")
	b.WriteString("tenant_id: " + yamlScalar(tenantID) + "\n")
	b.WriteString("namespace_key: " + yamlScalar(namespaceKey) + "\n")
	if strings.TrimSpace(namespaceName) != "" {
		b.WriteString("namespace_name: " + yamlScalar(namespaceName) + "\n")
	}
	b.WriteString(fmt.Sprintf("document_id: %d\n", item.ID))
	b.WriteString(fmt.Sprintf("namespace_id: %d\n", item.NamespaceID))
	b.WriteString("slug: " + yamlScalar(item.Slug) + "\n")
	b.WriteString("title: " + yamlScalar(item.Title) + "\n")
	b.WriteString("status: " + yamlScalar(item.Status) + "\n")
	b.WriteString(fmt.Sprintf("current_revision_no: %d\n", item.CurrentRevisionNo))
	b.WriteString("updated_at: " + yamlScalar(item.UpdatedAt) + "\n")
	b.WriteString("---\n\n")
	if strings.TrimSpace(item.Title) != "" {
		b.WriteString("# " + item.Title + "\n\n")
	}
	b.WriteString(item.Content)
	if !strings.HasSuffix(item.Content, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}

func sanitizePathSegment(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
		" ", "-",
	)
	value = replacer.Replace(value)
	value = strings.Trim(value, "-.")
	return value
}

func yamlScalar(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return `""`
	}
	escaped := strings.ReplaceAll(trimmed, `"`, `\"`)
	return `"` + escaped + `"`
}
