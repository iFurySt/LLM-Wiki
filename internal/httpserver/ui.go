package httpserver

import (
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
	Key         string
	DisplayName string
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
	NamespaceID       int64
	NamespaceLabel    string
	Slug              string
	Title             string
	ContentPreview    string
	Status            string
	CurrentRevisionNo int32
	UpdatedAt         string
}

type uiNamespaceTree struct {
	ID          int64
	Key         string
	DisplayName string
	Status      string
	Documents   []uiDocumentCard
}

type uiFlash struct {
	Kind    string
	Message string
}

type uiIndexData struct {
	CurrentPage         string
	TenantID            string
	InstallBaseURL      string
	InstallDocURL       string
	InstallScriptURL    string
	InstallSkillURL     string
	MCPURL              string
	Spaces              []uiSpaceCard
	Namespaces          []uiNamespaceCard
	NamespaceTree       []uiNamespaceTree
	Documents           []uiDocumentCard
	SelectedDocument    *api.DocumentResponse
	SelectedNamespaceID int64
	StatusFilter        string
	DocumentCount       int
	ArchivedCount       int
	NamespaceCount      int
	Flash               *uiFlash
}

func registerUIRoutes(engine *gin.Engine, svc *service.Service, cfg config.Config) {
	engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/ui")
	})

	engine.GET("/ui", func(c *gin.Context) {
		renderUIPage(c, svc, cfg, "wiki")
	})

	engine.GET("/ui/install", func(c *gin.Context) {
		renderUIPage(c, svc, cfg, "install")
	})

	engine.POST("/ui/namespaces", func(c *gin.Context) {
		req := api.CreateNamespaceRequest{
			Key:         c.PostForm("key"),
			DisplayName: c.PostForm("display_name"),
			Description: c.PostForm("description"),
			Visibility:  c.PostForm("visibility"),
		}
		if _, err := svc.CreateNamespace(c.Request.Context(), tenantIDFromRequest(c), req); err != nil {
			handleError(c, err)
			return
		}
		c.Redirect(http.StatusSeeOther, "/ui?kind=success&message=namespace+created")
	})

	engine.POST("/ui/namespaces/:id/archive", func(c *gin.Context) {
		namespaceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		if _, err := svc.ArchiveNamespace(c.Request.Context(), tenantIDFromRequest(c), namespaceID); err != nil {
			handleError(c, err)
			return
		}
		c.Redirect(http.StatusSeeOther, "/ui?kind=success&message=namespace+archived")
	})

	engine.POST("/ui/documents", func(c *gin.Context) {
		namespaceID, err := strconv.ParseInt(c.PostForm("namespace_id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		req := api.CreateDocumentRequest{
			NamespaceID:   namespaceID,
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

	engine.POST("/ui/documents/:id/update", func(c *gin.Context) {
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

	engine.POST("/ui/documents/:id/archive", func(c *gin.Context) {
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

func buildUIIndexData(c *gin.Context, svc *service.Service, cfg config.Config) (uiIndexData, error) {
	tenantID := tenantIDFromRequest(c)
	baseURL := strings.TrimRight(cfg.Install.BaseURL, "/")
	var namespaceFilter *int64
	var selectedNamespaceID int64
	if raw := c.Query("namespace_id"); raw != "" {
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

	spaces, err := svc.ListSpaces(c.Request.Context(), tenantID)
	if err != nil {
		return uiIndexData{}, err
	}
	namespaces, err := svc.ListNamespaces(c.Request.Context(), tenantID)
	if err != nil {
		return uiIndexData{}, err
	}
	documents, err := svc.ListDocuments(c.Request.Context(), tenantID, namespaceFilter, statusFilter)
	if err != nil {
		return uiIndexData{}, err
	}

	namespaceNames := make(map[int64]string, len(namespaces.Items))
	namespaceDocCount := make(map[int64]int, len(namespaces.Items))
	for _, item := range namespaces.Items {
		namespaceNames[item.ID] = item.DisplayName
	}
	for _, item := range documents.Items {
		namespaceDocCount[item.NamespaceID]++
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
	} else if len(documents.Items) > 0 {
		doc, err := svc.GetDocument(c.Request.Context(), tenantID, documents.Items[0].ID)
		if err != nil {
			return uiIndexData{}, err
		}
		selectedDocument = &doc
		if selectedNamespaceID == 0 {
			selectedNamespaceID = doc.NamespaceID
		}
	}

	spaceCards := make([]uiSpaceCard, 0, len(spaces.Items))
	for _, item := range spaces.Items {
		spaceCards = append(spaceCards, uiSpaceCard{
			ID:          item.ID,
			Key:         item.Key,
			DisplayName: item.DisplayName,
		})
	}

	namespaceCards := make([]uiNamespaceCard, 0, len(namespaces.Items))
	namespaceTree := make([]uiNamespaceTree, 0, len(namespaces.Items))
	for _, item := range namespaces.Items {
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
			ID:          item.ID,
			Key:         item.Key,
			DisplayName: item.DisplayName,
			Status:      item.Status,
		})
	}

	documentCards := make([]uiDocumentCard, 0, len(documents.Items))
	archivedCount := 0
	for _, item := range documents.Items {
		if item.Status == "archived" {
			archivedCount++
		}
		card := uiDocumentCard{
			ID:                item.ID,
			NamespaceID:       item.NamespaceID,
			NamespaceLabel:    namespaceNames[item.NamespaceID],
			Slug:              item.Slug,
			Title:             item.Title,
			ContentPreview:    item.Content,
			Status:            item.Status,
			CurrentRevisionNo: item.CurrentRevisionNo,
			UpdatedAt:         item.UpdatedAt,
		}
		documentCards = append(documentCards, card)
		for index := range namespaceTree {
			if namespaceTree[index].ID == item.NamespaceID {
				namespaceTree[index].Documents = append(namespaceTree[index].Documents, card)
				break
			}
		}
	}

	var flash *uiFlash
	if message := c.Query("message"); message != "" {
		flash = &uiFlash{
			Kind:    c.DefaultQuery("kind", "info"),
			Message: message,
		}
	}

	return uiIndexData{
		TenantID:            tenantID,
		InstallBaseURL:      baseURL,
		InstallDocURL:       baseURL + "/install/LLM-Wiki.md",
		InstallScriptURL:    baseURL + "/install/install-cli.sh",
		InstallSkillURL:     baseURL + "/install/skills/LLM-Wiki.skill",
		MCPURL:              baseURL + "/mcp",
		Spaces:              spaceCards,
		Namespaces:          namespaceCards,
		NamespaceTree:       namespaceTree,
		Documents:           documentCards,
		SelectedDocument:    selectedDocument,
		SelectedNamespaceID: selectedNamespaceID,
		StatusFilter:        c.DefaultQuery("status", "all"),
		DocumentCount:       len(documentCards),
		ArchivedCount:       archivedCount,
		NamespaceCount:      len(namespaceCards),
		Flash:               flash,
	}, nil
}
