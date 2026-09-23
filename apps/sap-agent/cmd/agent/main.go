package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"

	"github.com/kardianos/service"


	"github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap-agent/internal/etl"
	"github.com/NasTecSol/nembus-sap-agent/internal/server"
	"github.com/NasTecSol/nembus-sap-agent/internal/transport"
	"github.com/NasTecSol/nembus-sap-agent/ui"
)

type program struct {
	srv    *server.Server
	port   int
	sqlite *db.SQLiteStore
	mssql  *db.MSSQLClient
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	addr := fmt.Sprintf("http://127.0.0.1:%d", p.port)
	log.Printf("[SAP-AGENT] Starting Nembus SAP B1 Migration Agent on %s", addr)

	// If interactive (not running as headless Windows Service), auto-open browser after a short delay
	if service.Interactive() {
		go func() {
			time.Sleep(800 * time.Millisecond)
			openBrowser(addr)
		}()
	}

	if err := p.srv.Start(p.port); err != nil {
		log.Fatalf("[SAP-AGENT] Server failed: %v", err)
	}
}

func (p *program) Stop(s service.Service) error {
	log.Println("[SAP-AGENT] Stopping agent service...")
	if p.sqlite != nil {
		_ = p.sqlite.Close()
	}
	if p.mssql != nil {
		_ = p.mssql.Close()
	}
	return nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func main() {
	svcFlag := flag.String("service", "", "Control the Windows service ('install', 'uninstall', 'start', 'stop', 'status')")
	portFlag := flag.Int("port", 17890, "HTTP server port (default 17890)")
	configFlag := flag.String("config", "agent_config.json", "Path to agent configuration file")
	flag.Parse()

	// 1. Load Agent Configuration
	cfg, err := config.LoadConfig(*configFlag)
	if err != nil {
		log.Printf("[SAP-AGENT] Warning: failed to load config file: %v. Using defaults.", err)
		cfg = config.DefaultConfig()
	}
	if *portFlag != 17890 {
		cfg.Port = *portFlag
	}

	// 2. Initialize Local SQLite Checkpoint Storage
	sqliteStore, err := db.NewSQLiteStore(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("[SAP-AGENT] Failed to initialize SQLite storage at %s: %v", cfg.SQLitePath, err)
	}

	// 3. Optional SQL Server client init
	mssqlClient, err := db.NewMSSQLClient(cfg.MSSQL)
	if err != nil {
		log.Printf("[SAP-AGENT] Info: SQL Server connection not yet established (%v). Will configure via GUI.", err)
	}

	// 4. Initialize Cloud Transport and ETL Engine
	cloudClient := transport.NewCloudClient(cfg.Cloud)
	engine := etl.NewEngine(cfg, mssqlClient, sqliteStore, cloudClient)

	// 5. Initialize Embedded Server with Web UI
	srv := server.NewServer(cfg, engine, sqliteStore, mssqlClient, cloudClient, ui.FS)

	prg := &program{
		srv:    srv,
		port:   cfg.Port,
		sqlite: sqliteStore,
		mssql:  mssqlClient,
	}

	svcConfig := &service.Config{
		Name:        "NembusSAPMigrationAgent",
		DisplayName: "Nembus SAP B1 Migration Agent",
		Description: "Local native Windows extraction and migration agent for SAP Business One to Nembus Cloud ERP.",
	}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatalf("[SAP-AGENT] Service initialization error: %v", err)
	}

	if len(*svcFlag) != 0 {
		err := service.Control(s, *svcFlag)
		if err != nil {
			log.Fatalf("[SAP-AGENT] Failed to execute service action '%s': %v", *svcFlag, err)
		}
		fmt.Printf("[SAP-AGENT] Service action '%s' succeeded.\n", *svcFlag)
		return
	}

	if err := s.Run(); err != nil {
		log.Fatalf("[SAP-AGENT] Service run error: %v", err)
	}
}
