package httpserver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ifuryst/llm-wiki/internal/config"
	"github.com/ifuryst/llm-wiki/internal/mcpserver"
	"github.com/ifuryst/llm-wiki/internal/service"
	"github.com/ifuryst/llm-wiki/internal/ui"
	"github.com/ifuryst/llm-wiki/internal/version"
	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
}

func New(cfg config.Config, logger *zap.Logger, svc *service.Service) *Server {
	handler := NewHandler(cfg, logger, svc)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
		},
		logger: logger,
	}
}

func NewHandler(cfg config.Config, logger *zap.Logger, svc *service.Service) http.Handler {
	gin.SetMode(resolveGinMode(cfg.Environment))
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(requestLogger(logger))
	if templates, err := ui.ParseTemplates(); err == nil {
		engine.SetHTMLTemplate(templates)
	}
	if assetFS, err := ui.AssetFS(); err == nil {
		engine.StaticFS("/ui/assets", assetFS)
	}
	mcpManager := mcpserver.NewManager(svc)

	registerRoutes(engine, cfg)
	registerAPIRoutes(engine, svc)
	registerAuthBrowserRoutes(engine, svc, cfg)
	registerMCPRoutes(engine, svc, mcpManager)
	registerInstallRoutes(engine, cfg)
	registerUIRoutes(engine, svc, cfg)
	return engine
}

func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.logger.Info("http server listening", zap.String("addr", s.httpServer.Addr))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.logger.Info("shutting down http server")
		return s.httpServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func registerRoutes(engine *gin.Engine, cfg config.Config) {
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"ok":          true,
			"service":     version.Name,
			"version":     version.Version,
			"environment": cfg.Environment,
		})
	})

	engine.GET("/v1/system/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":        version.Name,
			"version":     version.Version,
			"environment": cfg.Environment,
			"server": gin.H{
				"host": cfg.Server.Host,
				"port": cfg.Server.Port,
			},
		})
	})
}

func resolveGinMode(environment string) string {
	if environment == "production" {
		return gin.ReleaseMode
	}
	return gin.DebugMode
}

func requestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()
		logger.Info(
			"http request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(startedAt)),
		)
	}
}

func NewTestServer(cfg config.Config, logger *zap.Logger, svc *service.Service) *httptestServer {
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	server := &http.Server{
		Handler:           NewHandler(cfg, logger, svc),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return &httptestServer{
		server:   server,
		listener: listener,
	}
}

type httptestServer struct {
	server   *http.Server
	listener net.Listener
}

func (s *httptestServer) URL() string {
	return "http://" + s.listener.Addr().String()
}

func (s *httptestServer) Start() {
	go func() {
		_ = s.server.Serve(s.listener)
	}()
}

func (s *httptestServer) Close(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
