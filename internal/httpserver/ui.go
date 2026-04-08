package httpserver

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ifuryst/llm-wiki/internal/api"
	"github.com/ifuryst/llm-wiki/internal/config"
	"github.com/ifuryst/llm-wiki/internal/service"
)

type uiSpaceCard struct {
	ID          int64
	NS          string
	Key         string
	DisplayName string
	Role        string
}

type uiNamespaceCard struct {
	ID            int64
	Key           string
	DisplayName   string
	Description   string
	Visibility    string
	Status        string
	DocumentCount int
}

type uiDocumentCard struct {
	ID                int64
	FolderID          int64
	NamespaceLabel    string
	Slug              string
	Title             string
	ContentPreview    string
	Status            string
	CurrentRevisionNo int32
	UpdatedAt         string
}

type uiNamespaceTree struct {
	ID            int64
	Key           string
	DisplayName   string
	Status        string
	DocumentCount int
	Documents     []uiDocumentCard
}

type uiFlash struct {
	Kind    string
	Message string
}

type uiActivityItem struct {
	ID             int64
	Title          string
	FolderID       int64
	NamespaceLabel string
	Status         string
	UpdatedAt      string
	ContentPreview string
	Href           string
}

type uiIndexData struct {
	CurrentPage         string
	NS                  string
	InstallBaseURL      string
	InstallDocURL       string
	InstallScriptURL    string
	InstallSkillURL     string
	MCPURL              string
	Spaces              []uiSpaceCard
	Folders             []uiNamespaceCard
	NamespaceTree       []uiNamespaceTree
	Documents           []uiDocumentCard
	SelectedDocument    *api.DocumentResponse
	SelectedRevision    *api.RevisionResponse
	SelectedNamespaceID int64
	StatusFilter        string
	DocumentCount       int
	ArchivedCount       int
	NamespaceCount      int
	CurrentSpaceName    string
	ActivityItems       []uiActivityItem
	NSSwitchStateJSON   template.JS
	TreeStateJSON       template.JS
	Flash               *uiFlash
}

func registerUIRoutes(engine *gin.Engine, svc *service.Service, cfg config.Config) {
	ui := engine.Group("")
	ui.Use(attachOptionalWebSession(svc))

	ui.GET("/", func(c *gin.Context) {
		status, err := svc.SetupStatus(c.Request.Context(), cfg.Auth.BootstrapTenantID)
		if err == nil && !status.Initialized {
			c.Redirect(http.StatusTemporaryRedirect, "/setup")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, "/ui")
	})

	ui.POST("/ui/ns/switch", func(c *gin.Context) {
		targetNS := strings.TrimSpace(c.PostForm("ns"))
		if targetNS == "" {
			c.Redirect(http.StatusSeeOther, "/ui?kind=error&message=missing+ns")
			return
		}
		session, err := svc.SwitchWebSessionNS(c.Request.Context(), targetNS)
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/ui?kind=error&message=unable+to+switch+ns")
			return
		}
		setWebSessionCookie(c, session.ID)
		c.Redirect(http.StatusSeeOther, "/ui?kind=success&message=switched+ns")
	})

	ui.GET("/ui", func(c *gin.Context) {
		status, err := svc.SetupStatus(c.Request.Context(), cfg.Auth.BootstrapTenantID)
		if err == nil && !status.Initialized {
			c.Redirect(http.StatusTemporaryRedirect, "/setup")
			return
		}
		renderUIPage(c, svc, cfg, "wiki")
	})

	ui.GET("/ui/install", func(c *gin.Context) {
		status, err := svc.SetupStatus(c.Request.Context(), cfg.Auth.BootstrapTenantID)
		if err == nil && !status.Initialized {
			c.Redirect(http.StatusTemporaryRedirect, "/setup")
			return
		}
		renderUIPage(c, svc, cfg, "install")
	})

	ui.GET("/ui/fragment/wiki", func(c *gin.Context) {
		renderUIFragment(c, svc, cfg, "wiki", "ui_wiki_content.html")
	})

	ui.GET("/ui/fragment/install", func(c *gin.Context) {
		renderUIFragment(c, svc, cfg, "install", "ui_install_content.html")
	})

	ui.GET("/ui/fragment/reader", func(c *gin.Context) {
		renderUIFragment(c, svc, cfg, "wiki", "ui_reader_content.html")
	})

	ui.POST("/ui/folders", func(c *gin.Context) {
		req := api.CreateFolderRequest{
			Key:         c.PostForm("key"),
			DisplayName: c.PostForm("display_name"),
			Description: c.PostForm("description"),
			Visibility:  c.PostForm("visibility"),
		}
		if _, err := svc.CreateFolder(c.Request.Context(), tenantIDFromRequest(c), req); err != nil {
			handleError(c, err)
			return
		}
		c.Redirect(http.StatusSeeOther, "/ui?kind=success&message=folder+created")
	})

	ui.POST("/ui/folders/:id/archive", func(c *gin.Context) {
		namespaceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		if _, err := svc.ArchiveFolder(c.Request.Context(), tenantIDFromRequest(c), namespaceID); err != nil {
			handleError(c, err)
			return
		}
		c.Redirect(http.StatusSeeOther, "/ui?kind=success&message=folder+archived")
	})

	ui.POST("/ui/documents", func(c *gin.Context) {
		namespaceID, err := strconv.ParseInt(c.PostForm("folder_id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		req := api.CreateDocumentRequest{
			FolderID:      namespaceID,
			Slug:          c.PostForm("slug"),
			Title:         c.PostForm("title"),
			Content:       c.PostForm("content"),
			AuthorType:    c.PostForm("author_type"),
			AuthorID:      c.PostForm("author_id"),
			ChangeSummary: c.PostForm("change_summary"),
		}
		if _, err := svc.CreateDocument(c.Request.Context(), tenantIDFromRequest(c), req); err != nil {
			handleError(c, err)
			return
		}
		c.Redirect(http.StatusSeeOther, "/ui?kind=success&message=document+created")
	})

	ui.POST("/ui/documents/:id/update", func(c *gin.Context) {
		documentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		req := api.UpdateDocumentRequest{
			Title:         c.PostForm("title"),
			Content:       c.PostForm("content"),
			AuthorType:    c.PostForm("author_type"),
			AuthorID:      c.PostForm("author_id"),
			ChangeSummary: c.PostForm("change_summary"),
		}
		if _, err := svc.UpdateDocument(c.Request.Context(), tenantIDFromRequest(c), documentID, req); err != nil {
			handleError(c, err)
			return
		}
		c.Redirect(http.StatusSeeOther, "/ui?kind=success&message=document+updated&document_id="+strconv.FormatInt(documentID, 10))
	})

	ui.POST("/ui/documents/:id/archive", func(c *gin.Context) {
		documentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		req := api.ArchiveDocumentRequest{
			AuthorType:    c.PostForm("author_type"),
			AuthorID:      c.PostForm("author_id"),
			ChangeSummary: c.PostForm("change_summary"),
		}
		if _, err := svc.ArchiveDocument(c.Request.Context(), tenantIDFromRequest(c), documentID, req); err != nil {
			handleError(c, err)
			return
		}
		c.Redirect(http.StatusSeeOther, "/ui?kind=success&message=document+archived&document_id="+strconv.FormatInt(documentID, 10))
	})
}

func renderUIPage(c *gin.Context, svc *service.Service, cfg config.Config, currentPage string) {
	data, err := buildUIIndexData(c, svc, cfg)
	if err != nil {
		handleError(c, err)
		return
	}
	data.CurrentPage = currentPage
	c.HTML(http.StatusOK, "index.html", data)
}

func renderUIFragment(c *gin.Context, svc *service.Service, cfg config.Config, currentPage string, templateName string) {
	data, err := buildUIIndexData(c, svc, cfg)
	if err != nil {
		handleError(c, err)
		return
	}
	data.CurrentPage = currentPage
	c.Header("Cache-Control", "no-store")
	c.HTML(http.StatusOK, templateName, data)
}

func buildUIIndexData(c *gin.Context, svc *service.Service, cfg config.Config) (uiIndexData, error) {
	tenantID := tenantIDFromRequest(c)
	baseURL := strings.TrimRight(cfg.Install.BaseURL, "/")
	var namespaceFilter *int64
	var selectedNamespaceID int64
	if raw := c.Query("folder_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return uiIndexData{}, err
		}
		namespaceFilter = &parsed
		selectedNamespaceID = parsed
	}

	var statusFilter *string
	statusRaw := c.Query("status")
	if statusRaw != "" && statusRaw != "all" {
		statusFilter = &statusRaw
	}

	ns, err := svc.ListNS(c.Request.Context(), tenantID)
	if err != nil {
		return uiIndexData{}, err
	}
	folders, err := svc.ListFolders(c.Request.Context(), tenantID)
	if err != nil {
		return uiIndexData{}, err
	}
	documents, err := svc.ListDocuments(c.Request.Context(), tenantID, namespaceFilter, statusFilter)
	if err != nil {
		return uiIndexData{}, err
	}

	namespaceNames := make(map[int64]string, len(folders.Items))
	namespaceDocCount := make(map[int64]int, len(folders.Items))
	for _, item := range folders.Items {
		namespaceNames[item.ID] = item.DisplayName
	}
	for _, item := range documents.Items {
		namespaceDocCount[item.FolderID]++
	}

	var selectedDocument *api.DocumentResponse
	if raw := c.Query("document_id"); raw != "" {
		documentID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return uiIndexData{}, err
		}
		doc, err := svc.GetDocument(c.Request.Context(), tenantID, documentID)
		if err != nil {
			return uiIndexData{}, err
		}
		selectedDocument = &doc
	}

	var selectedRevision *api.RevisionResponse
	if selectedDocument != nil {
		if selectedNamespaceID == 0 {
			selectedNamespaceID = selectedDocument.FolderID
		}
		if raw := c.Query("revision_id"); raw != "" {
			revisionID, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return uiIndexData{}, err
			}
			for index := range selectedDocument.Revisions {
				if selectedDocument.Revisions[index].ID == revisionID {
					selectedRevision = &selectedDocument.Revisions[index]
					break
				}
			}
		}
	}

	spaceCards := make([]uiSpaceCard, 0, len(ns.Items))
	currentSpaceName := tenantID
	for _, item := range ns.Items {
		if item.NS == tenantID {
			currentSpaceName = item.DisplayName
		}
		spaceCards = append(spaceCards, uiSpaceCard{
			ID:          item.ID,
			NS:          item.NS,
			Key:         item.Key,
			DisplayName: item.DisplayName,
			Role:        item.Role,
		})
	}

	namespaceCards := make([]uiNamespaceCard, 0, len(folders.Items))
	namespaceTree := make([]uiNamespaceTree, 0, len(folders.Items))
	for _, item := range folders.Items {
		namespaceCards = append(namespaceCards, uiNamespaceCard{
			ID:            item.ID,
			Key:           item.Key,
			DisplayName:   item.DisplayName,
			Description:   item.Description,
			Visibility:    item.Visibility,
			Status:        item.Status,
			DocumentCount: namespaceDocCount[item.ID],
		})
		namespaceTree = append(namespaceTree, uiNamespaceTree{
			ID:            item.ID,
			Key:           item.Key,
			DisplayName:   item.DisplayName,
			Status:        item.Status,
			DocumentCount: namespaceDocCount[item.ID],
		})
	}

	documentCards := make([]uiDocumentCard, 0, len(documents.Items))
	activityItems := make([]uiActivityItem, 0, minInt(len(documents.Items), 8))
	archivedCount := 0
	for _, item := range documents.Items {
		if item.Status == "archived" {
			archivedCount++
		}
		card := uiDocumentCard{
			ID:                item.ID,
			FolderID:          item.FolderID,
			NamespaceLabel:    namespaceNames[item.FolderID],
			Slug:              item.Slug,
			Title:             item.Title,
			ContentPreview:    item.Content,
			Status:            item.Status,
			CurrentRevisionNo: item.CurrentRevisionNo,
			UpdatedAt:         item.UpdatedAt,
		}
		documentCards = append(documentCards, card)
		if len(activityItems) < 8 {
			href := "/ui?document_id=" + strconv.FormatInt(item.ID, 10) + "&folder_id=" + strconv.FormatInt(item.FolderID, 10)
			if statusRaw != "" && statusRaw != "all" {
				href += "&status=" + statusRaw
			}
			activityItems = append(activityItems, uiActivityItem{
				ID:             item.ID,
				Title:          item.Title,
				FolderID:       item.FolderID,
				NamespaceLabel: namespaceNames[item.FolderID],
				Status:         item.Status,
				UpdatedAt:      item.UpdatedAt,
				ContentPreview: item.Content,
				Href:           href,
			})
		}
		for index := range namespaceTree {
			if namespaceTree[index].ID == item.FolderID {
				namespaceTree[index].Documents = append(namespaceTree[index].Documents, card)
				break
			}
		}
	}

	treeStateJSON, err := buildUITreeStateJSON(tenantID, c.DefaultQuery("status", "all"), selectedNamespaceID, selectedDocument, namespaceTree)
	if err != nil {
		return uiIndexData{}, err
	}
	nsSwitchStateJSON, err := buildUINSSwitchStateJSON(tenantID, spaceCards)
	if err != nil {
		return uiIndexData{}, err
	}

	var flash *uiFlash
	if message := c.Query("message"); message != "" {
		flash = &uiFlash{
			Kind:    c.DefaultQuery("kind", "info"),
			Message: message,
		}
	}

	return uiIndexData{
		NS:                  tenantID,
		InstallBaseURL:      baseURL,
		InstallDocURL:       baseURL + "/install/LLM-Wiki.md",
		InstallScriptURL:    baseURL + "/install/install-cli.sh",
		InstallSkillURL:     baseURL + "/install/skills/LLM-Wiki.skill",
		MCPURL:              baseURL + "/mcp",
		Spaces:              spaceCards,
		Folders:             namespaceCards,
		NamespaceTree:       namespaceTree,
		Documents:           documentCards,
		SelectedDocument:    selectedDocument,
		SelectedRevision:    selectedRevision,
		SelectedNamespaceID: selectedNamespaceID,
		StatusFilter:        c.DefaultQuery("status", "all"),
		DocumentCount:       len(documentCards),
		ArchivedCount:       archivedCount,
		NamespaceCount:      len(namespaceCards),
		CurrentSpaceName:    currentSpaceName,
		ActivityItems:       activityItems,
		NSSwitchStateJSON:   nsSwitchStateJSON,
		TreeStateJSON:       treeStateJSON,
		Flash:               flash,
	}, nil
}

type uiTreeState struct {
	NS              string       `json:"tenantId"`
	StatusFilter    string       `json:"statusFilter"`
	SelectedItemID  string       `json:"selectedItemId,omitempty"`
	ExpandedItemIDs []string     `json:"expandedItemIds,omitempty"`
	Folders         []uiTreeNode `json:"folders"`
}

type uiTreeNode struct {
	ID       string       `json:"id"`
	Label    string       `json:"label"`
	Href     string       `json:"href,omitempty"`
	Children []uiTreeNode `json:"children,omitempty"`
}

func buildUITreeStateJSON(tenantID string, statusFilter string, selectedNamespaceID int64, selectedDocument *api.DocumentResponse, namespaceTree []uiNamespaceTree) (template.JS, error) {
	state := uiTreeState{
		NS:           tenantID,
		StatusFilter: statusFilter,
		Folders:      make([]uiTreeNode, 0, len(namespaceTree)),
	}
	if selectedDocument != nil {
		state.SelectedItemID = "doc-" + strconv.FormatInt(selectedDocument.ID, 10)
		state.ExpandedItemIDs = append(state.ExpandedItemIDs, "ns-"+strconv.FormatInt(selectedDocument.FolderID, 10))
	} else if selectedNamespaceID != 0 {
		state.SelectedItemID = "ns-" + strconv.FormatInt(selectedNamespaceID, 10)
		state.ExpandedItemIDs = append(state.ExpandedItemIDs, "ns-"+strconv.FormatInt(selectedNamespaceID, 10))
	}

	for _, folder := range namespaceTree {
		nsNode := uiTreeNode{
			ID:       "ns-" + strconv.FormatInt(folder.ID, 10),
			Label:    folder.DisplayName,
			Children: make([]uiTreeNode, 0, len(folder.Documents)),
		}
		for _, document := range folder.Documents {
			href := "/ui?document_id=" + strconv.FormatInt(document.ID, 10) + "&folder_id=" + strconv.FormatInt(document.FolderID, 10)
			if statusFilter != "" && statusFilter != "all" {
				href += "&status=" + statusFilter
			}
			nsNode.Children = append(nsNode.Children, uiTreeNode{
				ID:    "doc-" + strconv.FormatInt(document.ID, 10),
				Label: document.Title,
				Href:  href,
			})
		}
		state.Folders = append(state.Folders, nsNode)
	}

	payload, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return template.JS(payload), nil
}

type uiNSSwitchState struct {
	CurrentNS string          `json:"currentNS"`
	Spaces    []uiNSSwitchRow `json:"spaces"`
}

type uiNSSwitchRow struct {
	NS          string `json:"ns"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role,omitempty"`
}

func buildUINSSwitchStateJSON(currentNS string, spaces []uiSpaceCard) (template.JS, error) {
	state := uiNSSwitchState{
		CurrentNS: currentNS,
		Spaces:    make([]uiNSSwitchRow, 0, len(spaces)),
	}
	for _, space := range spaces {
		state.Spaces = append(state.Spaces, uiNSSwitchRow{
			NS:          space.NS,
			DisplayName: space.DisplayName,
			Role:        space.Role,
		})
	}
	payload, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return template.JS(payload), nil
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
