package cli

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ifuryst/llm-wiki/internal/api"
	"github.com/ifuryst/llm-wiki/internal/httpclient"
	"github.com/ifuryst/llm-wiki/internal/version"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	var baseURL string
	var timeout time.Duration
	var accessToken string
	var tokenFile string
	var tenantID string
	var profileName string

	root := &cobra.Command{
		Use:   "llm-wiki",
		Short: "LLM-Wiki CLI",
	}

	root.PersistentFlags().StringVar(&baseURL, "base-url", "", "LLM-Wiki server base URL")
	root.PersistentFlags().DurationVar(&timeout, "timeout", 10*time.Second, "HTTP timeout")
	root.PersistentFlags().StringVar(&accessToken, "token", "", "Bearer token")
	root.PersistentFlags().StringVar(&tokenFile, "token-file", "", "Path to a bearer token file")
	root.PersistentFlags().StringVar(&tenantID, "tenant", "", "Tenant ID used for login/bootstrap flows")
	root.PersistentFlags().StringVar(&profileName, "profile", "", "CLI profile name under ~/.llm-wiki/config.json")

	root.AddCommand(newVersionCommand())
	root.AddCommand(newAuthCommand(&baseURL, &timeout, &accessToken, &tokenFile, &tenantID, &profileName))
	root.AddCommand(newWorkspaceCommand(&baseURL, &timeout, &accessToken, &tokenFile, &tenantID, &profileName))
	root.AddCommand(newSystemCommand(&baseURL, &timeout, &accessToken, &tokenFile, &tenantID, &profileName))
	root.AddCommand(newSpaceCommand(&baseURL, &timeout, &accessToken, &tokenFile, &tenantID, &profileName))
	root.AddCommand(newNamespaceCommand(&baseURL, &timeout, &accessToken, &tokenFile, &tenantID, &profileName))
	root.AddCommand(newDocumentCommand(&baseURL, &timeout, &accessToken, &tokenFile, &tenantID, &profileName))

	return root
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

func newAuthCommand(baseURL *string, timeout *time.Duration, accessToken *string, tokenFile *string, tenantID *string, profileName *string) *cobra.Command {
	var useDeviceCode bool
	var noOpen bool
	var displayName string
	var accessTokenToStore string
	var provider string

	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate and manage LLM-Wiki credentials",
	}

	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Sign in with browser, device code, or store an existing token",
		RunE: func(cmd *cobra.Command, args []string) error {
			options, err := resolvedRuntimeOptions(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			if accessTokenToStore != "" {
				client := httpclient.New(options.BaseURL, options.Timeout, accessTokenToStore)
				whoami, err := client.WhoAmI(context.Background())
				if err != nil {
					return err
				}
				return persistProfile(options.ProfileName, storedProfile{
					BaseURL:     options.BaseURL,
					TenantID:    whoami.TenantID,
					AccessToken: accessTokenToStore,
					PrincipalID: whoami.PrincipalID,
					DisplayName: whoami.DisplayName,
				})
			}
			if useDeviceCode {
				return deviceLogin(cmd, options, displayName, noOpen)
			}
			return browserLogin(cmd, options, displayName, provider, noOpen)
		},
	}
	loginCmd.Flags().BoolVar(&useDeviceCode, "device-code", false, "Use the device authorization flow")
	loginCmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not automatically open a browser")
	loginCmd.Flags().StringVar(&displayName, "name", "", "Preferred display name for the browser/device login")
	loginCmd.Flags().StringVar(&provider, "provider", "", "OAuth provider to use for browser login")
	loginCmd.Flags().StringVar(&accessTokenToStore, "access-token", "", "Store an existing bearer token into the active profile")

	cmd.AddCommand(loginCmd)
	cmd.AddCommand(&cobra.Command{
		Use:   "switch-tenant <tenant-id>",
		Short: "Switch the active profile to another workspace or tenant you can access",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, options, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			resp, err := client.SwitchTenant(context.Background(), args[0])
			if err != nil {
				return err
			}
			return persistLoginProfile(options, resp, "")
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show the active auth status",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, options, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			whoami, err := client.WhoAmI(context.Background())
			if err != nil {
				return err
			}
			payload := map[string]any{
				"profile":        options.ProfileName,
				"base_url":       options.BaseURL,
				"tenant_id":      whoami.TenantID,
				"principal_id":   whoami.PrincipalID,
				"principal_type": whoami.PrincipalType,
				"display_name":   whoami.DisplayName,
				"scopes":         whoami.Scopes,
			}
			return printJSON(cmd, payload)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "logout",
		Short: "Remove stored tokens from the active profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			options, err := resolvedRuntimeOptions(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			return clearProfileTokens(options.ProfileName)
		},
	})

	serviceCmd := &cobra.Command{
		Use:   "service-principal",
		Short: "Manage service principals",
	}
	var serviceDisplayName string
	createServiceCmd := &cobra.Command{
		Use:   "create",
		Short: "Create or get a service principal",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			resp, err := client.CreateServicePrincipal(context.Background(), api.CreateServicePrincipalRequest{DisplayName: serviceDisplayName})
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	}
	createServiceCmd.Flags().StringVar(&serviceDisplayName, "display-name", "", "Service principal display name")
	_ = createServiceCmd.MarkFlagRequired("display-name")
	serviceCmd.AddCommand(createServiceCmd)
	serviceCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List service principals",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			resp, err := client.ListServicePrincipals(context.Background())
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	})
	cmd.AddCommand(serviceCmd)

	tokenCmd := &cobra.Command{
		Use:   "token",
		Short: "Manage tenant-scoped service tokens",
	}
	var principalID string
	var tokenDisplayName string
	var expiresIn time.Duration
	var scopes []string
	issueCmd := &cobra.Command{
		Use:   "issue",
		Short: "Issue a service token",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			resp, err := client.IssueToken(context.Background(), api.IssueTokenRequest{
				PrincipalID:      principalID,
				DisplayName:      tokenDisplayName,
				Scopes:           scopes,
				ExpiresInSeconds: int(expiresIn.Seconds()),
			})
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	}
	issueCmd.Flags().StringVar(&principalID, "principal-id", "", "Target service principal ID")
	issueCmd.Flags().StringVar(&tokenDisplayName, "display-name", "", "Token display name")
	issueCmd.Flags().DurationVar(&expiresIn, "expires-in", 24*time.Hour, "Token TTL")
	issueCmd.Flags().StringSliceVar(&scopes, "scope", []string{}, "Scope to grant, repeatable")
	_ = issueCmd.MarkFlagRequired("principal-id")
	_ = issueCmd.MarkFlagRequired("display-name")
	tokenCmd.AddCommand(issueCmd)
	tokenCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List issued tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			resp, err := client.ListTokens(context.Background())
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	})
	tokenCmd.AddCommand(&cobra.Command{
		Use:   "revoke <id>",
		Short: "Revoke a token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			tokenID, err := parseInt64Arg(args[0])
			if err != nil {
				return err
			}
			resp, err := client.RevokeToken(context.Background(), tokenID)
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	})
	cmd.AddCommand(tokenCmd)

	return cmd
}

func newWorkspaceCommand(baseURL *string, timeout *time.Duration, accessToken *string, tokenFile *string, tenantID *string, profileName *string) *cobra.Command {
	var workspaceDisplayName string
	var workspaceTenantID string
	var inviteEmail string
	var inviteRole string
	var inviteTTL time.Duration
	var exportOutputDir string

	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage workspaces and invitations",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List workspaces available to the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			resp, err := client.ListWorkspaces(context.Background())
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	})
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			resp, err := client.CreateWorkspace(context.Background(), api.CreateWorkspaceRequest{
				TenantID:    workspaceTenantID,
				DisplayName: workspaceDisplayName,
			})
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	}
	createCmd.Flags().StringVar(&workspaceDisplayName, "display-name", "", "Workspace display name")
	createCmd.Flags().StringVar(&workspaceTenantID, "tenant-id", "", "Optional tenant/workspace key")
	_ = createCmd.MarkFlagRequired("display-name")
	cmd.AddCommand(createCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "invite-list",
		Short: "List invitations for the current workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			resp, err := client.ListInvites(context.Background())
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	})
	inviteCmd := &cobra.Command{
		Use:   "invite",
		Short: "Invite a user to the current workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			resp, err := client.CreateInvite(context.Background(), api.CreateInviteRequest{
				Email:          inviteEmail,
				Role:           inviteRole,
				ExpiresInHours: int(inviteTTL.Hours()),
			})
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	}
	inviteCmd.Flags().StringVar(&inviteEmail, "email", "", "Invitee email")
	inviteCmd.Flags().StringVar(&inviteRole, "role", "member", "Invite role")
	inviteCmd.Flags().DurationVar(&inviteTTL, "expires-in", 72*time.Hour, "Invite lifetime")
	_ = inviteCmd.MarkFlagRequired("email")
	cmd.AddCommand(inviteCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "accept-invite <token>",
		Short: "Accept an invite into another workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			resp, err := client.AcceptInvite(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	})
	exportCmd := &cobra.Command{
		Use:   "export-obsidian",
		Short: "Export the current workspace into an Obsidian-compatible vault folder",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, options, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			namespacesResp, err := client.ListNamespaces(context.Background())
			if err != nil {
				return err
			}
			documentsResp, err := client.ListDocuments(context.Background(), nil, nil)
			if err != nil {
				return err
			}
			outputDir := strings.TrimSpace(exportOutputDir)
			if outputDir == "" {
				outputDir = fmt.Sprintf("llm-wiki-%s-obsidian", options.TenantID)
			}
			if err := exportWorkspaceToObsidian(outputDir, options.TenantID, namespacesResp.Items, documentsResp.Items); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "exported %d documents to %s\n", len(documentsResp.Items), outputDir)
			return nil
		},
	}
	exportCmd.Flags().StringVar(&exportOutputDir, "output", "", "Output directory for the Obsidian vault export")
	cmd.AddCommand(exportCmd)
	return cmd
}

func newSpaceCommand(baseURL *string, timeout *time.Duration, accessToken *string, tokenFile *string, tenantID *string, profileName *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "space",
		Short: "Inspect spaces",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List spaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			resp, err := client.ListSpaces(context.Background())
			if err != nil {
				return err
			}
			return printJSON(cmd, resp)
		},
	})
	return cmd
}

func newSystemCommand(baseURL *string, timeout *time.Duration, accessToken *string, tokenFile *string, tenantID *string, profileName *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "Inspect LLM-Wiki system endpoints",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "info",
		Short: "Fetch system info from the LLM-Wiki server",
		RunE: func(cmd *cobra.Command, args []string) error {
			options, err := resolvedRuntimeOptions(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			client := httpclient.New(options.BaseURL, options.Timeout, options.AccessToken)
			info, err := client.GetSystemInfo(context.Background())
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "name=%s version=%s environment=%s server=%s:%d\n", info.Name, info.Version, info.Environment, info.Server.Host, info.Server.Port)
			return err
		},
	})
	return cmd
}

func newNamespaceCommand(baseURL *string, timeout *time.Duration, accessToken *string, tokenFile *string, tenantID *string, profileName *string) *cobra.Command {
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
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
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
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			namespaceID, err := parseInt64Arg(args[0])
			if err != nil {
				return err
			}
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
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
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
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			namespaceID, err := parseInt64Arg(args[0])
			if err != nil {
				return err
			}
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

func newDocumentCommand(baseURL *string, timeout *time.Duration, accessToken *string, tokenFile *string, tenantID *string, profileName *string) *cobra.Command {
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
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
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
	createCmd.Flags().StringVar(&authorType, "author-type", "", "Optional author type override")
	createCmd.Flags().StringVar(&authorID, "author-id", "", "Optional author ID override")
	createCmd.Flags().StringVar(&changeSummary, "change-summary", "", "Change summary")
	_ = createCmd.MarkFlagRequired("namespace-id")
	_ = createCmd.MarkFlagRequired("slug")
	_ = createCmd.MarkFlagRequired("title")

	getCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a document by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			documentID, err := parseInt64Arg(args[0])
			if err != nil {
				return err
			}
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
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			namespaceID, err := parseInt64Arg(args[0])
			if err != nil {
				return err
			}
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
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			var namespaceFilter *int64
			if namespaceIDForList > 0 {
				namespaceFilter = &namespaceIDForList
			}
			var status *string
			if statusForList != "" {
				status = &statusForList
			}
			resp, err := client.ListDocuments(context.Background(), namespaceFilter, status)
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
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			documentID, err := parseInt64Arg(args[0])
			if err != nil {
				return err
			}
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
	updateCmd.Flags().StringVar(&authorType, "author-type", "", "Optional author type override")
	updateCmd.Flags().StringVar(&authorID, "author-id", "", "Optional author ID override")
	updateCmd.Flags().StringVar(&changeSummary, "change-summary", "", "Change summary")
	_ = updateCmd.MarkFlagRequired("title")

	archiveCmd := &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := authorizedClient(*baseURL, *timeout, *accessToken, *tokenFile, *tenantID, *profileName)
			if err != nil {
				return err
			}
			documentID, err := parseInt64Arg(args[0])
			if err != nil {
				return err
			}
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
	archiveCmd.Flags().StringVar(&authorType, "author-type", "", "Optional author type override")
	archiveCmd.Flags().StringVar(&authorID, "author-id", "", "Optional author ID override")
	archiveCmd.Flags().StringVar(&changeSummary, "change-summary", "archive document", "Change summary")

	cmd.AddCommand(createCmd, getCmd, getBySlugCmd, listCmd, updateCmd, archiveCmd)
	return cmd
}

type runtimeOptions struct {
	resolvedOptions
	Timeout time.Duration
}

func resolvedRuntimeOptions(baseURL string, timeout time.Duration, accessToken string, tokenFile string, tenantID string, profileName string) (runtimeOptions, error) {
	if tokenFile == "" {
		tokenFile = os.Getenv("LLM_WIKI_TOKEN_FILE")
	}
	if tokenFile != "" {
		payload, err := os.ReadFile(tokenFile)
		if err != nil {
			return runtimeOptions{}, err
		}
		accessToken = strings.TrimSpace(string(payload))
	}
	resolved, err := resolveOptions(baseURL, tenantID, accessToken, "", profileName)
	if err != nil {
		return runtimeOptions{}, err
	}
	return runtimeOptions{resolvedOptions: resolved, Timeout: timeout}, nil
}

func authorizedClient(baseURL string, timeout time.Duration, accessToken string, tokenFile string, tenantID string, profileName string) (*httpclient.Client, runtimeOptions, error) {
	options, err := resolvedRuntimeOptions(baseURL, timeout, accessToken, tokenFile, tenantID, profileName)
	if err != nil {
		return nil, runtimeOptions{}, err
	}
	client := httpclient.New(options.BaseURL, options.Timeout, options.AccessToken)
	if (options.AccessToken == "" || (!options.ExpiresAt.IsZero() && time.Now().After(options.ExpiresAt))) && options.RefreshToken != "" {
		resp, err := client.ExchangeToken(context.Background(), api.TokenExchangeRequest{
			GrantType:    "refresh_token",
			RefreshToken: options.RefreshToken,
		})
		if err == nil {
			client.SetAccessToken(resp.AccessToken)
			options.AccessToken = resp.AccessToken
			options.RefreshToken = resp.RefreshToken
			expiresAt := time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
			options.ExpiresAt = expiresAt
			_ = persistProfile(options.ProfileName, storedProfile{
				BaseURL:      options.BaseURL,
				TenantID:     resp.TenantID,
				AccessToken:  resp.AccessToken,
				RefreshToken: resp.RefreshToken,
				ExpiresAt:    expiresAt.Format(time.RFC3339),
				PrincipalID:  resp.PrincipalID,
			})
			return client, options, nil
		}
	}
	return client, options, nil
}

func browserLogin(cmd *cobra.Command, options runtimeOptions, displayName string, provider string, noOpen bool) error {
	verifier, challenge, state, err := pkceValues()
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer listener.Close()

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("state") != state {
				http.Error(w, "state mismatch", http.StatusBadRequest)
				errCh <- fmt.Errorf("state mismatch")
				return
			}
			code := r.URL.Query().Get("code")
			if code == "" {
				http.Error(w, "missing code", http.StatusBadRequest)
				errCh <- fmt.Errorf("missing authorization code")
				return
			}
			_, _ = w.Write([]byte("Authentication complete. Return to the CLI."))
			codeCh <- code
		}),
	}
	go func() {
		_ = server.Serve(listener)
	}()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	redirectURI := "http://" + listener.Addr().String() + "/auth/callback"
	client := httpclient.New(options.BaseURL, options.Timeout, "")
	startResp, err := client.StartBrowserLogin(context.Background(), api.StartBrowserLoginRequest{
		TenantID:            options.TenantID,
		Provider:            provider,
		DisplayName:         displayName,
		State:               state,
		RedirectURI:         redirectURI,
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
	})
	if err != nil {
		return err
	}
	if !noOpen {
		_ = openBrowser(startResp.AuthorizeURL)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Finish signing in via your browser:\n\n%s\n\n", startResp.AuthorizeURL)

	select {
	case code := <-codeCh:
		tokenResp, err := client.ExchangeToken(context.Background(), api.TokenExchangeRequest{
			GrantType:    "authorization_code",
			Code:         code,
			CodeVerifier: verifier,
		})
		if err != nil {
			return err
		}
		return persistLoginProfile(options, tokenResp, displayName)
	case err := <-errCh:
		return err
	case <-time.After(2 * time.Minute):
		return fmt.Errorf("browser login timed out")
	}
}

func deviceLogin(cmd *cobra.Command, options runtimeOptions, displayName string, noOpen bool) error {
	client := httpclient.New(options.BaseURL, options.Timeout, "")
	startResp, err := client.StartDeviceLogin(context.Background(), api.StartDeviceLoginRequest{
		TenantID:    options.TenantID,
		DisplayName: displayName,
	})
	if err != nil {
		return err
	}
	deviceURL := startResp.VerificationURI + "?code=" + startResp.UserCode
	if !noOpen {
		_ = openBrowser(deviceURL)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Finish signing in via your browser:\n\n%s\n\nEnter code: %s\n", deviceURL, startResp.UserCode)
	ticker := time.NewTicker(time.Duration(startResp.IntervalSeconds) * time.Second)
	defer ticker.Stop()
	timeoutAt, _ := time.Parse(time.RFC3339, startResp.ExpiresAt)
	for {
		select {
		case <-ticker.C:
			tokenResp, err := client.ExchangeToken(context.Background(), api.TokenExchangeRequest{
				GrantType:  "urn:ietf:params:oauth:grant-type:device_code",
				DeviceCode: startResp.DeviceCode,
			})
			if err != nil {
				if strings.Contains(err.Error(), "authorization_pending") {
					continue
				}
				return err
			}
			return persistLoginProfile(options, tokenResp, displayName)
		default:
			if !timeoutAt.IsZero() && time.Now().After(timeoutAt) {
				return fmt.Errorf("device authorization timed out")
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func persistLoginProfile(options runtimeOptions, tokenResp api.TokenExchangeResponse, displayName string) error {
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	if displayName == "" {
		displayName = tokenResp.PrincipalID
	}
	return persistProfile(options.ProfileName, storedProfile{
		BaseURL:      options.BaseURL,
		TenantID:     tokenResp.TenantID,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    expiresAt.Format(time.RFC3339),
		PrincipalID:  tokenResp.PrincipalID,
		DisplayName:  displayName,
	})
}

func pkceValues() (string, string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", "", err
	}
	verifier := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", "", "", err
	}
	return verifier, challenge, hex.EncodeToString(stateBytes), nil
}

func openBrowser(url string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", url)
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		command = exec.Command("xdg-open", url)
	}
	return command.Start()
}

func parseInt64Arg(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}

func printJSON(cmd *cobra.Command, value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", payload)
	return err
}
