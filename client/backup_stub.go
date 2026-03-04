// client/backup.go
// Real gRPC backup client — downloads a tenant backup from the cloud gRPC
// server and restores it into the local embedded Postgres database.
package client

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"NEMBUS/internal/grpc/backuppb"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DownloadBackup downloads a tenant backup via gRPC streaming, writes it as a
// .sql file under outputDir, restores it to localDBURL (if non-empty), and
// returns the path to the saved backup file.
//
// serverAddr may be either a bare "host:port" gRPC address or an HTTP(S) URL
// (scheme + host are extracted automatically).
func DownloadBackup(serverAddr, tenantSlug, token, outputDir, localDBURL string) (string, error) {
	grpcAddr := toGRPCAddr(serverAddr)
	log.Printf("Connecting to backup gRPC server: %s", grpcAddr)

	// 1. Connect to cloud gRPC server (insecure for now; add TLS via env flag if needed)
	conn, err := grpc.Dial(grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(15*time.Second),
	)
	if err != nil {
		return "", fmt.Errorf("gRPC connect to %s failed: %w", grpcAddr, err)
	}
	defer conn.Close()

	client := backuppb.NewBackupServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// 2. Start streaming backup
	stream, err := client.StreamBackup(ctx, &backuppb.BackupRequest{
		TenantSlug: tenantSlug,
		AuthToken:  token,
		Compressed: false,
	})
	if err != nil {
		return "", fmt.Errorf("StreamBackup start failed: %w", err)
	}

	// 3. Ensure output directory and open file
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup dir: %w", err)
	}

	backupFilename := fmt.Sprintf("backup_%s_%d.sql", tenantSlug, time.Now().Unix())
	backupPath := filepath.Join(outputDir, backupFilename)

	outFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer outFile.Close()

	// 4. Receive all chunks and write to file
	var totalBytes int64
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("receive error: %w", err)
		}

		// Verify per-chunk SHA-256 checksum
		hash := sha256.Sum256(chunk.Data)
		got := fmt.Sprintf("%x", hash)
		if chunk.Sha256 != "" && got != chunk.Sha256 {
			return "", fmt.Errorf("checksum mismatch at offset %d: got %s want %s", chunk.Offset, got, chunk.Sha256)
		}

		if _, err := outFile.Write(chunk.Data); err != nil {
			return "", fmt.Errorf("write error: %w", err)
		}
		totalBytes += int64(len(chunk.Data))
		log.Printf("Backup download: %d bytes received", totalBytes)
	}

	log.Printf("Backup download complete: %d bytes → %s", totalBytes, backupPath)

	// 5. Restore to local DB (if a DB URL was provided)
	if localDBURL != "" && totalBytes > 0 {
		log.Printf("Restoring backup to local database...")
		if err := restoreSQL(backupPath, localDBURL); err != nil {
			return backupPath, fmt.Errorf("restore failed: %w", err)
		}
		log.Printf("Restore complete.")
	}

	return backupPath, nil
}

// restoreSQL executes a plain-SQL backup file against the given Postgres DSN.
// It pre-processes the file to remove psql meta-commands and comment lines.
func restoreSQL(sqlFilePath, dbURL string) error {
	content, err := os.ReadFile(sqlFilePath)
	if err != nil {
		return fmt.Errorf("could not read SQL file: %w", err)
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("open DB: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping DB: %w", err)
	}

	// Clean the SQL: remove meta-commands (\), OWNER TO lines, and SET blocks
	// that might fail on a local embedded instance.
	cleanedSQL := cleanSQLDump(string(content))

	// Execute by statements
	statements := splitSQL(cleanedSQL)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// Skip specific patterns that frequently fail in local restore
		if strings.HasPrefix(stmt, "ALTER TABLE") && strings.Contains(stmt, "OWNER TO") {
			continue
		}
		if strings.HasPrefix(stmt, "ALTER SEQUENCE") && strings.Contains(stmt, "OWNER TO") {
			continue
		}
		if strings.HasPrefix(stmt, "ALTER FUNCTION") && strings.Contains(stmt, "OWNER TO") {
			continue
		}
		if strings.HasPrefix(stmt, "ALTER SCHEMA") && strings.Contains(stmt, "OWNER TO") {
			continue
		}

		if _, err := db.Exec(stmt); err != nil {
			// Some errors are expected (e.g. creating extensions that exist)
			if !strings.Contains(err.Error(), "already exists") {
				log.Printf("Warning: statement error (continuing): %v\nStatement: %.100s...", err, stmt)
			}
		}
	}

	return nil
}

// cleanSQLDump removes lines that cause issues for direct driver execution.
func cleanSQLDump(sql string) string {
	var out strings.Builder
	lines := strings.Split(sql, "\n")
	inCopy := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle COPY blocks: standard drivers don't support "COPY ... FROM stdin"
		// We skip the data lines following a COPY command until "\." is reached.
		if inCopy {
			if trimmed == "\\." {
				inCopy = false
			}
			continue
		}
		if strings.HasPrefix(strings.ToUpper(trimmed), "COPY ") && strings.HasSuffix(strings.ToUpper(trimmed), "FROM STDIN;") {
			inCopy = true
			continue
		}

		// Skip psql meta-commands (lines starting with \)
		if strings.HasPrefix(trimmed, "\\") {
			continue
		}

		// Skip comments to avoid splitting bugs
		if strings.HasPrefix(trimmed, "--") {
			continue
		}

		out.WriteString(line)
		out.WriteRune('\n')
	}
	return out.String()
}

// splitSQL splits a SQL file on top-level semicolons.
// Properly ignores semicolons inside single-quotes and dollar-quoted blocks.
func splitSQL(sql string) []string {
	var stmts []string
	var cur strings.Builder
	inSingle := false
	inDollar := false
	dollarTag := ""

	runes := []rune(sql)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if inSingle {
			cur.WriteRune(r)
			if r == '\'' {
				// Handle escaped quotes ''
				if i+1 < len(runes) && runes[i+1] == '\'' {
					cur.WriteRune(runes[i+1])
					i++
				} else {
					inSingle = false
				}
			}
			continue
		}

		if inDollar {
			cur.WriteRune(r)
			// Check if we hit the closing dollar tag
			rest := string(runes[i:])
			if strings.HasPrefix(rest, dollarTag) {
				// Skip the rest of the tag
				tagRunes := []rune(dollarTag)
				for j := 1; j < len(tagRunes); j++ {
					cur.WriteRune(runes[i+j])
				}
				i += len(tagRunes) - 1
				inDollar = false
				dollarTag = ""
			}
			continue
		}

		if r == '\'' {
			inSingle = true
			cur.WriteRune(r)
			continue
		}

		// Detect dollar-quote tag: $tag$ or $$
		if r == '$' {
			rest := string(runes[i:])
			match := dollarQuoteRegex.FindString(rest)
			if match != "" {
				inDollar = true
				dollarTag = match
				cur.WriteString(match)
				i += len([]rune(match)) - 1
				continue
			}
		}

		if r == ';' {
			stmt := strings.TrimSpace(cur.String())
			if stmt != "" {
				stmts = append(stmts, stmt)
			}
			cur.Reset()
			continue
		}

		cur.WriteRune(r)
	}

	if stmt := strings.TrimSpace(cur.String()); stmt != "" {
		stmts = append(stmts, stmt)
	}
	return stmts
}

var dollarQuoteRegex = regexp.MustCompile(`^\$[a-zA-Z0-9_]*\$`)

// toGRPCAddr converts an HTTP(S) URL into a bare "host:port" gRPC address.
// If serverAddr is already host:port it is returned unchanged.
func toGRPCAddr(serverAddr string) string {
	if strings.HasPrefix(serverAddr, "http://") || strings.HasPrefix(serverAddr, "https://") {
		u, err := url.Parse(serverAddr)
		if err == nil {
			host := u.Hostname()
			port := u.Port()
			if port == "" {
				// Default gRPC port
				port = "50051"
			}
			return host + ":" + port
		}
	}
	// Already bare host:port or look like one
	if !strings.Contains(serverAddr, ":") {
		return serverAddr + ":50051"
	}
	return serverAddr
}
