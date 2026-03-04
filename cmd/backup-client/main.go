// cmd/backup-client is a CLI tool for POS terminals and local offline apps.
// It connects to the NEMBUS gRPC backup service, downloads a full pg_dump
// of the requested tenant's database, and saves it to a local file.
//
// Usage:
//
//	go run ./cmd/backup-client \
//	  -server cloud.nembus.com:50051 \
//	  -tenant store-riyadh-01 \
//	  -token <jwt_token> \
//	  [-compressed]             # optional: pg_dump custom format (smaller, needs pg_restore)
//	  [-out myfile.sql]         # optional: explicit output filename
package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"NEMBUS/internal/grpc/backuppb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	serverAddr := flag.String("server", "localhost:50051", "gRPC server address (host:port)")
	tenantSlug := flag.String("tenant", "", "Tenant slug to back up (required)")
	authToken := flag.String("token", "", "JWT auth token (required)")
	compressed := flag.Bool("compressed", false, "Use pg_dump custom format (smaller, needs pg_restore)")
	outFile := flag.String("out", "", "Output file path (default: auto-generated)")
	flag.Parse()

	if *tenantSlug == "" || *authToken == "" {
		fmt.Fprintln(os.Stderr, "Error: -tenant and -token are required")
		flag.Usage()
		os.Exit(1)
	}

	// ── Determine output filename ────────────────────────────────────────────
	if *outFile == "" {
		ext := ".sql"
		if *compressed {
			ext = ".pgdump"
		}
		*outFile = fmt.Sprintf("backup_%s_%d%s", *tenantSlug, time.Now().Unix(), ext)
	}

	// Connect to gRPC server — generated proto types use the default proto codec
	conn, err := grpc.NewClient(*serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("❌ Failed to connect to %s: %v", *serverAddr, err)
	}
	defer conn.Close()

	client := backuppb.NewBackupServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	// ── Initiate backup stream ───────────────────────────────────────────────
	stream, err := client.StreamBackup(ctx, &backuppb.BackupRequest{
		TenantSlug: *tenantSlug,
		AuthToken:  *authToken,
		Compressed: *compressed,
	})
	if err != nil {
		log.Fatalf("❌ StreamBackup failed: %v", err)
	}

	// ── Open output file ─────────────────────────────────────────────────────
	f, err := os.Create(filepath.Clean(*outFile))
	if err != nil {
		log.Fatalf("❌ Cannot create output file %s: %v", *outFile, err)
	}
	defer f.Close()

	fmt.Printf("📥  Streaming backup → %s\n", *outFile)
	startTime := time.Now()
	var totalBytes uint64
	lastPrint := uint64(0)
	const printEvery = 1 * 1024 * 1024 // print progress every 1 MB

	// ── Receive chunks ────────────────────────────────────────────────────────
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("❌ Stream error: %v", err)
		}

		// Verify checksum if provided by the server
		if chunk.Sha256 != "" {
			h := sha256.Sum256(chunk.Data)
			got := fmt.Sprintf("%x", h)
			if got != chunk.Sha256 {
				log.Fatalf("❌ Checksum mismatch at offset %d: want %s got %s", chunk.Offset, chunk.Sha256, got)
			}
		}

		if len(chunk.Data) > 0 {
			if _, werr := f.Write(chunk.Data); werr != nil {
				log.Fatalf("❌ Write error: %v", werr)
			}
			totalBytes += uint64(len(chunk.Data))
		}

		// Progress reporting
		if totalBytes-lastPrint >= printEvery {
			fmt.Printf("  → %.2f MB received...\n", float64(totalBytes)/(1024*1024))
			lastPrint = totalBytes
		}

		if chunk.IsLast {
			break
		}
	}

	elapsed := time.Since(startTime)

	fmt.Printf("\n✅ Backup complete!\n")
	fmt.Printf("   File    : %s\n", *outFile)
	fmt.Printf("   Size    : %.2f MB (%d bytes)\n", float64(totalBytes)/(1024*1024), totalBytes)
	fmt.Printf("   Elapsed : %s\n", elapsed.Round(time.Second))

	if *compressed {
		fmt.Printf("\nTo restore:\n  pg_restore --dbname=<target_dsn> %s\n", *outFile)
	} else {
		fmt.Printf("\nTo restore:\n  psql <target_dsn> < %s\n", *outFile)
	}
}
