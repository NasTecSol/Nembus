package telemetry

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

func TestHTTPTraceContextPropagation(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	defer func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
		_ = tp.Shutdown(context.Background())
	}()

	var serverTraceID string
	srv := httptest.NewServer(WrapHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverTraceID = w.Header().Get("X-Trace-ID")
		w.WriteHeader(http.StatusOK)
	}), "test-server"))
	defer srv.Close()

	client := WrapHTTPClient(&http.Client{})
	ctx := WithTenantSlug(context.Background(), "acme-store")
	ctx, rootSpan := Tracer("test").Start(ctx, "root-op")
	expectedTraceID := rootSpan.SpanContext().TraceID().String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/ping", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("http do: %v", err)
	}
	_ = resp.Body.Close()
	rootSpan.End()

	if serverTraceID != expectedTraceID {
		t.Fatalf("expected trace ID %s propagated to server, got %s", expectedTraceID, serverTraceID)
	}
	if got := resp.Header.Get("X-Trace-ID"); got != expectedTraceID {
		t.Fatalf("expected X-Trace-ID header %s, got %s", expectedTraceID, got)
	}
}

func TestGRPCStatsHandlerPropagation(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	defer func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
		_ = tp.Shutdown(context.Background())
	}()

	clientHandler := GRPCClientStatsHandler()
	serverHandler := GRPCServerStatsHandler()

	clientCtx := clientHandler.TagRPC(WithTenantSlug(context.Background(), "retail-1"), &stats.RPCTagInfo{
		FullMethodName: "/syncpb.SyncService/StreamPush",
	})
	md, ok := metadata.FromOutgoingContext(clientCtx)
	if !ok || len(md.Get("traceparent")) == 0 {
		t.Fatalf("expected traceparent in outgoing gRPC metadata, got %v", md)
	}
	md.Set("x-tenant-id", "retail-1")

	incomingCtx := metadata.NewIncomingContext(context.Background(), md)
	serverCtx := serverHandler.TagRPC(incomingCtx, &stats.RPCTagInfo{
		FullMethodName: "/syncpb.SyncService/StreamPush",
	})

	if gotTenant := TenantSlugFromContext(serverCtx); gotTenant != "retail-1" {
		t.Fatalf("expected tenant retail-1 on server context, got %q", gotTenant)
	}

	serverHandler.HandleRPC(serverCtx, &stats.End{Error: status.Error(codes.Internal, "boom")})
	clientHandler.HandleRPC(clientCtx, &stats.End{})

	spans := exporter.GetSpans()
	if len(spans) != 2 {
		t.Fatalf("expected 2 gRPC spans, got %d", len(spans))
	}
	if spans[0].SpanContext.TraceID() != spans[1].SpanContext.TraceID() {
		t.Fatalf("client and server spans do not share TraceID: %s vs %s",
			spans[0].SpanContext.TraceID(), spans[1].SpanContext.TraceID())
	}
}

func TestDiskSpoolRoundTripperBuffersAndFlushes(t *testing.T) {
	var failMode atomic.Bool
	failMode.Store(true)
	var receivedCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Otel-Token") != "secret-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if failMode.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = io.ReadAll(r.Body)
		receivedCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	spoolRoot := t.TempDir()
	rt, err := NewDiskSpoolRoundTripper(spoolRoot, 1024, "secret-token", http.DefaultTransport)
	if err != nil {
		t.Fatalf("create spool round tripper: %v", err)
	}
	defer func() { _ = rt.Close(context.Background()) }()

	// First request fails -> spooled locally and returns 202 Accepted.
	req1, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/traces", strings.NewReader(strings.Repeat("a", 400)))
	req1.Header.Set("Content-Type", "application/x-protobuf")
	resp1, err := rt.RoundTrip(req1)
	if err != nil {
		t.Fatalf("unexpected error during offline spool: %v", err)
	}
	_ = resp1.Body.Close()
	if resp1.StatusCode != http.StatusAccepted || resp1.Header.Get("X-Otel-Spooled") != "true" {
		t.Fatalf("expected 202 with X-Otel-Spooled=true, got %d", resp1.StatusCode)
	}

	// Write two more ~400B entries -> maxBytes=1024 forces FIFO eviction of oldest file.
	for i := 0; i < 2; i++ {
		reqN, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/logs", strings.NewReader(strings.Repeat("b", 400)))
		respN, err := rt.RoundTrip(reqN)
		if err != nil {
			t.Fatalf("round trip %d: %v", i, err)
		}
		_ = respN.Body.Close()
	}

	files, err := os.ReadDir(filepath.Join(spoolRoot, "otel-spool"))
	if err != nil {
		t.Fatalf("readdir spool: %v", err)
	}
	if len(files) == 0 || len(files) > 2 {
		t.Fatalf("expected FIFO eviction to keep 1..2 files under 1024B cap, got %d", len(files))
	}

	// Recover server and flush.
	failMode.Store(false)
	if err := rt.Flush(context.Background()); err != nil {
		t.Fatalf("flush failed: %v", err)
	}
	if receivedCount.Load() == 0 {
		t.Fatalf("expected spooled payloads to be replayed after recovery")
	}
	filesAfter, _ := os.ReadDir(filepath.Join(spoolRoot, "otel-spool"))
	if len(filesAfter) != 0 {
		t.Fatalf("expected empty spool after flush, found %d files", len(filesAfter))
	}
}

func TestPGXTracerRecordsSpan(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	prevTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() {
		otel.SetTracerProvider(prevTP)
		_ = tp.Shutdown(context.Background())
	}()

	tracer := NewPGXTracer("tenant:demo")
	ctx := tracer.TraceQueryStart(context.Background(), nil, pgx.TraceQueryStartData{
		SQL: "/* comment */ SELECT id FROM products WHERE id = $1",
	})
	tracer.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{
		Err: errors.New("connection lost"),
	})

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 pgx span, got %d", len(spans))
	}
	if spans[0].Name != "postgresql.SELECT" {
		t.Fatalf("expected span name postgresql.SELECT, got %s", spans[0].Name)
	}
}

func TestApplyAuthHeadersFormats(t *testing.T) {
	wantBasic := "Basic " + base64.StdEncoding.EncodeToString([]byte("123456:glc_secret_key"))
	cases := []struct {
		name      string
		token     string
		wantAuth  string
		wantXTok  string
	}{
		{name: "empty", token: "", wantAuth: "", wantXTok: ""},
		{name: "instance_and_key", token: "123456:glc_secret_key", wantAuth: wantBasic, wantXTok: ""},
		{name: "otel_headers_basic", token: "Authorization=Basic%20abc123", wantAuth: "Basic abc123", wantXTok: ""},
		{name: "bearer_prefixed", token: "Bearer my-tok", wantAuth: "Bearer my-tok", wantXTok: "my-tok"},
		{name: "bare_ingest_token", token: "nginx-tok", wantAuth: "Bearer nginx-tok", wantXTok: "nginx-tok"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rt := &DiskSpoolRoundTripper{ingestToken: tc.token}
			req, _ := http.NewRequest(http.MethodPost, "https://example.com/v1/traces", nil)
			rt.applyAuthHeaders(req)
			if got := req.Header.Get("Authorization"); got != tc.wantAuth {
				t.Fatalf("Authorization mismatch: want %q, got %q", tc.wantAuth, got)
			}
			if got := req.Header.Get("X-Otel-Token"); got != tc.wantXTok {
				t.Fatalf("X-Otel-Token mismatch: want %q, got %q", tc.wantXTok, got)
			}
		})
	}
}

func TestResolveOTLPURLsGrafanaCloud(t *testing.T) {
	tr, mt, lg, err := resolveOTLPURLs("https://otlp-gateway-prod-us-east-0.grafana.net")
	if err != nil {
		t.Fatalf("resolve grafana cloud root url: %v", err)
	}
	if tr != "https://otlp-gateway-prod-us-east-0.grafana.net/otlp/v1/traces" ||
		mt != "https://otlp-gateway-prod-us-east-0.grafana.net/otlp/v1/metrics" ||
		lg != "https://otlp-gateway-prod-us-east-0.grafana.net/otlp/v1/logs" {
		t.Fatalf("unexpected grafana cloud URLs: %s | %s | %s", tr, mt, lg)
	}

	tr2, _, _, err := resolveOTLPURLs("https://otlp-gateway-prod-us-east-0.grafana.net/otlp")
	if err != nil {
		t.Fatalf("resolve grafana cloud /otlp url: %v", err)
	}
	if tr2 != "https://otlp-gateway-prod-us-east-0.grafana.net/otlp/v1/traces" {
		t.Fatalf("unexpected explicit /otlp trace URL: %s", tr2)
	}

	tr3, _, _, err := resolveOTLPURLs("127.0.0.1:4318")
	if err != nil {
		t.Fatalf("resolve local alloy url: %v", err)
	}
	if tr3 != "http://127.0.0.1:4318/v1/traces" {
		t.Fatalf("unexpected local trace URL: %s", tr3)
	}
}
