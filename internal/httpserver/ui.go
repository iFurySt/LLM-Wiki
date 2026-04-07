package httpserver

import (
	"net/http"
	"strconv"

	"github.com/bytedance/docmesh/internal/api"
	"github.com/bytedance/docmesh/internal/service"
	"github.com/gin-gonic/gin"
)

type uiIndexData struct {
	TenantID   string
	Spaces     []api.SpaceResponse
	Namespaces []api.NamespaceResponse
	Documents  []api.DocumentResponse
}

func registerUIRoutes(engine *gin.Engine, svc *service.Service) {
	engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/ui")
	})

	engine.GET("/ui", func(c *gin.Context) {
		tenantID := tenantIDFromRequest(c)

		spaces, err := svc.ListSpaces(c.Request.Context(), tenantID)
		if err != nil {
			handleError(c, err)
			return
		}
		namespaces, err := svc.ListNamespaces(c.Request.Context(), tenantID)
		if err != nil {
			handleError(c, err)
			return
		}
		documents, err := svc.ListDocuments(c.Request.Context(), tenantID, nil, nil)
		if err != nil {
			handleError(c, err)
			return
		}

		c.HTML(http.StatusOK, "index.html", uiIndexData{
			TenantID:   tenantID,
			Spaces:     spaces.Items,
			Namespaces: namespaces.Items,
			Documents:  documents.Items,
		})
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
		c.Redirect(http.StatusSeeOther, "/ui")
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
		c.Redirect(http.StatusSeeOther, "/ui")
	})
}
