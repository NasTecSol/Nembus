package server

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"


	"github.com/NasTecSol/nembus-sap/contracts"
	"github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap-agent/internal/discovery"
	"github.com/NasTecSol/nembus-sap-agent/internal/etl"
	"github.com/NasTecSol/nembus-sap-agent/internal/reconciliation"
	"github.com/NasTecSol/nembus-sap-agent/internal/transport"
)

type Server struct {
	cfg         *config.AgentConfig
	engine      *etl.Engine
	sqlite      *db.SQLiteStore
	mssql       *db.MSSQLClient
	cloudClient *transport.CloudClient
	router      *gin.Engine
	uiFS        embed.FS
}

func NewServer(cfg *config.AgentConfig, engine *etl.Engine, sqlite *db.SQLiteStore, mssql *db.MSSQLClient, cloudClient *transport.CloudClient, uiFS embed.FS) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	s := &Server{
		cfg:         cfg,
		engine:      engine,
		sqlite:      sqlite,
		mssql:       mssql,
		cloudClient: cloudClient,
		router:      r,
		uiFS:        uiFS,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// 1. Static embedded UI
	s.router.GET("/", func(c *gin.Context) {
		data, err := s.uiFS.ReadFile("index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load index.html: %v", err)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})
	s.router.GET("/style.css", func(c *gin.Context) {
		data, _ := s.uiFS.ReadFile("style.css")
		c.Data(http.StatusOK, "text/css; charset=utf-8", data)
	})
	s.router.GET("/app.js", func(c *gin.Context) {
		data, _ := s.uiFS.ReadFile("app.js")
		c.Data(http.StatusOK, "application/javascript; charset=utf-8", data)
	})


	// 2. WebSocket Route
	s.router.GET("/ws", s.handleWebSocket)

	// 3. REST API Routes
	api := s.router.Group("/api/v1")
	{
		// Config
		api.GET("/config", s.handleGetConfig)
		api.POST("/config", s.handleSaveConfig)

		// Test Connection
		api.POST("/test-connection/mssql", s.handleTestMSSQL)
		api.POST("/test-connection/cloud", s.handleTestCloud)

		// Discovery
		api.POST("/discovery", s.handleDiscovery)

		// Migration Control
		api.POST("/migration/start", s.handleStartMigration)
		api.POST("/migration/cancel", s.handleCancelMigration)
		api.GET("/migration/status", s.handleGetMigrationStatus)

		// Reconciliation
		api.POST("/reconciliation", s.handleReconciliation)

		// History
		api.GET("/history", s.handleGetHistory)
		api.GET("/logs", s.handleGetLogs)
	}
}

func (s *Server) Start(port int) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	return s.router.Run(addr)
}

// Handlers

func (s *Server) handleGetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, config.Get())
}

func (s *Server) handleSaveConfig(c *gin.Context) {
	var req config.AgentConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	current := config.Get()
	if req.MSSQL.Host != "" {
		current.MSSQL = req.MSSQL
	}
	if req.Cloud.BaseURL != "" {
		current.Cloud = req.Cloud
	}
	if req.BatchSize > 0 {
		current.BatchSize = req.BatchSize
	}

	if err := config.SaveConfig(current); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "config": current})
}

func (s *Server) handleTestMSSQL(c *gin.Context) {
	cfg := config.Get().MSSQL
	var req config.MSSQLConfig
	if err := c.ShouldBindJSON(&req); err == nil && req.Host != "" {
		cfg = req
	}

	client, err := db.NewMSSQLClient(cfg)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(c.Request.Context(), 6*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "SQL Server connection verified."})
}

func (s *Server) handleTestCloud(c *gin.Context) {
	cfg := config.Get().Cloud
	var req config.CloudConfig
	if err := c.ShouldBindJSON(&req); err == nil && req.BaseURL != "" {
		cfg = req
	}

	client := transport.NewCloudClient(cfg)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	ok, msg, err := client.PingCloud(ctx)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": ok, "message": msg})
}


func (s *Server) handleDiscovery(c *gin.Context) {
	cfg := config.Get().MSSQL
	client, err := db.NewMSSQLClient(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to MSSQL: " + err.Error()})
		return
	}
	defer client.Close()

	disc := discovery.NewDiscoveryEngine(client)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	res, err := disc.Discover(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (s *Server) handleStartMigration(c *gin.Context) {
	var req struct {
		Mode    contracts.MigrationMode `json:"mode"`
		Domains []contracts.DomainType  `json:"domains"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	run, err := s.engine.StartMigration(c.Request.Context(), req.Mode, req.Domains)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "run_id": run.ID})
}

func (s *Server) handleCancelMigration(c *gin.Context) {
	if err := s.engine.CancelMigration(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (s *Server) handleGetMigrationStatus(c *gin.Context) {
	run, err := s.sqlite.GetLatestRun(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"active": false})
		return
	}

	steps, _ := s.sqlite.GetSteps(c.Request.Context(), run.ID)
	c.JSON(http.StatusOK, gin.H{
		"active": run.Status == contracts.StatusRunning,
		"run":    run,
		"steps":  steps,
	})
}

func (s *Server) handleReconciliation(c *gin.Context) {
	cfg := config.Get()
	mssql, err := db.NewMSSQLClient(cfg.MSSQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to SAP MSSQL: " + err.Error()})
		return
	}
	defer mssql.Close()

	// Accept explicit run_id; fall back to latest run (handled inside Reconcile)
	runID := c.Query("run_id")

	cloudClient := transport.NewCloudClient(cfg.Cloud)
	auditor := reconciliation.NewAuditEngine(mssql, cloudClient, s.sqlite)

	report, err := auditor.Reconcile(c.Request.Context(), runID, cfg.Cloud.OrganizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (s *Server) handleGetHistory(c *gin.Context) {
	runs, err := s.sqlite.GetRuns(c.Request.Context(), 50, 0)
	if err != nil || len(runs) == 0 {
		c.JSON(http.StatusOK, gin.H{"runs": []interface{}{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"runs": runs})
}

func (s *Server) handleGetLogs(c *gin.Context) {
	runID := c.Query("run_id")
	if runID == "" {
		latest, err := s.sqlite.GetLatestRun(c.Request.Context())
		if err == nil && latest != nil {
			runID = latest.ID
		}
	}

	logs, err := s.sqlite.GetRecentLogs(c.Request.Context(), runID, 100)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"logs": []interface{}{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}
