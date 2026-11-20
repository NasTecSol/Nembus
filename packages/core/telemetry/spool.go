package telemetry

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultMaxSpoolBytes int64 = 50 * 1024 * 1024 // 50 MB
	defaultFlushInterval       = 30 * time.Second
	maxSpoolHeaderBytes        = 64 * 1024
)

type spoolEnvelope struct {
	Method          string `json:"method"`
	URL             string `json:"url"`
	ContentType     string `json:"content_type,omitempty"`
	ContentEncoding string `json:"content_encoding,omitempty"`
}

// DiskSpoolRoundTripper wraps an http.RoundTripper to inject authentication
// headers and buffer failed OTLP HTTP payloads on disk up to a bounded size.
type DiskSpoolRoundTripper struct {
	base        http.RoundTripper
	spoolDir    string
	maxBytes    int64
	ingestToken string

	mu       sync.Mutex
	seq      atomic.Uint64
	stopCh   chan struct{}
	doneCh   chan struct{}
	stopOnce sync.Once
}

// NewDiskSpoolRoundTripper creates a DiskSpoolRoundTripper. If spoolDir is empty,
// disk buffering is disabled and only auth header injection is performed.
func NewDiskSpoolRoundTripper(spoolDir string, maxBytes int64, ingestToken string, base http.RoundTripper) (*DiskSpoolRoundTripper, error) {
	if base == nil {
		base = &http.Transport{
			MaxIdleConns:        16,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		}
	}
	if maxBytes <= 0 {
		maxBytes = defaultMaxSpoolBytes
	}

	var resolvedDir string
	if strings.TrimSpace(spoolDir) != "" {
		resolvedDir = filepath.Join(spoolDir, "otel-spool")
		if err := os.MkdirAll(resolvedDir, 0o755); err != nil {
			return nil, fmt.Errorf("create otel spool directory %s: %w", resolvedDir, err)
		}
	}

	rt := &DiskSpoolRoundTripper{
		base:        base,
		spoolDir:    resolvedDir,
		maxBytes:    maxBytes,
		ingestToken: strings.TrimSpace(ingestToken),
		stopCh:      make(chan struct{}),
		doneCh:      make(chan struct{}),
	}

	if resolvedDir != "" {
		go rt.flushLoop()
	} else {
		close(rt.doneCh)
	}

	return rt, nil
}

// RoundTrip sends the OTLP HTTP request, spooling the payload to disk if the
// collector is unreachable or returns a transient error (5xx / 429).
func (rt *DiskSpoolRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err != nil {
			return nil, err
		}
	}

	cloned := req.Clone(req.Context())
	if bodyBytes != nil {
		cloned.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		cloned.ContentLength = int64(len(bodyBytes))
	}
	rt.applyAuthHeaders(cloned)

	resp, err := rt.base.RoundTrip(cloned)
	if err == nil && resp != nil && !shouldSpoolStatus(resp.StatusCode) {
		return resp, nil
	}

	if rt.spoolDir == "" || req.Method != http.MethodPost || len(bodyBytes) == 0 {
		return resp, err
	}

	if resp != nil && resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}

	if spoolErr := rt.enqueue(req, bodyBytes); spoolErr != nil {
		fmt.Fprintf(os.Stderr, "[telemetry] failed to spool OTLP payload: %v\n", spoolErr)
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Status:     "503 Service Unavailable",
			Body:       io.NopCloser(strings.NewReader("")),
			Request:    req,
		}, nil
	}

	// Return 202 Accepted once durably buffered locally so the OTel SDK stops
	// retrying synchronously and avoids stalling offline terminals.
	header := make(http.Header)
	header.Set("X-Otel-Spooled", "true")
	return &http.Response{
		StatusCode: http.StatusAccepted,
		Status:     "202 Accepted",
		Header:     header,
		Body:       io.NopCloser(strings.NewReader("")),
		Request:    req,
	}, nil
}

// Flush replays all spooled OTLP payloads currently waiting on disk.
func (rt *DiskSpoolRoundTripper) Flush(ctx context.Context) error {
	if rt.spoolDir == "" {
		return nil
	}

	rt.mu.Lock()
	entries, err := rt.listSpoolFilesLocked()
	rt.mu.Unlock()
	if err != nil {
		return err
	}

	for _, path := range entries {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		env, body, readErr := readSpoolFile(path)
		if readErr != nil {
			_ = os.Remove(path)
			continue
		}

		reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		req, reqErr := http.NewRequestWithContext(reqCtx, env.Method, env.URL, bytes.NewReader(body))
		if reqErr != nil {
			cancel()
			_ = os.Remove(path)
			continue
		}
		if env.ContentType != "" {
			req.Header.Set("Content-Type", env.ContentType)
		}
		if env.ContentEncoding != "" {
			req.Header.Set("Content-Encoding", env.ContentEncoding)
		}
		rt.applyAuthHeaders(req)

		resp, sendErr := rt.base.RoundTrip(req)
		cancel()
		if sendErr != nil {
			// Stop replaying on connectivity failure; retry on next tick.
			return sendErr
		}
		if resp.Body != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			_ = os.Remove(path)
			continue
		}
		if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusRequestEntityTooLarge {
			// Permanent payload error; drop to avoid head-of-line blocking.
			_ = os.Remove(path)
			continue
		}
		if shouldSpoolStatus(resp.StatusCode) {
			return fmt.Errorf("collector returned status %d during spool flush", resp.StatusCode)
		}
		_ = os.Remove(path)
	}
	return nil
}

// Close stops the background flush loop and attempts a best-effort flush.
func (rt *DiskSpoolRoundTripper) Close(ctx context.Context) error {
	rt.stopOnce.Do(func() {
		close(rt.stopCh)
	})
	<-rt.doneCh
	return rt.Flush(ctx)
}

func (rt *DiskSpoolRoundTripper) flushLoop() {
	defer close(rt.doneCh)
	ticker := time.NewTicker(defaultFlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rt.stopCh:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			_ = rt.Flush(ctx)
			cancel()
		}
	}
}

func (rt *DiskSpoolRoundTripper) applyAuthHeaders(req *http.Request) {
	token := strings.TrimSpace(rt.ingestToken)
	if token == "" {
		return
	}
	lower := strings.ToLower(token)
	switch {
	case strings.HasPrefix(lower, "authorization="):
		val := strings.TrimSpace(token[len("authorization="):])
		if idx := strings.IndexByte(val, ','); idx >= 0 {
			val = strings.TrimSpace(val[:idx])
		}
		val = strings.ReplaceAll(val, "%20", " ")
		if val != "" {
			req.Header.Set("Authorization", val)
		}
	case strings.HasPrefix(lower, "basic "):
		req.Header.Set("Authorization", "Basic "+strings.TrimSpace(token[6:]))
	case strings.HasPrefix(lower, "bearer "):
		bare := strings.TrimSpace(token[7:])
		req.Header.Set("Authorization", "Bearer "+bare)
		if bare != "" {
			req.Header.Set("X-Otel-Token", bare)
		}
	case strings.Contains(token, ":"):
		encoded := base64.StdEncoding.EncodeToString([]byte(token))
		req.Header.Set("Authorization", "Basic "+encoded)
	default:
		req.Header.Set("X-Otel-Token", token)
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

func shouldSpoolStatus(code int) bool {
	return code == http.StatusTooManyRequests || code >= 500
}

func (rt *DiskSpoolRoundTripper) enqueue(req *http.Request, body []byte) error {
	env := spoolEnvelope{
		Method:          req.Method,
		URL:             req.URL.String(),
		ContentType:     req.Header.Get("Content-Type"),
		ContentEncoding: req.Header.Get("Content-Encoding"),
	}
	headerBytes, err := json.Marshal(env)
	if err != nil {
		return err
	}
	payloadSize := int64(4 + len(headerBytes) + len(body))
	if payloadSize > rt.maxBytes {
		return fmt.Errorf("payload size %d exceeds max spool limit %d", payloadSize, rt.maxBytes)
	}

	rt.mu.Lock()
	defer rt.mu.Unlock()

	if err := rt.evictForSpaceLocked(payloadSize); err != nil {
		return err
	}

	seq := rt.seq.Add(1)
	name := fmt.Sprintf("%020d_%08d.bin", time.Now().UTC().UnixNano(), seq)
	path := filepath.Join(rt.spoolDir, name)
	tmpPath := path + ".tmp"

	var buf bytes.Buffer
	buf.Grow(int(payloadSize))
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(headerBytes)))
	buf.Write(lenBuf[:])
	buf.Write(headerBytes)
	buf.Write(body)

	if err := os.WriteFile(tmpPath, buf.Bytes(), 0o600); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func (rt *DiskSpoolRoundTripper) evictForSpaceLocked(incoming int64) error {
	entries, err := os.ReadDir(rt.spoolDir)
	if err != nil {
		return err
	}

	type fileInfo struct {
		path string
		name string
		size int64
	}
	var files []fileInfo
	var total int64

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".bin") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, fileInfo{
			path: filepath.Join(rt.spoolDir, entry.Name()),
			name: entry.Name(),
			size: info.Size(),
		})
		total += info.Size()
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].name < files[j].name
	})

	for _, f := range files {
		if total+incoming <= rt.maxBytes {
			break
		}
		if removeErr := os.Remove(f.path); removeErr == nil {
			total -= f.size
		}
	}
	return nil
}

func (rt *DiskSpoolRoundTripper) listSpoolFilesLocked() ([]string, error) {
	entries, err := os.ReadDir(rt.spoolDir)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".bin") {
			continue
		}
		paths = append(paths, filepath.Join(rt.spoolDir, entry.Name()))
	}
	sort.Strings(paths)
	return paths, nil
}

func readSpoolFile(path string) (spoolEnvelope, []byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return spoolEnvelope{}, nil, err
	}
	if len(data) < 4 {
		return spoolEnvelope{}, nil, fmt.Errorf("corrupt spool file %s: too short", path)
	}
	headerLen := binary.BigEndian.Uint32(data[:4])
	if headerLen == 0 || headerLen > maxSpoolHeaderBytes || int(4+headerLen) > len(data) {
		return spoolEnvelope{}, nil, fmt.Errorf("corrupt spool header in %s", path)
	}
	var env spoolEnvelope
	if err := json.Unmarshal(data[4:4+headerLen], &env); err != nil {
		return spoolEnvelope{}, nil, err
	}
	return env, data[4+headerLen:], nil
}
