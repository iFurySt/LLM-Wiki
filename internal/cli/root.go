package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ifuryst/llm-wiki/internal/api"
	"github.com/ifuryst/llm-wiki/internal/httpclient"
	"github.com/ifuryst/llm-wiki/internal/version"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	var baseURL string
	var timeout time.Duration
	var tenantID string

	root := &cobra.Command{
		Use:   "llm-wiki",
		Short: "LLM-Wiki CLI",
	}

	root.PersistentFlags().StringVar(&baseURL, "base-url", defaultString("LLM_WIKI_CLI_BASE_URL", "https://llm-wiki.ifuryst.com/"), "LLM-Wiki server base URL")
	root.PersistentFlags().DurationVar(&timeout, "timeout", 10*time.Second, "HTTP timeout")
	root.PersistentFlags().StringVar(&tenantID, "tenant", "default", "LLM-Wiki tenant ID")

	root.AddCommand(newVersionCommand())
	root.AddCommand(newSystemCommand(&baseURL, &timeout, &tenantID))
	root.AddCommand(newSpaceCommand(&baseURL, &timeout, &tenantID))
	root.AddCommand(newNamespaceCommand(&baseURL, &timeout, &tenantID))
	root.AddCommand(newDocumentCommand(&baseURL, &timeout, &tenantID))

	return root
}

func defaultString(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func newSpaceCommand(baseURL *string, timeout *time.Duration, tenantID *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "space",
		Short: "Inspect spaces",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List spaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := httpclient.New(*baseURL, *timeout, *tenantID)
			resp, err := client.ListSpaces(context.Background())
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	})
	return cmd
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", version.Name, version.Version)
		},
	}
}

func newSystemCommand(baseURL *string, timeout *time.Duration, tenantID *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "Inspect LLM-Wiki system endpoints",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "info",
		Short: "Fetch system info from the LLM-Wiki server",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := httpclient.New(*baseURL, *timeout, *tenantID)
			info, err := client.GetSystemInfo(context.Background())
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(
				cmd.OutOrStdout(),
				"name=%s version=%s environment=%s server=%s:%d\n",
				info.Name,
				info.Version,
				info.Environment,
				info.Server.Host,
				info.Server.Port,
			)
			return err
		},
	})

	return cmd
}

func newNamespaceCommand(baseURL *string, timeout *time.Duration, tenantID *string) *cobra.Command {
	var key string
	var displayName string
	var description string
	var visibility string

	cmd := &cobra.Command{
		Use:   "namespace",
		Short: "Manage namespaces",
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a namespace",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := httpclient.New(*baseURL, *timeout, *tenantID)
			resp, err := client.CreateNamespace(context.Background(), api.CreateNamespaceRequest{
				Key:         key,
				DisplayName: displayName,
				Description: description,
				Visibility:  visibility,
			})
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	}
	createCmd.Flags().StringVar(&key, "key", "", "Namespace key")
	createCmd.Flags().StringVar(&displayName, "display-name", "", "Namespace display name")
	createCmd.Flags().StringVar(&description, "description", "", "Namespace description")
	createCmd.Flags().StringVar(&visibility, "visibility", "private", "Namespace visibility")
	_ = createCmd.MarkFlagRequired("key")
	_ = createCmd.MarkFlagRequired("display-name")

	getCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a namespace by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			namespaceID, err := parseInt64Arg(args[0])
			if err != nil {
				return err
			}
			client := httpclient.New(*baseURL, *timeout, *tenantID)
			resp, err := client.GetNamespace(context.Background(), namespaceID)
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List namespaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := httpclient.New(*baseURL, *timeout, *tenantID)
			resp, err := client.ListNamespaces(context.Background())
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	}

	archiveCmd := &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive a namespace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			namespaceID, err := parseInt64Arg(args[0])
			if err != nil {
				return err
			}
			client := httpclient.New(*baseURL, *timeout, *tenantID)
			resp, err := client.ArchiveNamespace(context.Background(), namespaceID)
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	}

	cmd.AddCommand(createCmd, getCmd, listCmd, archiveCmd)
	return cmd
}

func newDocumentCommand(baseURL *string, timeout *time.Duration, tenantID *string) *cobra.Command {
	var namespaceID int64
	var slug string
	var title string
	var content string
	var authorType string
	var authorID string
	var changeSummary string
	var namespaceIDForList int64
	var statusForList string

	cmd := &cobra.Command{
		Use:   "document",
		Short: "Manage documents",
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a document",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := httpclient.New(*baseURL, *timeout, *tenantID)
			resp, err := client.CreateDocument(context.Background(), api.CreateDocumentRequest{
				NamespaceID:   namespaceID,
				Slug:          slug,
				Title:         title,
				Content:       content,
				AuthorType:    authorType,
				AuthorID:      authorID,
				ChangeSummary: changeSummary,
			})
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	}
	createCmd.Flags().Int64Var(&namespaceID, "namespace-id", 0, "Namespace ID")
	createCmd.Flags().StringVar(&slug, "slug", "", "Document slug")
	createCmd.Flags().StringVar(&title, "title", "", "Document title")
	createCmd.Flags().StringVar(&content, "content", "", "Document content")
	createCmd.Flags().StringVar(&authorType, "author-type", "agent", "Author type")
	createCmd.Flags().StringVar(&authorID, "author-id", "", "Author ID")
	createCmd.Flags().StringVar(&changeSummary, "change-summary", "", "Change summary")
	_ = createCmd.MarkFlagRequired("namespace-id")
	_ = createCmd.MarkFlagRequired("slug")
	_ = createCmd.MarkFlagRequired("title")

	getCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a document by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			documentID, err := parseInt64Arg(args[0])
			if err != nil {
				return err
			}
			client := httpclient.New(*baseURL, *timeout, *tenantID)
			resp, err := client.GetDocument(context.Background(), documentID)
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	}

	getBySlugCmd := &cobra.Command{
		Use:   "get-by-slug <namespace-id> <slug>",
		Short: "Get a document by namespace and slug",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			namespaceID, err := parseInt64Arg(args[0])
			if err != nil {
				return err
			}
			client := httpclient.New(*baseURL, *timeout, *tenantID)
			resp, err := client.GetDocumentBySlug(context.Background(), namespaceID, args[1])
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List documents",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := httpclient.New(*baseURL, *timeout, *tenantID)
			var namespaceID *int64
			if namespaceIDForList > 0 {
				namespaceID = &namespaceIDForList
			}
			var status *string
			if statusForList != "" {
				status = &statusForList
			}
			resp, err := client.ListDocuments(context.Background(), namespaceID, status)
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	}
	listCmd.Flags().Int64Var(&namespaceIDForList, "namespace-id", 0, "Optional namespace ID filter")
	listCmd.Flags().StringVar(&statusForList, "status", "", "Optional status filter")

	updateCmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a document and create a new revision",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			documentID, err := parseInt64Arg(args[0])
			if err != nil {
				return err
			}
			client := httpclient.New(*baseURL, *timeout, *tenantID)
			resp, err := client.UpdateDocument(context.Background(), documentID, api.UpdateDocumentRequest{
				Title:         title,
				Content:       content,
				AuthorType:    authorType,
				AuthorID:      authorID,
				ChangeSummary: changeSummary,
			})
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	}
	updateCmd.Flags().StringVar(&title, "title", "", "Document title")
	updateCmd.Flags().StringVar(&content, "content", "", "Document content")
	updateCmd.Flags().StringVar(&authorType, "author-type", "agent", "Author type")
	updateCmd.Flags().StringVar(&authorID, "author-id", "", "Author ID")
	updateCmd.Flags().StringVar(&changeSummary, "change-summary", "", "Change summary")
	_ = updateCmd.MarkFlagRequired("title")

	archiveCmd := &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			documentID, err := parseInt64Arg(args[0])
			if err != nil {
				return err
			}
			client := httpclient.New(*baseURL, *timeout, *tenantID)
			resp, err := client.ArchiveDocument(context.Background(), documentID, api.ArchiveDocumentRequest{
				AuthorType:    authorType,
				AuthorID:      authorID,
				ChangeSummary: changeSummary,
			})
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	}
	archiveCmd.Flags().StringVar(&authorType, "author-type", "agent", "Author type")
	archiveCmd.Flags().StringVar(&authorID, "author-id", "", "Author ID")
	archiveCmd.Flags().StringVar(&changeSummary, "change-summary", "archive document", "Change summary")

	cmd.AddCommand(createCmd, getCmd, getBySlugCmd, listCmd, updateCmd, archiveCmd)
	return cmd
}

func parseInt64Arg(raw string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(raw, "%d", &id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func printJSON(cmd *cobra.Command, value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", payload)
	return err
}
