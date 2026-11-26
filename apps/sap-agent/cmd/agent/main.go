package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kardianos/service"

	"github.com/NasTecSol/nembus-core/telemetry"
	"github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap-agent/internal/etl"
	"github.com/NasTecSol/nembus-sap-agent/internal/server"
	"github.com/NasTecSol/nembus-sap-agent/internal/transport"
	"github.com/NasTecSol/nembus-sap-agent/internal/tray"
	"github.com/NasTecSol/nembus-sap-agent/internal/updater"
	"github.com/NasTecSol/nembus-sap-agent/ui"
)

// Application Version & Update configuration (injected via -ldflags at build time)
var (
	Version         = "v1.0.0"
	AppUpdateKey    = ""
	CloudURL        = ""
	OTelEndpoint    = ""
	OTelIngestToken = ""
)

type program struct {
	srv          *server.Server
	engine       *etl.Engine
	port         int
	sqlite       *db.SQLiteStore
	mssql        *db.MSSQLClient
	cfg          *config.AgentConfig
	otelShutdown func(context.Context) error
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	addr := fmt.Sprintf("http://127.0.0.1:%d", p.port)
	log.Printf("[SAP-AGENT] Starting Nembus SAP B1 Migration Agent %s on %s (config: %s)", p.cfg.Version, addr, config.ConfigPath())

	// Only auto-open browser on first-run setup; once onboarded, run silently in System Tray.
	if service.Interactive() && !p.cfg.OnboardingCompleted {
		go func() {
			time.Sleep(800 * time.Millisecond)
			openBrowser(addr)
		}()
	}

	// Background Auto-Update Polling (every 2 hours)
	go func() {
		time.Sleep(15 * time.Second) // Initial delay on startup
		ticker := time.NewTicker(2 * time.Hour)
		defer ticker.Stop()
		for {
			currCfg := config.Get()
			currVer := currCfg.Version
			if currVer == "" {
				currVer = "v1.0.0"
			}
			info, err := updater.CheckForUpdate(currVer, currCfg.Cloud.BaseURL, currCfg.Cloud.AppUpdateClientKey)
			if err == nil && info != nil && info.HasUpdate {
				log.Printf("[SAP-UPDATER] 🚀 New update available: %s (Current: %s)", info.LatestVersion, currVer)
				if info.IsMandatory {
					if p.engine != nil && p.engine.IsRunning() {
						log.Println("[SAP-UPDATER] Mandatory update detected, but a migration run is active. Deferring until next check.")
					} else {
						log.Println("[SAP-UPDATER] Mandatory update detected. Applying automatically...")
						if applyErr := updater.ApplyUpdateWithHook(
							info.DownloadURL,
							info.ChecksumSHA256,
							currCfg.Cloud.AppUpdateClientKey,
							!service.Interactive(),
							"NembusSAPMigrationAgent",
							func() { _ = p.Stop(nil) },
						); applyErr != nil {
							log.Printf("[SAP-UPDATER] Mandatory auto-update failed: %v", applyErr)
						}
					}
				}
			}
			<-ticker.C
		}
	}()

	if err := p.srv.Start(p.port); err != nil {
		log.Fatalf("[SAP-AGENT] Server failed: %v", err)
	}
}

func (p *program) Stop(s service.Service) error {
	log.Println("[SAP-AGENT] Stopping agent service...")
	if p.srv != nil {
		_ = p.srv.Shutdown(context.Background())
	}
	if p.sqlite != nil {
		_ = p.sqlite.Close()
	}
	if p.mssql != nil {
		_ = p.mssql.Close()
	}
	if p.otelShutdown != nil {
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = p.otelShutdown(shutCtx)
		cancel()
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
	// Clean up legacy executable backups from prior update cycles
	updater.CleanupOldExecutables()

	svcFlag := flag.String("service", "", "Control the Windows service ('install', 'uninstall', 'start', 'stop', 'status')")
	portFlag := flag.Int("port", 17890, "HTTP server port (default 17890)")
	configFlag := flag.String("config", "", "Path to agent configuration file (defaults to persistent OS AppData path)")
	headlessFlag := flag.Bool("headless", false, "Run without Windows System Tray GUI")
	flag.Parse()

	isService := !service.Interactive()

	// Set up persistent rotating file logging so tray/service runs always keep debug logs.
	logPath := config.DefaultLogPath(isService)
	_ = os.MkdirAll(filepath.Dir(logPath), 0755)
	if logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		log.SetOutput(io.MultiWriter(os.Stderr, logFile))
	}

	// Resolve persistent config path & seamlessly migrate legacy working-directory files once.
	cfgPath := *configFlag
	if cfgPath == "" {
		cfgPath = config.DefaultConfigPath(isService)
		config.MigrateLegacyFileIfNeeded("agent_config.json", cfgPath)
		config.MigrateLegacyFileIfNeeded("agent.db", config.DefaultSQLitePath(isService))
	}

	// 1. Load Agent Configuration
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Printf("[SAP-AGENT] Warning: failed to load config file: %v. Using defaults.", err)
		cfg = config.DefaultConfig()
	}
	if *portFlag != 17890 {
		cfg.Port = *portFlag
	}

	// Apply build-time defaults without overwriting user-persisted onboarding URLs/keys
	cfg.Version = Version
	if CloudURL != "" && (cfg.Cloud.BaseURL == "" || (!cfg.OnboardingCompleted && cfg.Cloud.BaseURL == config.DefaultConfig().Cloud.BaseURL)) {
		cfg.Cloud.BaseURL = CloudURL
	}
	if AppUpdateKey != "" && (cfg.Cloud.AppUpdateClientKey == "" || cfg.Cloud.AppUpdateClientKey == "nembus-pos-update-client-key-default") {
		cfg.Cloud.AppUpdateClientKey = AppUpdateKey
	}

	otelShutdown, otelErr := telemetry.Init(context.Background(), telemetry.Config{
		ServiceName:    "nembus-sap-agent",
		ServiceVersion: cfg.Version,
		TenantSlug:     cfg.TenantSlug,
		OTLPEndpoint:   OTelEndpoint,
		IngestToken:    OTelIngestToken,
		SpoolDir:       filepath.Join(os.TempDir(), "nembus-otel-sap"),
		MaxSpoolBytes:  50 * 1024 * 1024,
	})
	if otelErr != nil {
		log.Printf("[Telemetry] Warning: OpenTelemetry init failed: %v", otelErr)
	}

	if cfg.OnboardingCompleted {
		log.Printf("[SAP-AGENT] Resuming persisted configuration for tenant %q (downstream_auto=%v, upstream_auto=%v).",
			cfg.TenantSlug, cfg.DownstreamAutoActive, cfg.UpstreamAutoActive)
	} else {
		log.Println("[SAP-AGENT] Startup complete: waiting for onboarding completion before starting Downstream and Upstream syncs.")
	}

	// 2. Initialize Local SQLite Checkpoint Storage
	sqliteStore, err := db.NewSQLiteStore(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("[SAP-AGENT] Failed to initialize SQLite storage at %s: %v", cfg.SQLitePath, err)
	}

	// 3. SQL Server client will be established upon onboarding completion or on demand
	var mssqlClient *db.MSSQLClient

	// 4. Initialize Cloud Transport and ETL Engine
	cloudClient := transport.NewCloudClient(cfg.Cloud)
	engine := etl.NewEngine(cfg, mssqlClient, sqliteStore, cloudClient)

	// 5. Initialize Embedded Server with Web UI
	srv := server.NewServer(cfg, engine, sqliteStore, mssqlClient, cloudClient, ui.FS)

	// Handle one-shot CLI subcommands if provided (unified entrypoint)
	if args := flag.Args(); len(args) > 0 {
		defer func() {
			if otelShutdown != nil {
				shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_ = otelShutdown(shutCtx)
				cancel()
			}
		}()
		runCLICommand(args[0], cfg, engine, srv, sqliteStore, cloudClient)
		return
	}

	prg := &program{
		srv:          srv,
		engine:       engine,
		port:         cfg.Port,
		sqlite:       sqliteStore,
		mssql:        mssqlClient,
		cfg:          cfg,
		otelShutdown: otelShutdown,
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

	// Interactive desktop mode: run HTTP server in background and host Windows System Tray.
	if service.Interactive() && !*headlessFlag {
		go prg.run()
		addr := fmt.Sprintf("http://127.0.0.1:%d", cfg.Port)
		tray.Run(tray.Config{
			BaseURL:   addr,
			Version:   cfg.Version,
			GetStatus: srv.TrayStatus,
			OnCheckUpdate: func() (string, string, bool) {
				currCfg := config.Get()
				currVer := currCfg.Version
				if currVer == "" {
					currVer = Version
				}
				info, err := updater.CheckForUpdate(currVer, currCfg.Cloud.BaseURL, currCfg.Cloud.AppUpdateClientKey)
				if err != nil {
					log.Printf("[SAP-UPDATER] Manual tray update check failed: %v", err)
					return "Status: Update check failed", addr + "/#check-update", false
				}
				if info != nil && info.HasUpdate {
					log.Printf("[SAP-UPDATER] Tray check found update %s (mandatory=%v)", info.LatestVersion, info.IsMandatory)
					if info.IsMandatory && (engine == nil || !engine.IsRunning()) {
						go func() {
							_ = updater.ApplyUpdateWithHook(
								info.DownloadURL,
								info.ChecksumSHA256,
								currCfg.Cloud.AppUpdateClientKey,
								false,
								"NembusSAPMigrationAgent",
								func() { _ = prg.Stop(nil) },
							)
						}()
						return fmt.Sprintf("Status: Updating to %s...", info.LatestVersion), "", true
					}
					return fmt.Sprintf("Status: Update %s available!", info.LatestVersion), addr + "/#check-update", true
				}
				return fmt.Sprintf("Status: Up to date (%s)", currVer), "", false
			},
			OnExit: func() {
				_ = prg.Stop(nil)
			},
		})
		return
	}

	if err := s.Run(); err != nil {
		log.Fatalf("[SAP-AGENT] Service run error: %v", err)
	}
}

func runCLICommand(cmd string, cfg *config.AgentConfig, engine *etl.Engine, srv *server.Server, sqliteStore *db.SQLiteStore, cloudClient *transport.CloudClient) {
	defer sqliteStore.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	switch cmd {
	case "version":
		fmt.Printf("Nembus SAP Agent %s\n", cfg.Version)
	case "test-connection", "test":
		mssql, err := db.NewMSSQLClient(cfg.MSSQL)
		if err != nil {
			log.Printf("[SAP-AGENT] MSSQL connect failed: %v", err)
		} else {
			defer mssql.Close()
			if err := mssql.Ping(ctx); err != nil {
				log.Printf("[SAP-AGENT] MSSQL ping failed: %v", err)
			} else {
				log.Println("[SAP-AGENT] MSSQL connection: OK")
			}
		}
		ok, msg, err := cloudClient.PingCloud(ctx)
		if err != nil || !ok {
			log.Printf("[SAP-AGENT] Cloud connection failed: ok=%v msg=%s err=%v", ok, msg, err)
		} else {
			log.Printf("[SAP-AGENT] Cloud connection: OK (%s)", msg)
		}
	case "sync-downstream", "downstream":
		stats, err := engine.ExecuteDownstreamSync(ctx, server.DownstreamMasterDomains)
		if err != nil {
			log.Fatalf("[SAP-AGENT] Downstream sync failed: %v", err)
		}
		log.Printf("[SAP-AGENT] Downstream sync complete: %+v", stats)
	case "sync-upstream", "upstream":
		stats, err := srv.RunUpstreamOnce(ctx)
		if err != nil {
			log.Fatalf("[SAP-AGENT] Upstream sync failed: %v", err)
		}
		log.Printf("[SAP-AGENT] Upstream sync complete: %+v", stats)
	default:
		log.Fatalf("[SAP-AGENT] Unknown command %q. Valid commands: version, test-connection, sync-downstream, sync-upstream", cmd)
	}
}
