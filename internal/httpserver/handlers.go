package httpserver

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ifuryst/docmesh/internal/api"
	"github.com/ifuryst/docmesh/internal/repository"
	"github.com/ifuryst/docmesh/internal/service"
	"github.com/gin-gonic/gin"
)

const tenantHeader = "X-DocMesh-Tenant-ID"

func registerAPIRoutes(engine *gin.Engine, svc *service.Service) {
	engine.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := svc.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"ok": false, "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	v1 := engine.Group("/v1")
	v1.GET("/spaces", func(c *gin.Context) {
		resp, err := svc.ListSpaces(c.Request.Context(), tenantIDFromRequest(c))
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	v1.GET("/namespaces", func(c *gin.Context) {
		resp, err := svc.ListNamespaces(c.Request.Context(), tenantIDFromRequest(c))
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	v1.POST("/namespaces", func(c *gin.Context) {
		var req api.CreateNamespaceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}

		resp, err := svc.CreateNamespace(c.Request.Context(), tenantIDFromRequest(c), req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusCreated, resp)
	})

	v1.GET("/namespaces/:id", func(c *gin.Context) {
		namespaceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.GetNamespace(c.Request.Context(), tenantIDFromRequest(c), namespaceID)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	v1.POST("/namespaces/:id/archive", func(c *gin.Context) {
		namespaceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.ArchiveNamespace(c.Request.Context(), tenantIDFromRequest(c), namespaceID)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	v1.POST("/documents", func(c *gin.Context) {
		var req api.CreateDocumentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.CreateDocument(c.Request.Context(), tenantIDFromRequest(c), req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusCreated, resp)
	})

	v1.GET("/documents", func(c *gin.Context) {
		var namespaceID *int64
		var status *string
		if raw := c.Query("namespace_id"); raw != "" {
			parsed, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				badRequest(c, err)
				return
			}
			namespaceID = &parsed
		}
		if raw := c.Query("status"); raw != "" {
			status = &raw
		}
		resp, err := svc.ListDocuments(c.Request.Context(), tenantIDFromRequest(c), namespaceID, status)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	v1.GET("/documents/:id", func(c *gin.Context) {
		documentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.GetDocument(c.Request.Context(), tenantIDFromRequest(c), documentID)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	v1.GET("/document-by-slug", func(c *gin.Context) {
		namespaceID, err := strconv.ParseInt(c.Query("namespace_id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		slug := c.Query("slug")
		resp, err := svc.GetDocumentBySlug(c.Request.Context(), tenantIDFromRequest(c), namespaceID, slug)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	v1.PUT("/documents/:id", func(c *gin.Context) {
		documentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		var req api.UpdateDocumentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.UpdateDocument(c.Request.Context(), tenantIDFromRequest(c), documentID, req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	v1.POST("/documents/:id/archive", func(c *gin.Context) {
		documentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		var req api.ArchiveDocumentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.ArchiveDocument(c.Request.Context(), tenantIDFromRequest(c), documentID, req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})
}

func tenantIDFromRequest(c *gin.Context) string {
	if tenantID := c.GetHeader(tenantHeader); tenantID != "" {
		return tenantID
	}
	return "default"
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		c.JSON(http.StatusNotFound, api.ErrorResponse{Error: api.ErrorDetail{Code: "not_found", Message: "resource not found"}})
	case errors.Is(err, repository.ErrConflict):
		c.JSON(http.StatusConflict, api.ErrorResponse{Error: api.ErrorDetail{Code: "conflict", Message: "resource state conflict"}})
	default:
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: api.ErrorDetail{Code: "internal_error", Message: err.Error()}})
	}
}

func badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: api.ErrorDetail{Code: "bad_request", Message: err.Error()}})
}
