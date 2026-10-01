package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/ssh"
)

// ProgressReader wraps an io.Reader to track transfer progress in real-time.
type ProgressReader struct {
	reader     io.Reader
	totalBytes int64
	readBytes  int64
	startTime  time.Time
	lastPrint  time.Time
	done       bool
	mu         sync.Mutex
}

func NewProgressReader(r io.Reader, total int64) *ProgressReader {
	return &ProgressReader{
		reader:     r,
		totalBytes: total,
		startTime:  time.Now(),
		lastPrint:  time.Now(),
	}
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		atomic.AddInt64(&pr.readBytes, int64(n))
		pr.printProgress(false)
	}
	if err == io.EOF {
		pr.printProgress(true)
	}
	return n, err
}

func (pr *ProgressReader) printProgress(force bool) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	now := time.Now()
	if !force && now.Sub(pr.lastPrint) < 500*time.Millisecond {
		return
	}
	pr.lastPrint = now

	current := atomic.LoadInt64(&pr.readBytes)
	elapsed := now.Sub(pr.startTime).Seconds()
	if elapsed <= 0 {
		elapsed = 0.001
	}

	percent := float64(current) / float64(pr.totalBytes) * 100
	if percent > 100 {
		percent = 100
	}

	speedMB := (float64(current) / (1024 * 1024)) / elapsed
	remainingBytes := pr.totalBytes - current
	var eta string
	if speedMB > 0 && remainingBytes > 0 {
		remSecs := float64(remainingBytes) / (speedMB * 1024 * 1024)
		eta = (time.Duration(remSecs) * time.Second).Round(time.Second).String()
	} else {
		eta = "0s"
	}

	barWidth := 30
	completed := int((percent / 100) * float64(barWidth))
	if completed > barWidth {
		completed = barWidth
	}
	bar := strings.Repeat("=", completed) + strings.Repeat(" ", barWidth-completed)

	fmt.Printf("\r[%s] %5.1f%% | %.2f GB / %.2f GB | Speed: %5.2f MB/s | ETA: %-8s",
		bar,
		percent,
		float64(current)/(1024*1024*1024),
		float64(pr.totalBytes)/(1024*1024*1024),
		speedMB,
		eta,
	)
	if force {
		fmt.Println()
	}
}

func isPortOpen(host, port string) bool {
	timeout := 1 * time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func StartSSHTunnel(sshUser, sshHost, keyPath, localPort, remoteHost, remotePort string) (net.Listener, *ssh.Client, error) {
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read SSH private key at %s: %w", keyPath, err)
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	sshAddr := sshHost
	if !strings.Contains(sshAddr, ":") {
		sshAddr += ":22"
	}

	fmt.Printf("[1/4] Connecting SSH tunnel to %s@%s...\n", sshUser, sshAddr)
	client, err := ssh.Dial("tcp", sshAddr, config)
	if err != nil {
		return nil, nil, fmt.Errorf("SSH connection failed: %w", err)
	}

	localListener, err := net.Listen("tcp", "127.0.0.1:"+localPort)
	if err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("failed to bind local port %s: %w", localPort, err)
	}

	remoteAddr := fmt.Sprintf("%s:%s", remoteHost, remotePort)
	go func() {
		for {
			localConn, err := localListener.Accept()
			if err != nil {
				return
			}
			go func(lConn net.Conn) {
				defer lConn.Close()
				rConn, err := client.Dial("tcp", remoteAddr)
				if err != nil {
					fmt.Printf("\n[Error] Failed to dial remote %s via SSH: %v\n", remoteAddr, err)
					return
				}
				defer rConn.Close()

				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					io.Copy(rConn, lConn)
				}()
				go func() {
					defer wg.Done()
					io.Copy(lConn, rConn)
				}()
				wg.Wait()
			}(localConn)
		}
	}()

	return localListener, client, nil
}

func findPsqlExecutable() (string, error) {
	if path, err := exec.LookPath("psql"); err == nil {
		return path, nil
	}

	commonPaths := []string{
		`C:\Program Files\PostgreSQL\18\bin\psql.exe`,
		`C:\Program Files\PostgreSQL\17\bin\psql.exe`,
		`C:\Program Files\PostgreSQL\16\bin\psql.exe`,
		`C:\Program Files\PostgreSQL\15\bin\psql.exe`,
		`C:\Program Files\PostgreSQL\14\bin\psql.exe`,
		`C:\Program Files\PostgreSQL\18\pgAdmin 4\runtime\psql.exe`,
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("psql.exe not found in PATH or standard PostgreSQL directories")
}

func main() {
	sshHost := flag.String("ssh-host", "18.221.228.140", "EC2 Bastion Host IP/Hostname")
	sshUser := flag.String("ssh-user", "ec2-user", "SSH Username")
	keyPath := flag.String("key", "", "Path to .pem private key file (optional if tunnel already running)")
	localPort := flag.String("port", "5433", "Local forwarded port (default 5433)")
	dbHost := flag.String("db-host", "127.0.0.1", "Remote DB Host from EC2 perspective")
	dbPort := flag.String("db-port", "5432", "Remote DB Port")
	dbUser := flag.String("db-user", "nembus_admin", "PostgreSQL DB User")
	dbPassword := flag.String("db-password", "", "PostgreSQL DB Password")
	dbName := flag.String("db-name", "sap", "Target Database Name (STRICT)")
	sqlFile := flag.String("file", "", "Path to .sql backup file")

	flag.Parse()

	fmt.Println("===============================================================")
	fmt.Println("       NEMBUS POSTGRESQL BACKUP RESTORE STREAMER (GO)         ")
	fmt.Println("===============================================================")

	// Strict Safety check: enforce sap database
	if *dbName != "sap" {
		fmt.Printf("[SECURITY CHECK] ERROR: Target DB is set to '%s'. This script is restricted to 'sap' only.\n", *dbName)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	// Check if local tunnel port is already active
	tunnelActive := isPortOpen("127.0.0.1", *localPort)
	if tunnelActive {
		fmt.Printf("[1/4] Detected active SSH tunnel on 127.0.0.1:%s! Using existing tunnel.\n", *localPort)
	} else {
		fmt.Printf("[1/4] Port 127.0.0.1:%s is not open yet.\n", *localPort)
		if *keyPath == "" {
			fmt.Println("Please run your SSH tunnel command in a PowerShell terminal:")
			fmt.Printf("   ssh -i \"path\\to\\key.pem\" -N -L %s:127.0.0.1:5432 %s@%s\n\n", *localPort, *sshUser, *sshHost)
			fmt.Print("Or enter path to .pem key to let this script open the tunnel automatically: ")
			input, _ := reader.ReadString('\n')
			*keyPath = strings.TrimSpace(input)
			*keyPath = strings.Trim(*keyPath, `"'`)
		}

		if *keyPath != "" {
			listener, client, err := StartSSHTunnel(*sshUser, *sshHost, *keyPath, *localPort, *dbHost, *dbPort)
			if err != nil {
				fmt.Printf("ERROR: Failed to establish SSH tunnel: %v\n", err)
				os.Exit(1)
			}
			defer listener.Close()
			defer client.Close()
			fmt.Printf("[OK] SSH Tunnel active: 127.0.0.1:%s -> %s:%s\n", *localPort, *dbHost, *dbPort)
		} else {
			// Wait for user to start external SSH tunnel
			fmt.Println("Waiting for tunnel to be opened on 127.0.0.1:" + *localPort + "...")
			for i := 0; i < 30; i++ {
				if isPortOpen("127.0.0.1", *localPort) {
					fmt.Printf("[OK] Tunnel detected on 127.0.0.1:%s!\n", *localPort)
					tunnelActive = true
					break
				}
				time.Sleep(1 * time.Second)
			}
			if !tunnelActive {
				fmt.Printf("ERROR: Timed out waiting for tunnel on 127.0.0.1:%s\n", *localPort)
				os.Exit(1)
			}
		}
	}

	// 2. Locate SQL file
	if *sqlFile == "" {
		fmt.Print("\nEnter full path to your .sql backup file: ")
		input, _ := reader.ReadString('\n')
		*sqlFile = strings.TrimSpace(input)
		*sqlFile = strings.Trim(*sqlFile, `"'`)
	}

	fileInfo, err := os.Stat(*sqlFile)
	if err != nil {
		fmt.Printf("ERROR: SQL backup file not found at: %s\n", *sqlFile)
		os.Exit(1)
	}

	// 3. DB password
	if *dbPassword == "" {
		fmt.Printf("Enter password for DB user '%s' on database '%s': ", *dbUser, *dbName)
		input, _ := reader.ReadString('\n')
		*dbPassword = strings.TrimSpace(input)
	}

	psqlPath, err := findPsqlExecutable()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[OK] PostgreSQL Engine: %s\n", psqlPath)

	// Prepare file
	f, err := os.Open(*sqlFile)
	if err != nil {
		fmt.Printf("ERROR: Failed to open backup file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	totalSize := fileInfo.Size()
	fmt.Printf("\n[2/4] Target Configuration:\n")
	fmt.Printf("      Backup File:   %s (%.2f GB)\n", filepath.Base(*sqlFile), float64(totalSize)/(1024*1024*1024))
	fmt.Printf("      Host & Port:   127.0.0.1:%s (SSH Tunnel)\n", *localPort)
	fmt.Printf("      Database:      %s (LOCKED: SAP ONLY)\n", *dbName)
	fmt.Printf("      User:          %s\n", *dbUser)
	fmt.Println("---------------------------------------------------------------")
	fmt.Print("Type 'yes' to start restoring: ")
	confirm, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(confirm)) != "yes" {
		fmt.Println("Restore aborted by user.")
		return
	}

	fmt.Println("\n[3/4] Streaming backup directly to PostgreSQL 'sap'...")

	cmd := exec.Command(psqlPath,
		"-h", "127.0.0.1",
		"-p", *localPort,
		"-U", *dbUser,
		"-d", *dbName,
		"-v", "ON_ERROR_STOP=0",
	)

	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PGPASSWORD=%s", *dbPassword),
		"PGCONNECT_TIMEOUT=15",
	)

	progReader := NewProgressReader(f, totalSize)
	cmd.Stdin = progReader

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	restoreStartTime := time.Now()
	err = cmd.Run()

	duration := time.Since(restoreStartTime)

	if err != nil {
		fmt.Printf("\n[!] Restore finished with warnings or error: %v\n", err)
	} else {
		fmt.Printf("\n[SUCCESS] Backup restored successfully to database '%s' in %s!\n", *dbName, duration.Round(time.Second))
	}
}
