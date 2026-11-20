package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	scopeName                   = "github.com/NasTecSol/nembus-core/telemetry"
	defaultMetricExportInterval = 60 * time.Second
)

// Config controls OpenTelemetry initialization for any Nembus application.
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	TenantSlug     string
	InstanceID     string
	OTLPEndpoint   string
	IngestToken    string
	SpoolDir       string
	MaxSpoolBytes  int64
}

type tenantCtxKey struct{}

var (
	dynamicTenantSlug atomic.Value
	errorHandlerOnce  sync.Once
)

func init() {
	dynamicTenantSlug.Store("")
}

// WithTenantSlug attaches a tenant slug to ctx for tracing, SQL spans, and logging.
func WithTenantSlug(ctx context.Context, slug string) context.Context {
	slug = strings.TrimSpace(slug)
	if slug == "" || ctx == nil {
		return ctx
	}
	return context.WithValue(ctx, tenantCtxKey{}, slug)
}

// TenantSlugFromContext extracts the tenant slug from ctx, falling back to the
// process-level dynamic tenant slug (used by single-tenant edge agents).
func TenantSlugFromContext(ctx context.Context) string {
	if ctx != nil {
		if v, ok := ctx.Value(tenantCtxKey{}).(string); ok && v != "" {
			return v
		}
	}
	return DynamicTenantSlug()
}

// SetDynamicTenantSlug updates the process-wide tenant slug attribute used by
// edge apps (`pos-client`, `sap-agent`) once onboarding finishes.
func SetDynamicTenantSlug(slug string) {
	dynamicTenantSlug.Store(strings.TrimSpace(slug))
}

// DynamicTenantSlug returns the process-wide tenant slug, if set.
func DynamicTenantSlug() string {
	v, _ := dynamicTenantSlug.Load().(string)
	return v
}

// Tracer returns a named OpenTelemetry tracer from the global TracerProvider.
func Tracer(name string) trace.Tracer {
	if name == "" {
		name = scopeName
	}
	return otel.Tracer(name)
}

// Meter returns a named OpenTelemetry meter from the global MeterProvider.
func Meter(name string) metric.Meter {
	if name == "" {
		name = scopeName
	}
	return otel.Meter(name)
}

// Init configures W3C propagation and, when an OTLP endpoint is configured,
// initializes OTLP HTTP tracing, metrics (including Go runtime metrics), and logs
// backed by a bounded disk spool for offline edge resilience.
func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	if cfg.ServiceName == "" {
		cfg.ServiceName = strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME"))
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "nembus-service"
	}
	if cfg.Environment == "" {
		cfg.Environment = strings.TrimSpace(os.Getenv("ENV"))
	}
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}
	if cfg.ServiceVersion == "" {
		cfg.ServiceVersion = "dev"
	}
	if cfg.InstanceID == "" {
		if host, err := os.Hostname(); err == nil && host != "" {
			cfg.InstanceID = host
		}
	}
	if cfg.TenantSlug != "" {
		SetDynamicTenantSlug(cfg.TenantSlug)
	}
	if cfg.OTLPEndpoint == "" {
		cfg.OTLPEndpoint = strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	}
	if cfg.OTLPEndpoint == "" {
		cfg.OTLPEndpoint = strings.TrimSpace(os.Getenv("GRAFANA_CLOUD_OTLP_ENDPOINT"))
	}
	if cfg.IngestToken == "" {
		cfg.IngestToken = strings.TrimSpace(os.Getenv("OTEL_INGEST_TOKEN"))
	}
	if cfg.IngestToken == "" {
		cfg.IngestToken = strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"))
	}
	if cfg.IngestToken == "" {
		instanceID := strings.TrimSpace(os.Getenv("GRAFANA_CLOUD_INSTANCE_ID"))
		apiKey := strings.TrimSpace(os.Getenv("GRAFANA_CLOUD_API_KEY"))
		if instanceID != "" && apiKey != "" {
			cfg.IngestToken = instanceID + ":" + apiKey
		}
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Never route OTel internal SDK errors through stdlib log.Printf, because
	// stdlib log output is forwarded to the OTLP log bridge.
	errorHandlerOnce.Do(func() {
		otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
			if err != nil {
				fmt.Fprintf(os.Stderr, "[telemetry] sdk error: %v\n", err)
			}
		}))
	})

	if strings.TrimSpace(cfg.OTLPEndpoint) == "" {
		refreshInstruments()
		return func(context.Context) error { return nil }, nil
	}

	tracesURL, metricsURL, logsURL, err := resolveOTLPURLs(cfg.OTLPEndpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid OTLP endpoint %q: %w", cfg.OTLPEndpoint, err)
	}

	spoolRT, err := NewDiskSpoolRoundTripper(cfg.SpoolDir, cfg.MaxSpoolBytes, cfg.IngestToken, nil)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: spoolRT,
	}

	attrs := []attribute.KeyValue{
		attribute.String("service.name", cfg.ServiceName),
		attribute.String("service.namespace", "nembus"),
		attribute.String("service.version", cfg.ServiceVersion),
		attribute.String("deployment.environment", cfg.Environment),
		attribute.String("deployment.environment.name", cfg.Environment),
	}
	if cfg.InstanceID != "" {
		attrs = append(attrs, attribute.String("service.instance.id", cfg.InstanceID))
	}
	if cfg.TenantSlug != "" {
		attrs = append(attrs, attribute.String("nembus.tenant.slug", cfg.TenantSlug))
	}
	res := resource.NewWithAttributes("", attrs...)

	metricInterval := resolveMetricExportInterval()

	traceExp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(tracesURL),
		otlptracehttp.WithHTTPClient(httpClient),
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
		otlptracehttp.WithRetry(otlptracehttp.RetryConfig{Enabled: false}),
	)
	if err != nil {
		_ = spoolRT.Close(ctx)
		return nil, fmt.Errorf("init OTLP trace exporter: %w", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	metricExp, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpointURL(metricsURL),
		otlpmetrichttp.WithHTTPClient(httpClient),
		otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
		otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{Enabled: false}),
	)
	if err != nil {
		_ = tp.Shutdown(ctx)
		_ = spoolRT.Close(ctx)
		return nil, fmt.Errorf("init OTLP metric exporter: %w", err)
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp, sdkmetric.WithInterval(metricInterval))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)
	refreshInstruments()

	if rtErr := runtime.Start(
		runtime.WithMeterProvider(mp),
		runtime.WithMinimumReadMemStatsInterval(metricInterval),
	); rtErr != nil {
		fmt.Fprintf(os.Stderr, "[telemetry] warning: runtime metrics failed to start: %v\n", rtErr)
	}

	logExp, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(logsURL),
		otlploghttp.WithHTTPClient(httpClient),
		otlploghttp.WithCompression(otlploghttp.GzipCompression),
		otlploghttp.WithRetry(otlploghttp.RetryConfig{Enabled: false}),
	)
	if err != nil {
		_ = mp.Shutdown(ctx)
		_ = tp.Shutdown(ctx)
		_ = spoolRT.Close(ctx)
		return nil, fmt.Errorf("init OTLP log exporter: %w", err)
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
		sdklog.WithResource(res),
	)
	global.SetLoggerProvider(lp)

	restoreLog := installLogBridge(cfg.ServiceName, cfg.Environment, lp)

	var shutdownOnce sync.Once
	shutdown := func(shutdownCtx context.Context) error {
		var firstErr error
		shutdownOnce.Do(func() {
			if restoreLog != nil {
				restoreLog()
			}
			for _, fn := range []func(context.Context) error{
				lp.Shutdown,
				mp.Shutdown,
				tp.Shutdown,
				spoolRT.Close,
			} {
				if err := fn(shutdownCtx); err != nil && firstErr == nil {
					firstErr = err
				}
			}
		})
		return firstErr
	}
	return shutdown, nil
}

func resolveOTLPURLs(rawEndpoint string) (tracesURL, metricsURL, logsURL string, err error) {
	ep := strings.TrimSpace(rawEndpoint)
	if !strings.Contains(ep, "://") {
		ep = "http://" + ep
	}
	u, err := url.Parse(ep)
	if err != nil || u.Host == "" {
		return "", "", "", fmt.Errorf("missing host in %q", rawEndpoint)
	}
	basePath := strings.TrimRight(u.Path, "/")
	if basePath == "" && strings.HasSuffix(strings.ToLower(u.Hostname()), ".grafana.net") {
		basePath = "/otlp"
	}
	makeURL := func(suffix string) string {
		copyU := *u
		copyU.Path = basePath + suffix
		return copyU.String()
	}
	return makeURL("/v1/traces"), makeURL("/v1/metrics"), makeURL("/v1/logs"), nil
}

func resolveMetricExportInterval() time.Duration {
	raw := strings.TrimSpace(os.Getenv("OTEL_METRIC_EXPORT_INTERVAL"))
	if raw == "" {
		return defaultMetricExportInterval
	}
	if d, err := time.ParseDuration(raw); err == nil && d >= 5*time.Second {
		return d
	}
	if secs, err := strconv.Atoi(raw); err == nil && secs >= 5 {
		return time.Duration(secs) * time.Second
	}
	return defaultMetricExportInterval
}

// WrapHTTPClient returns a shallow copy of client (or a new client if nil)
// whose Transport creates client spans and injects W3C traceparent headers.
func WrapHTTPClient(client *http.Client) *http.Client {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	cloned := *client
	base := cloned.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	if _, already := base.(*otelRoundTripper); !already {
		cloned.Transport = &otelRoundTripper{base: base}
	}
	return &cloned
}

type otelRoundTripper struct {
	base http.RoundTripper
}

func (rt *otelRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, span := Tracer("nembus/http-client").Start(
		req.Context(),
		fmt.Sprintf("HTTP %s", req.Method),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("http.request.method", req.Method),
			attribute.String("url.full", req.URL.Redacted()),
		),
	)
	defer span.End()

	if tenant := TenantSlugFromContext(ctx); tenant != "" {
		span.SetAttributes(attribute.String("nembus.tenant.slug", tenant))
	}

	outReq := req.Clone(ctx)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(outReq.Header))

	resp, err := rt.base.RoundTrip(outReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetAttributes(attribute.Int("http.response.status_code", resp.StatusCode))
	if resp.StatusCode >= 500 {
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode))
	}
	return resp, nil
}

type statusCaptureWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusCaptureWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// WrapHTTPHandler wraps a standard net/http handler to extract W3C trace
// headers, record server spans, populate X-Trace-ID, and emit HTTP metrics.
func WrapHTTPHandler(next http.Handler, serviceName string) http.Handler {
	if next == nil {
		return nil
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		route := r.URL.Path

		ctx, span := Tracer("nembus/http-server").Start(
			ctx,
			fmt.Sprintf("%s %s", r.Method, route),
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.request.method", r.Method),
				attribute.String("url.path", route),
			),
		)
		defer span.End()

		if sc := span.SpanContext(); sc.HasTraceID() {
			w.Header().Set("X-Trace-ID", sc.TraceID().String())
		}

		sw := &statusCaptureWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r.WithContext(ctx))

		durationMs := float64(time.Since(start).Microseconds()) / 1000.0
		tenant := TenantSlugFromContext(ctx)
		span.SetAttributes(attribute.Int("http.response.status_code", sw.status))
		if tenant != "" {
			span.SetAttributes(attribute.String("nembus.tenant.slug", tenant))
		}
		if sw.status >= 500 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", sw.status))
		}
		RecordHTTPRequest(ctx, serviceName, r.Method, route, tenant, sw.status, durationMs)
	})
}
