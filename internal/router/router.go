package router

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/agenthub/internal/config"
	"github.com/agenthub/internal/hub"
	"github.com/dekeky/rssmanager/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type AppRouter struct {
	cfg config.Server
	r   *gin.Engine
	svc *hub.Service
}

func New(cfg config.Server, svc *hub.Service) *AppRouter {
	return &AppRouter{cfg: cfg, svc: svc}
}

func (ar *AppRouter) Init() error {
	gin.SetMode(gin.ReleaseMode)
	ar.r = gin.New()
	ar.r.Use(ginx.Recover())
	ar.r.Use(corsMiddleware())
	ar.r.MaxMultipartMemory = 200 << 20 // 200MB

	hub.NewRouter(ar.svc, ar.r, ar.cfg.UploadToken).ConfigRouter()

	ar.r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 静态文件服务 - Vite 构建的前端界面
	distDir := filepath.Join(ar.cfg.StorageDir, "..", "web", "dist")
	assetsDir := filepath.Join(distDir, "assets")
	if info, err := os.Stat(assetsDir); err == nil && info.IsDir() {
		ar.r.Static("/assets", assetsDir)
	}
	// Serve favicon and other root-level static assets
	ar.r.StaticFile("/favicon.ico", filepath.Join(distDir, "favicon.ico"))
	ar.r.StaticFile("/vite.svg", filepath.Join(distDir, "vite.svg"))

	// SPA fallback: serve index.html for any non-API route.
	// Inject upload token as a <meta> tag so the frontend can use it.
	ar.r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/health") {
			ginx.NewRender(c, http.StatusNotFound).Err(fmt.Errorf("not found"))
			return
		}
		indexPath := filepath.Join(distDir, "index.html")
		raw, err := os.ReadFile(indexPath)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "errMsg": "frontend not built - run 'cd web && npm run build'"})
			return
		}
		html := string(raw)
		if ar.cfg.UploadToken != "" {
			metaTag := fmt.Sprintf(`<meta name="upload-token" content="%s">`, ar.cfg.UploadToken)
			html = strings.Replace(html, "</head>", metaTag+"\n</head>", 1)
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	log.Println("routes initialized")
	return nil
}

func (ar *AppRouter) Run() error {
	uploadAuth := "disabled"
	if ar.cfg.UploadToken != "" {
		uploadAuth = "enabled"
	}
	log.Printf("agenthub listening on %s (storage: %s, upload auth: %s)", ar.cfg.Addr, ar.cfg.StorageDir, uploadAuth)
	return ar.r.Run(ar.cfg.Addr)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Upload-Token")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func MustInitStore(cfg config.Server) (*hub.Store, error) {
	store, err := hub.NewStore(cfg.StorageDir)
	if err != nil {
		return nil, fmt.Errorf("init store: %w", err)
	}
	return store, nil
}
