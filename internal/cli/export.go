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
	NS         string `json:"ns"`
	DocCount   int    `json:"doc_count"`
}

func exportNSToObsidian(outputDir string, ns string, folders []api.FolderResponse, documents []api.DocumentResponse) error {
	root := strings.TrimSpace(outputDir)
	if root == "" {
		return fmt.Errorf("output directory is required")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}

	namespaceKeys := make(map[int64]string, len(folders))
	namespaceNames := make(map[int64]string, len(folders))
	for _, item := range folders {
		namespaceKeys[item.ID] = sanitizePathSegment(item.Key)
		namespaceNames[item.ID] = item.DisplayName
	}

	for _, item := range documents {
		nsKey := namespaceKeys[item.FolderID]
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
		body := renderObsidianMarkdown(item, ns, nsKey, namespaceNames[item.FolderID])
		if err := os.WriteFile(filepath.Join(dir, filename+".md"), []byte(body), 0o644); err != nil {
			return err
		}
	}

	manifest := obsidianManifest{
		ExportedAt: time.Now().Format(time.RFC3339),
		NS:         ns,
		DocCount:   len(documents),
	}
	payload, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, ".llm-wiki-export.json"), payload, 0o644)
}

func renderObsidianMarkdown(item api.DocumentResponse, ns string, folderKey string, folderName string) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("llm_wiki: true\n")
	b.WriteString("ns: " + yamlScalar(ns) + "\n")
	b.WriteString("folder_key: " + yamlScalar(folderKey) + "\n")
	if strings.TrimSpace(folderName) != "" {
		b.WriteString("folder_name: " + yamlScalar(folderName) + "\n")
	}
	b.WriteString(fmt.Sprintf("document_id: %d\n", item.ID))
	b.WriteString(fmt.Sprintf("folder_id: %d\n", item.FolderID))
	b.WriteString("slug: " + yamlScalar(item.Slug) + "\n")
	b.WriteString("title: " + yamlScalar(item.Title) + "\n")
	b.WriteString("status: " + yamlScalar(item.Status) + "\n")
	b.WriteString(fmt.Sprintf("current_revision_no: %d\n", item.CurrentRevisionNo))
	b.WriteString("updated_at: " + yamlScalar(item.UpdatedAt) + "\n")
	if item.Source != nil {
		if item.Source.ID != "" {
			b.WriteString("source_id: " + yamlScalar(item.Source.ID) + "\n")
		}
		if item.Source.Label != "" {
			b.WriteString("source_label: " + yamlScalar(item.Source.Label) + "\n")
		}
		if item.Source.Category != "" {
			b.WriteString("source_category: " + yamlScalar(item.Source.Category) + "\n")
		}
		if item.Source.InputMode != "" {
			b.WriteString("source_input_mode: " + yamlScalar(item.Source.InputMode) + "\n")
		}
		if item.Source.OriginalRef != "" {
			b.WriteString("source_original_ref: " + yamlScalar(item.Source.OriginalRef) + "\n")
		}
		if item.Source.CapturedAt != "" {
			b.WriteString("source_captured_at: " + yamlScalar(item.Source.CapturedAt) + "\n")
		}
		if item.Source.ContentType != "" {
			b.WriteString("source_content_type: " + yamlScalar(item.Source.ContentType) + "\n")
		}
		if item.Source.Adapter != "" {
			b.WriteString("source_adapter: " + yamlScalar(item.Source.Adapter) + "\n")
		}
	}
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
