package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ifuryst/llm-wiki/internal/api"
	"github.com/ifuryst/llm-wiki/internal/ingest"
	"github.com/spf13/cobra"
)

type documentCreateInput struct {
	FolderID      int64
	Slug          string
	Title         string
	Content       string
	Source        *api.DocumentSource
	AuthorType    string
	AuthorID      string
	ChangeSummary string
}

func newDocumentCreateCommand(baseURL *string, timeout *time.Duration, accessToken *string, tokenFile *string, profileName *string) *cobra.Command {
	inlineInput := documentCreateInput{}
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a document from inline text, a file, or a URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocumentCreateInline(cmd, baseURL, timeout, accessToken, tokenFile, profileName, inlineInput)
		},
	}
	addCommonCreateFlags(createCmd, &inlineInput)
	_ = createCmd.MarkFlagRequired("folder-id")
	_ = createCmd.MarkFlagRequired("title")

	textInput := documentCreateInput{}
	textCmd := &cobra.Command{
		Use:     "text",
		Aliases: []string{"inline", "raw"},
		Short:   "Create a document from inline text",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocumentCreateInline(cmd, baseURL, timeout, accessToken, tokenFile, profileName, textInput)
		},
	}
	addCommonCreateFlags(textCmd, &textInput)
	_ = textCmd.MarkFlagRequired("folder-id")
	_ = textCmd.MarkFlagRequired("title")

	fileInput := documentCreateInput{}
	var filePath string
	fileCmd := &cobra.Command{
		Use:   "file",
		Short: "Create a document from a local text file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocumentCreateFile(cmd, baseURL, timeout, accessToken, tokenFile, profileName, fileInput, filePath)
		},
	}
	addCommonCreateFlags(fileCmd, &fileInput)
	fileCmd.Flags().StringVar(&filePath, "path", "", "Path to a local file to import")
	_ = fileCmd.MarkFlagRequired("folder-id")
	_ = fileCmd.MarkFlagRequired("path")

	urlInput := documentCreateInput{}
	var rawURL string
	var sourceKind string
	urlCmd := &cobra.Command{
		Use:   "url",
		Short: "Create a document by fetching a URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocumentCreateURL(cmd, baseURL, timeout, accessToken, tokenFile, profileName, urlInput, rawURL, sourceKind)
		},
	}
	addCommonCreateFlags(urlCmd, &urlInput)
	urlCmd.Flags().StringVar(&rawURL, "url", "", "Source URL to fetch")
	urlCmd.Flags().StringVar(&sourceKind, "source", "auto", "Source kind: auto, web, x, twitter, zhihu, wechat, youtube, xiaohongshu")
	_ = urlCmd.MarkFlagRequired("folder-id")
	_ = urlCmd.MarkFlagRequired("url")

	createCmd.AddCommand(textCmd, fileCmd, urlCmd)
	createCmd.AddCommand(newFixedSourceURLCreateCommand("x", "Create a document from an X URL", "x_twitter", baseURL, timeout, accessToken, tokenFile, profileName))
	createCmd.AddCommand(newFixedSourceURLCreateCommand("twitter", "Create a document from a Twitter URL", "x_twitter", baseURL, timeout, accessToken, tokenFile, profileName))
	createCmd.AddCommand(newFixedSourceURLCreateCommand("zhihu", "Create a document from a Zhihu URL", "zhihu_article", baseURL, timeout, accessToken, tokenFile, profileName))
	createCmd.AddCommand(newFixedSourceURLCreateCommand("wechat", "Create a document from a WeChat article URL", "wechat_article", baseURL, timeout, accessToken, tokenFile, profileName))
	createCmd.AddCommand(newFixedSourceURLCreateCommand("youtube", "Create a document from a YouTube URL", "youtube_video", baseURL, timeout, accessToken, tokenFile, profileName))
	createCmd.AddCommand(newFixedSourceURLCreateCommand("xiaohongshu", "Reserve a Xiaohongshu-specific import path", "xiaohongshu_post", baseURL, timeout, accessToken, tokenFile, profileName))

	return createCmd
}

func newFixedSourceURLCreateCommand(name string, short string, sourceID string, baseURL *string, timeout *time.Duration, accessToken *string, tokenFile *string, profileName *string) *cobra.Command {
	input := documentCreateInput{}
	var rawURL string
	cmd := &cobra.Command{
		Use:   name,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocumentCreateURL(cmd, baseURL, timeout, accessToken, tokenFile, profileName, input, rawURL, sourceID)
		},
	}
	addCommonCreateFlags(cmd, &input)
	cmd.Flags().StringVar(&rawURL, "url", "", "Source URL to fetch")
	_ = cmd.MarkFlagRequired("folder-id")
	_ = cmd.MarkFlagRequired("url")
	return cmd
}

func addCommonCreateFlags(cmd *cobra.Command, input *documentCreateInput) {
	cmd.Flags().Int64Var(&input.FolderID, "folder-id", 0, "Folder ID")
	cmd.Flags().StringVar(&input.Slug, "slug", "", "Document slug; defaults to a slugified title or source name")
	cmd.Flags().StringVar(&input.Title, "title", "", "Document title")
	cmd.Flags().StringVar(&input.Content, "content", "", "Document content")
	cmd.Flags().StringVar(&input.AuthorType, "author-type", "", "Optional author type override")
	cmd.Flags().StringVar(&input.AuthorID, "author-id", "", "Optional author ID override")
	cmd.Flags().StringVar(&input.ChangeSummary, "change-summary", "", "Change summary")
}

func runDocumentCreateInline(cmd *cobra.Command, baseURL *string, timeout *time.Duration, accessToken *string, tokenFile *string, profileName *string, input documentCreateInput) error {
	input.Content = strings.TrimSpace(input.Content)
	if input.Title == "" {
		return fmt.Errorf("title is required")
	}
	if input.Slug == "" {
		input.Slug = ingest.Slugify(input.Title)
	}
	if input.Slug == "" {
		return fmt.Errorf("slug is required or must be derivable from title")
	}
	if input.Source == nil {
		input.Source = ingest.SourceFromInlineText()
	}
	return submitDocumentCreate(cmd, baseURL, timeout, accessToken, tokenFile, profileName, input)
}

func runDocumentCreateFile(cmd *cobra.Command, baseURL *string, timeout *time.Duration, accessToken *string, tokenFile *string, profileName *string, input documentCreateInput, path string) error {
	imported, err := ingest.ImportFile(path)
	if err != nil {
		return err
	}
	if input.Title == "" {
		input.Title = imported.Title
	}
	if input.Slug == "" {
		input.Slug = ingest.Slugify(firstNonEmptyValue(imported.Title, input.Title))
	}
	if input.Slug == "" || input.Title == "" {
		return fmt.Errorf("failed to derive title or slug from %q; pass --title or --slug explicitly", path)
	}
	if input.ChangeSummary == "" {
		input.ChangeSummary = "import file"
	}
	input.Content = imported.Content
	input.Source = imported.Source
	return submitDocumentCreate(cmd, baseURL, timeout, accessToken, tokenFile, profileName, input)
}

func runDocumentCreateURL(cmd *cobra.Command, baseURL *string, timeout *time.Duration, accessToken *string, tokenFile *string, profileName *string, input documentCreateInput, rawURL string, sourceKind string) error {
	imported, err := ingest.ImportURL(rawURL, *timeout, sourceKind)
	if err != nil {
		return err
	}

	if input.Title == "" {
		input.Title = imported.Title
	}
	if input.Slug == "" {
		input.Slug = ingest.Slugify(firstNonEmptyValue(input.Title, imported.Title))
	}
	if input.Title == "" || input.Slug == "" {
		return fmt.Errorf("failed to derive title or slug from %q; pass --title or --slug explicitly", rawURL)
	}
	if input.ChangeSummary == "" {
		label := "URL"
		if imported.Source != nil && strings.TrimSpace(imported.Source.Label) != "" {
			label = imported.Source.Label
		}
		input.ChangeSummary = "import " + label + " URL"
	}
	input.Content = imported.Content
	input.Source = imported.Source
	return submitDocumentCreate(cmd, baseURL, timeout, accessToken, tokenFile, profileName, input)
}

func submitDocumentCreate(cmd *cobra.Command, baseURL *string, timeout *time.Duration, accessToken *string, tokenFile *string, profileName *string, input documentCreateInput) error {
	client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, "", *profileName)
	if err != nil {
		return err
	}
	resp, err := client.CreateDocument(context.Background(), api.CreateDocumentRequest{
		FolderID:      input.FolderID,
		Slug:          input.Slug,
		Title:         input.Title,
		Content:       input.Content,
		Source:        input.Source,
		AuthorType:    input.AuthorType,
		AuthorID:      input.AuthorID,
		ChangeSummary: input.ChangeSummary,
	})
	if err != nil {
		return err
	}
	return printJSON(cmd, resp)
}

func firstNonEmptyValue(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
