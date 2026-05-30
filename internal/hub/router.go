package hub

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dekeky/rssmanager/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type Router struct {
	svc         *Service
	r           *gin.Engine
	uploadToken string
}

func NewRouter(svc *Service, r *gin.Engine, uploadToken string) *Router {
	return &Router{svc: svc, r: r, uploadToken: uploadToken}
}

func (hr *Router) ConfigRouter() {
	group := hr.r.Group("/api/hub")
	group.GET("/categories", hr.listCategories)
	group.GET("/agents", hr.listAgents)
	group.POST("/agents", hr.uploadAuth(), hr.uploadAgent)
	group.PUT("/agents/:agentName", hr.uploadAuth(), hr.updateAgent)
	// Sub-resource routes before /:agentName to avoid route shadowing.
	group.GET("/agents/:agentName/files/*filepath", hr.getFile)
	group.PUT("/agents/:agentName/files/*filepath", hr.uploadAuth(), hr.updateFile)
	group.GET("/agents/:agentName/download", hr.downloadAgent)
	group.DELETE("/agents/:agentName", hr.uploadAuth(), hr.deleteAgent)
	group.GET("/agents/:agentName", hr.getAgent)
}

type listAgentsResp struct {
	Agents []AgentMeta `json:"agents"`
	Total  int         `json:"total"`
}

func (hr *Router) listCategories(c *gin.Context) {
	ginx.NewRender(c).Data(gin.H{"categories": KnownCategories})
}

func (hr *Router) listAgents(c *gin.Context) {
	agents, err := hr.svc.List(c.Query("category"))
	if err != nil {
		status := http.StatusInternalServerError
		if IsInvalidCategory(err) {
			status = http.StatusBadRequest
		}
		ginx.NewRender(c, status).Err(err)
		return
	}
	ginx.NewRender(c).Data(listAgentsResp{Agents: agents, Total: len(agents)})
}

func (hr *Router) getAgent(c *gin.Context) {
	agentName := c.Param("agentName")
	detail, err := hr.svc.Get(agentName, c.Query("version"))
	if err != nil {
		ginx.NewRender(c, http.StatusNotFound).Err(err)
		return
	}
	ginx.NewRender(c).Data(detail)
}

func (hr *Router) getFile(c *gin.Context) {
	agentName := c.Param("agentName")
	path := strings.TrimPrefix(c.Param("filepath"), "/")
	version := c.Query("version")

	content, err := hr.svc.GetFile(agentName, version, path)
	if err != nil {
		ginx.NewRender(c, http.StatusNotFound).Err(err)
		return
	}
	ginx.NewRender(c).Data(gin.H{
		"agentName": agentName,
		"version":   version,
		"path":      path,
		"content":   string(content),
	})
}

func (hr *Router) downloadAgent(c *gin.Context) {
	agentName := c.Param("agentName")
	version := c.Query("version")

	zipPath, cleanup, err := hr.svc.OpenDownload(agentName, version)
	if err != nil {
		ginx.NewRender(c, http.StatusNotFound).Err(err)
		return
	}
	defer cleanup()

	filename := fmt.Sprintf("%s.zip", agentName)
	c.FileAttachment(zipPath, filename)
}

type uploadAgentResp struct {
	Agent AgentMeta `json:"agent"`
}

type updateAgentReq struct {
	DisplayName string `json:"displayName"`
	Summary     string `json:"summary"`
	Category    string `json:"category"`
}

func (hr *Router) updateAgent(c *gin.Context) {
	agentName := c.Param("agentName")
	var req updateAgentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.NewRender(c, http.StatusBadRequest).Err(fmt.Errorf("invalid request body"))
		return
	}
	meta, err := hr.svc.UpdateMeta(agentName, req.DisplayName, req.Summary, req.Category)
	if err != nil {
		ginx.NewRender(c, http.StatusBadRequest).Err(err)
		return
	}
	ginx.NewRender(c).Data(meta)
}

type updateFileReq struct {
	Version string `json:"version"`
	Content string `json:"content"`
}

func (hr *Router) updateFile(c *gin.Context) {
	agentName := c.Param("agentName")
	filePath := strings.TrimPrefix(c.Param("filepath"), "/")
	var req updateFileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.NewRender(c, http.StatusBadRequest).Err(fmt.Errorf("invalid request body"))
		return
	}
	if req.Version == "" {
		ginx.NewRender(c, http.StatusBadRequest).Err(fmt.Errorf("version is required"))
		return
	}
	if err := hr.svc.WriteFile(agentName, req.Version, filePath, []byte(req.Content)); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		ginx.NewRender(c, status).Err(err)
		return
	}
	ginx.NewRender(c).Data(gin.H{
		"agentName": agentName,
		"version":   req.Version,
		"path":      filePath,
		"updated":   true,
	})
}

func (hr *Router) deleteAgent(c *gin.Context) {
	agentName := c.Param("agentName")
	if err := hr.svc.Delete(agentName); err != nil {
		ginx.NewRender(c, http.StatusNotFound).Err(err)
		return
	}
	ginx.NewRender(c).Data(gin.H{"deleted": true})
}

func (hr *Router) uploadAgent(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		ginx.NewRender(c, http.StatusBadRequest).Err(fmt.Errorf("file is required"))
		return
	}

	meta, err := hr.svc.Upload(UploadInput{
		AgentName: c.PostForm("agentName"),
		Category:  c.PostForm("category"),
		Version:   c.PostForm("version"),
		File:      file,
	})
	if err != nil {
		ginx.NewRender(c, http.StatusBadRequest).Err(err)
		return
	}
	ginx.NewRender(c, http.StatusCreated).Data(uploadAgentResp{Agent: meta})
}
