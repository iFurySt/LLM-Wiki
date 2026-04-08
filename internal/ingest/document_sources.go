package ingest

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/ifuryst/llm-wiki/internal/api"
	"github.com/ifuryst/llm-wiki/internal/version"
)

const MaxImportBytes = 10 << 20

type SourceSpec struct {
	ID        string
	Label     string
	Category  string
	InputMode string
	Adapter   string
}

type ImportedDocument struct {
	Title   string
	Content string
	Source  *api.DocumentSource
}

var (
	SourcePlainText = SourceSpec{ID: "plain_text", Label: "plain text", Category: "core_builtin", InputMode: "text"}
	SourceFile      = SourceSpec{ID: "local_document", Label: "local file", Category: "core_builtin", InputMode: "file"}
	SourcePDF       = SourceSpec{ID: "local_pdf", Label: "PDF", Category: "core_builtin", InputMode: "file"}
	SourceWeb       = SourceSpec{ID: "web_article", Label: "web article", Category: "optional_adapter", InputMode: "url", Adapter: "generic_fetch"}
	SourceX         = SourceSpec{ID: "x_twitter", Label: "X/Twitter", Category: "optional_adapter", InputMode: "url", Adapter: "generic_fetch"}
	SourceWechat    = SourceSpec{ID: "wechat_article", Label: "WeChat article", Category: "optional_adapter", InputMode: "url", Adapter: "generic_fetch"}
	SourceYouTube   = SourceSpec{ID: "youtube_video", Label: "YouTube", Category: "optional_adapter", InputMode: "url", Adapter: "generic_fetch"}
	SourceZhihu     = SourceSpec{ID: "zhihu_article", Label: "Zhihu", Category: "optional_adapter", InputMode: "url", Adapter: "generic_fetch"}
	SourceXHS       = SourceSpec{ID: "xiaohongshu_post", Label: "Xiaohongshu", Category: "manual_only", InputMode: "url"}
)

func SourceFromInlineText() *api.DocumentSource {
	return sourceMetadata(SourcePlainText, "inline", "text/plain; charset=utf-8")
}

func ImportFile(path string) (ImportedDocument, error) {
	source, err := detectFileSource(path)
	if err != nil {
		return ImportedDocument{}, err
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		return ImportedDocument{}, err
	}
	if !utf8.Valid(payload) {
		return ImportedDocument{}, fmt.Errorf("file %q is not valid UTF-8 text; use a text export or add a source-specific extractor first", path)
	}

	return ImportedDocument{
		Title:   humanizeName(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))),
		Content: string(payload),
		Source:  sourceMetadata(source, path, "text/plain; charset=utf-8"),
	}, nil
}

func ImportURL(rawURL string, timeout time.Duration, sourceHint string) (ImportedDocument, error) {
	source, err := resolveURLSource(rawURL, sourceHint)
	if err != nil {
		return ImportedDocument{}, err
	}
	if source.Category == "manual_only" {
		return ImportedDocument{}, fmt.Errorf("%s imports are manual-only right now; copy the text into `lw document create text` or save it locally and use `lw document create file`", source.Label)
	}

	fetched, err := fetchURLContent(rawURL, timeout)
	if err != nil {
		return ImportedDocument{}, err
	}

	return ImportedDocument{
		Title:   firstNonEmpty(fetched.Title, deriveTitleFromURL(rawURL)),
		Content: fetched.Body,
		Source:  sourceMetadata(source, rawURL, fetched.ContentType),
	}, nil
}

func ResolveSourceByID(sourceID string) (SourceSpec, error) {
	switch strings.ToLower(strings.TrimSpace(sourceID)) {
	case "plain_text":
		return SourcePlainText, nil
	case "local_document":
		return SourceFile, nil
	case "local_pdf":
		return SourcePDF, nil
	case "web", "web_article", "url":
		return SourceWeb, nil
	case "x", "twitter", "x_twitter":
		return SourceX, nil
	case "wechat", "wechat_article":
		return SourceWechat, nil
	case "youtube", "youtube_video":
		return SourceYouTube, nil
	case "zhihu", "zhihu_article":
		return SourceZhihu, nil
	case "xiaohongshu", "xhs", "xiaohongshu_post":
		return SourceXHS, nil
	default:
		return SourceSpec{}, fmt.Errorf("unknown source %q", sourceID)
	}
}

func DetectURLSource(rawURL string) SourceSpec {
	parsed, err := neturl.Parse(rawURL)
	if err != nil {
		return SourceWeb
	}
	host := strings.ToLower(parsed.Hostname())
	switch {
	case host == "x.com" || host == "twitter.com" || strings.HasSuffix(host, ".x.com") || strings.HasSuffix(host, ".twitter.com"):
		return SourceX
	case host == "mp.weixin.qq.com" || strings.HasSuffix(host, ".mp.weixin.qq.com"):
		return SourceWechat
	case host == "youtube.com" || host == "youtu.be" || strings.HasSuffix(host, ".youtube.com"):
		return SourceYouTube
	case host == "zhihu.com" || strings.HasSuffix(host, ".zhihu.com"):
		return SourceZhihu
	case host == "xiaohongshu.com" || host == "xhslink.com" || strings.HasSuffix(host, ".xiaohongshu.com") || strings.HasSuffix(host, ".xhslink.com"):
		return SourceXHS
	default:
		return SourceWeb
	}
}

func Slugify(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	var b strings.Builder
	lastDash := false
	for _, r := range raw {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		case r == '-' || r == '_' || unicode.IsSpace(r):
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

func HumanizeName(raw string) string {
	return humanizeName(raw)
}

type fetchedURLContent struct {
	Title       string
	Body        string
	ContentType string
}

func fetchURLContent(rawURL string, timeout time.Duration) (fetchedURLContent, error) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, nil)
	if err != nil {
		return fetchedURLContent{}, err
	}
	req.Header.Set("User-Agent", version.Name+"/"+version.Version)
	req.Header.Set("Accept", "text/plain, text/markdown, text/html, application/xhtml+xml, application/json;q=0.9, */*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return fetchedURLContent{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fetchedURLContent{}, fmt.Errorf("fetch %q failed with status %s", rawURL, resp.Status)
	}
	limited := io.LimitReader(resp.Body, MaxImportBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return fetchedURLContent{}, err
	}
	if len(body) > MaxImportBytes {
		return fetchedURLContent{}, fmt.Errorf("fetched body exceeds %d bytes", MaxImportBytes)
	}
	if !utf8.Valid(body) {
		return fetchedURLContent{}, fmt.Errorf("fetched body from %q is not UTF-8 text", rawURL)
	}

	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	text := string(body)
	return fetchedURLContent{
		Title:       firstNonEmpty(extractHTMLTitle(text), deriveTitleFromURL(rawURL)),
		Body:        text,
		ContentType: firstNonEmpty(contentType, "text/plain; charset=utf-8"),
	}, nil
}

func sourceMetadata(spec SourceSpec, originalRef string, contentType string) *api.DocumentSource {
	return &api.DocumentSource{
		ID:          spec.ID,
		Label:       spec.Label,
		Category:    spec.Category,
		InputMode:   spec.InputMode,
		OriginalRef: strings.TrimSpace(originalRef),
		CapturedAt:  time.Now().UTC().Format(time.RFC3339),
		ContentType: strings.TrimSpace(contentType),
		Adapter:     spec.Adapter,
	}
}

func detectFileSource(path string) (SourceSpec, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".pdf":
		return SourcePDF, fmt.Errorf("PDF import is not implemented yet; convert it to text first or add a PDF extractor")
	default:
		return SourceFile, nil
	}
}

func resolveURLSource(rawURL string, sourceKind string) (SourceSpec, error) {
	switch strings.ToLower(strings.TrimSpace(sourceKind)) {
	case "", "auto":
		return DetectURLSource(rawURL), nil
	default:
		return ResolveSourceByID(sourceKind)
	}
}

func extractHTMLTitle(text string) string {
	re := regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	matches := re.FindStringSubmatch(text)
	if len(matches) < 2 {
		return ""
	}
	title := html.UnescapeString(matches[1])
	title = strings.Join(strings.Fields(title), " ")
	return strings.TrimSpace(title)
}

func deriveTitleFromURL(rawURL string) string {
	parsed, err := neturl.Parse(rawURL)
	if err != nil {
		return ""
	}
	base := strings.Trim(parsed.Path, "/")
	if base != "" {
		parts := strings.Split(base, "/")
		return humanizeName(parts[len(parts)-1])
	}
	return humanizeName(parsed.Hostname())
}

func humanizeName(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimSuffix(raw, filepath.Ext(raw))
	replacer := strings.NewReplacer("-", " ", "_", " ", "+", " ")
	raw = replacer.Replace(raw)
	raw = strings.Join(strings.Fields(raw), " ")
	if raw == "" {
		return ""
	}
	words := strings.Fields(raw)
	for i, word := range words {
		runes := []rune(strings.ToLower(word))
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
