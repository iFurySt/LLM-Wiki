package httpserver

import (
	"archive/zip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ifuryst/docmesh/internal/config"
)

func registerInstallRoutes(engine *gin.Engine, cfg config.Config) {
	engine.GET("/install", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/install/DocMesh.md")
	})

	engine.GET("/install/DocMesh.md", func(c *gin.Context) {
		serveInstallMarkdown(c, cfg)
	})

	engine.GET("/install/install-cli.sh", func(c *gin.Context) {
		serveInstallFile(c, "install/install-cli.sh", "text/x-shellscript; charset=utf-8")
	})

	engine.GET("/install/skills/DocMesh.zip", func(c *gin.Context) {
		serveSkillArchive(c, "DocMesh.zip")
	})

	engine.GET("/install/skills/DocMesh.skill", func(c *gin.Context) {
		serveSkillArchive(c, "DocMesh.skill")
	})

	engine.StaticFS("/install/releases", gin.Dir(resolveInstallPath("dist/install/releases"), false))
}

func serveInstallMarkdown(c *gin.Context, cfg config.Config) {
	target := resolveInstallPath("install/DocMesh.md")
	body, err := os.ReadFile(target)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "install asset not found",
			"path":  "install/DocMesh.md",
		})
		return
	}

	baseURL := strings.TrimRight(cfg.Install.BaseURL, "/")
	replacer := strings.NewReplacer(
		"{{DOCMESH_BASE_URL}}", baseURL,
		"{{DOCMESH_INSTALL_DOC_URL}}", baseURL+"/install/DocMesh.md",
		"{{DOCMESH_INSTALL_SCRIPT_URL}}", baseURL+"/install/install-cli.sh",
	)

	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.String(http.StatusOK, replacer.Replace(string(body)))
}

func serveInstallFile(c *gin.Context, relativePath string, contentType string) {
	target := resolveInstallPath(relativePath)
	if _, err := os.Stat(target); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "install asset not found",
			"path":  relativePath,
		})
		return
	}
	c.Header("Content-Type", contentType)
	c.File(target)
}

func resolveInstallPath(relativePath string) string {
	if filepath.IsAbs(relativePath) {
		return relativePath
	}
	candidate := filepath.Clean(relativePath)
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return candidate
	}

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	fallback := filepath.Join(repoRoot, relativePath)
	if _, err := os.Stat(fallback); err == nil {
		return fallback
	}
	return candidate
}

func serveSkillArchive(c *gin.Context, fileName string) {
	root := resolveInstallPath("skills/docmesh")
	if _, err := os.Stat(root); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "skill directory not found",
			"path":  "skills/docmesh",
		})
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", `attachment; filename="`+fileName+`"`)

	writer := zip.NewWriter(c.Writer)
	defer writer.Close()

	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		archivePath := filepath.ToSlash(filepath.Join("docmesh", relative))
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = archivePath
		header.Method = zip.Deflate

		entryWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(entryWriter, file)
		return err
	}); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "broken pipe") {
			status = http.StatusBadGateway
		}
		c.Status(status)
		return
	}
}
