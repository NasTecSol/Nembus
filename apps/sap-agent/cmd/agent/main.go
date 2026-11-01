package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/kardianos/service"

	"github.com/NasTecSol/nembus-sap-agent/internal/cloudgrpc"
	"github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap-agent/internal/etl"
	"github.com/NasTecSol/nembus-sap-agent/internal/pipeline"
	"github.com/NasTecSol/nembus-sap-agent/internal/sap"
	"github.com/NasTecSol/nembus-sap-agent/internal/scheduler"
	"github.com/NasTecSol/nembus-sap-agent/internal/server"
	"github.com/NasTecSol/nembus-sap-agent/internal/updater"
	"github.com/NasTecSol/nembus-sap-agent/ui"
)

var (
	Version      = "v1.0.0"
	AppUpdateKey = ""
	CloudURL     = ""
)

type program struct {
	srv       *server.Server
	sched     *scheduler.SyncScheduler
	cancel    context.CancelFunc
	port      int
	sqlite    *db.SQLiteStore
	mssql     *db.MSSQLClient
	cloud     *cloudgrpc.Client
	cfg       *config.AgentConfig
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	addr := fmt.Sprintf("http://127.0.0.1:%d", p.port)
	log.Printf("[SAP-AGENT] Starting Nembus SAP B1 Agent on %s", addr)

	if p.cfg.OnboardingCompleted && p.sched != nil {
		go p.sched.StartBackground(ctx)
	}

	if service.Interactive() {
		go func() {
			time.Sleep(800 * time.Millisecond)
			openBrowser(addr)
		}()
	}

	go p.pollUpdates()

	if err := p.srv.Start(p.port); err != nil {
		log.Fatalf("[SAP-AGENT] Server failed: %v", err)
	}
}

func (p *program) pollUpdates() {
	go func() {
		time.Sleep(15 * time.Second)
		ticker := time.NewTicker(2 * time.Hour)
		defer ticker.Stop()
		for {
			currVer := p.cfg.Version
			if currVer == "" {
				currVer = "v1.0.0"
			}
			info, err := updater.CheckForUpdate(currVer, p.cfg.Cloud.BaseURL, p.cfg.Cloud.AppUpdateClientKey)
			if err == nil && info != nil && info.HasUpdate && info.IsMandatory {
				_ = updater.ApplyUpdate(info.DownloadURL, info.ChecksumSHA256, p.cfg.Cloud.AppUpdateClientKey, !service.Interactive(), "NembusSAPMigrationAgent")
			}
			<-ticker.C
		}
	}()
}

func (p *program) Stop(s service.Service) error {
	if p.cancel != nil {
		p.cancel()
	}
	if p.cloud != nil {
		_ = p.cloud.Close()
	}
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
	updater.CleanupOldExecutables()

	svcFlag := flag.String("service", "", "Control the Windows service (install|uninstall|start|stop|status)")
	portFlag := flag.Int("port", 17890, "HTTP server port")
	configFlag := flag.String("config", "agent_config.json", "Path to agent configuration file")
	flag.Parse()

	args := flag.Args()
	command := "service"
	if len(args) > 0 {
		command = args[0]
	}

	cfg, err := config.LoadConfig(*configFlag)
	if err != nil {
		log.Printf("[SAP-AGENT] Warning: %v. Using defaults.", err)
		cfg = config.DefaultConfig()
	}
	if *portFlag != 17890 {
		cfg.Port = *portFlag
	}
	cfg.Version = Version
	if CloudURL != "" {
		cfg.Cloud.BaseURL = CloudURL
	}
	if AppUpdateKey != "" {
		cfg.Cloud.AppUpdateClientKey = AppUpdateKey
	}

	switch command {
	case "version", "-v", "--version":
		fmt.Printf("Nembus SAP Agent %s\n", Version)
		return
	case "help", "-h", "--help":
		printHelp()
		return
	case "sync-downstream", "downstream", "sync-upstream", "upstream", "sync-all", "sync", "test-connection", "test":
		runCLI(command, cfg)
		return
	}

	sqliteStore, err := db.NewSQLiteStore(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("[SAP-AGENT] SQLite init failed: %v", err)
	}

	mssqlClient, _ := db.NewMSSQLClient(cfg.MSSQL)

	cloudClient, err := cloudgrpc.NewClient(cfg)
	if err != nil {
		log.Fatalf("[SAP-AGENT] Cloud gRPC client failed: %v", err)
	}

	sapClient, err := sap.NewClientFromAgent(cfg)
	if err != nil {
		log.Printf("[SAP-AGENT] SAP Service Layer client: %v", err)
	}

	engine := etl.NewEngine(cfg, mssqlClient, sqliteStore, cloudClient)
	sched := scheduler.NewSyncScheduler(cfg, cloudClient, sapClient)
	srv := server.NewServer(cfg, engine, sqliteStore, mssqlClient, cloudClient, sched, ui.FS)

	prg := &program{
		srv: srv, sched: sched, port: cfg.Port,
		sqlite: sqliteStore, mssql: mssqlClient, cloud: cloudClient, cfg: cfg,
	}

	svcConfig := &service.Config{
		Name: "NembusSAPMigrationAgent", DisplayName: "Nembus SAP B1 Migration Agent",
		Description: "SAP B1 edge agent for Nembus cloud migration and sync.",
	}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatalf("[SAP-AGENT] service init: %v", err)
	}

	if *svcFlag != "" {
		if err := service.Control(s, *svcFlag); err != nil {
			log.Fatalf("[SAP-AGENT] service control: %v", err)
		}
		return
	}

	if err := s.Run(); err != nil {
		log.Fatalf("[SAP-AGENT] service run: %v", err)
	}
}

func runCLI(command string, cfg *config.AgentConfig) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cloudClient, err := cloudgrpc.NewClient(cfg)
	if err != nil {
		log.Fatalf("cloud grpc: %v", err)
	}
	defer cloudClient.Close()

	sapClient, err := sap.NewClientFromAgent(cfg)
	if err != nil {
		log.Fatalf("sap client: %v", err)
	}

	settings := pipeline.FromAgent(cfg)
	down := pipeline.NewDownstreamSync(settings, sapClient, cloudClient)
	up := pipeline.NewUpstreamSync(settings, sapClient, cloudClient)

	switch command {
	case "test-connection", "test":
		if err := sapClient.TestConnection(ctx); err != nil {
			log.Printf("SAP: FAILED (%v)", err)
		} else {
			log.Println("SAP Service Layer: OK")
		}
		if err := cloudClient.TestConnection(ctx); err != nil {
			log.Printf("Cloud gRPC: FAILED (%v)", err)
		} else {
			log.Println("Cloud gRPC tenant resolve: OK")
		}
	case "sync-downstream", "downstream":
		if _, err := down.SyncAll(ctx, true); err != nil {
			log.Fatalf("downstream: %v", err)
		}
	case "sync-upstream", "upstream":
		if _, err := up.ProcessOutbox(ctx); err != nil {
			log.Fatalf("upstream: %v", err)
		}
	case "sync-all", "sync":
		_, _ = down.SyncAll(ctx, true)
		_, _ = up.ProcessOutbox(ctx)
	}
}

func printHelp() {
	fmt.Printf(`Nembus SAP Agent %s

Usage:
  sap-agent [command] [flags]

Commands:
  (default)          Run Windows service + local UI
  sync-downstream    SAP master data → cloud (gRPC)
  sync-upstream      Cloud POS outbox → SAP
  sync-all           Run both directions once
  test-connection    Test SAP SL and cloud gRPC
  version            Print version

Flags:
  -config path       agent_config.json (default)
  -port N            UI port (default 17890)
  -service ACTION    install|uninstall|start|stop|status

`, Version)
}
