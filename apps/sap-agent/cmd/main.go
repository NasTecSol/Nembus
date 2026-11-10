package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/nembus"
	"github.com/NasTecSol/nembus-sap-agent/internal/pipeline"
	"github.com/NasTecSol/nembus-sap-agent/internal/sap"
)

const version = "1.0.0"

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix("[SAP-Agent] ")

	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "version" || arg == "--version" || arg == "-v" {
			fmt.Printf("Nembus SAP Sync Agent v%s\n", version)
			return
		}
		if arg == "help" || arg == "--help" || arg == "-h" {
			printHelp()
			return
		}
	}

	command := "daemon"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("🛑 Shutdown signal received, terminating...")
		cancel()
	}()

	// Initialize Clients
	sapClient, err := sap.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize SAP client: %v", err)
	}

	nembusClient, err := nembus.NewClient(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Nembus client: %v", err)
	}
	defer nembusClient.Close()

	downstream := pipeline.NewDownstreamSync(cfg, sapClient, nembusClient)
	upstream := pipeline.NewUpstreamSync(cfg, sapClient, nembusClient)

	switch command {
	case "test-connection", "test":
		runTestConnection(ctx, sapClient, nembusClient)

	case "sync-downstream", "downstream":
		log.Println("Running one-shot downstream sync (SAP -> Nembus)...")
		if _, err := downstream.SyncAll(ctx, true); err != nil {
			log.Fatalf("Downstream sync failed: %v", err)
		}

	case "sync-upstream", "upstream":
		log.Println("Running one-shot upstream sync (Nembus POS -> SAP)...")
		if _, err := upstream.ProcessOutbox(ctx); err != nil {
			log.Fatalf("Upstream sync failed: %v", err)
		}

	case "sync-all", "sync":
		log.Println("Running full bidirectional sync...")
		if _, err := downstream.SyncAll(ctx, true); err != nil {
			log.Printf("Downstream sync error: %v", err)
		}
		if _, err := upstream.ProcessOutbox(ctx); err != nil {
			log.Printf("Upstream sync error: %v", err)
		}

	case "daemon":
		runDaemon(ctx, cfg, downstream, upstream)

	default:
		printHelp()
	}
}

func runTestConnection(ctx context.Context, sapClient *sap.Client, nembusClient *nembus.Client) {
	log.Println("🔍 Testing connectivity to SAP Business One and Nembus Database...")

	if err := sapClient.TestConnection(ctx); err != nil {
		log.Printf("❌ SAP B1 Connection: FAILED (%v)", err)
	} else {
		log.Println(" SAP B1 Service Layer Connection: SUCCESS")
	}

	if err := nembusClient.TestConnection(ctx); err != nil {
		log.Printf("❌ Nembus Database Connection: FAILED (%v)", err)
	} else {
		log.Println(" Nembus Database Connection: SUCCESS")
	}
}

func runDaemon(ctx context.Context, cfg *config.Config, downstream *pipeline.DownstreamSync, upstream *pipeline.UpstreamSync) {
	log.Printf("🚀 Starting SAP Sync Agent Daemon v%s (Tenant ID: %d, Store ID: %d)",
		version, cfg.NembusOrganizationID, cfg.NembusStoreID)
	log.Printf("⚙️ Downstream Master Data Interval: %ds | Upstream Outbox Poll: %dms",
		cfg.DownstreamIntervalSeconds, cfg.UpstreamOutboxPollMs)

	// 1. Initial downstream master data sync
	go func() {
		log.Println("Performing initial master data sync on startup...")
		if _, err := downstream.SyncAll(ctx, false); err != nil {
			log.Printf("⚠️ Initial downstream sync error: %v", err)
		}
	}()

	// 2. Downstream Ticker (Periodic Cron / Interval)
	downstreamTicker := time.NewTicker(time.Duration(cfg.DownstreamIntervalSeconds) * time.Second)
	defer downstreamTicker.Stop()

	// 3. Upstream Outbox Watcher (High-frequency Poll)
	upstreamTicker := time.NewTicker(time.Duration(cfg.UpstreamOutboxPollMs) * time.Millisecond)
	defer upstreamTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Daemon stopped.")
			return

		case <-downstreamTicker.C:
			if _, err := downstream.SyncAll(ctx, false); err != nil {
				log.Printf("⚠️ Downstream sync periodic error: %v", err)
			}

		case <-upstreamTicker.C:
			if _, err := upstream.ProcessOutbox(ctx); err != nil {
				log.Printf("⚠️ Upstream outbox processing error: %v", err)
			}
		}
	}
}

func printHelp() {
	fmt.Printf(`Nembus SAP Sync Agent v%s

Usage:
  sap-agent [command]

Available Commands:
  daemon           Run continuous background daemon with cron sync & outbox watcher (default)
  sync             Run one-shot full bidirectional synchronization
  sync-downstream  Run one-shot master data sync (SAP -> Nembus)
  sync-upstream    Run one-shot transaction posting (Nembus POS -> SAP)
  test-connection  Validate SAP Service Layer and Nembus DB connections
  version          Show application version

`, version)
}
