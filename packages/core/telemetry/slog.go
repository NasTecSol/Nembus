package telemetry

import (
	"bytes"
	"context"
	"io"
	stdlog "log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
)

func installLogBridge(serviceName, environment string, lp *sdklog.LoggerProvider) func() {
	if lp == nil {
		return nil
	}

	origStdWriter := stdlog.Writer()
	origSlog := slog.Default()

	stdlibLogger := lp.Logger("nembus/stdlib")
	slogOTelLogger := lp.Logger("nembus/slog")

	stdTee := &otlpStdLogWriter{
		passthrough: origStdWriter,
		logger:      stdlibLogger,
		serviceName: serviceName,
		environment: environment,
	}
	stdlog.SetOutput(stdTee)

	jsonOut := slog.NewJSONHandler(origStdWriter, &slog.HandlerOptions{Level: slog.LevelInfo})
	multi := &multiSlogHandler{
		stdout:      jsonOut,
		otelLogger:  slogOTelLogger,
		serviceName: serviceName,
		environment: environment,
	}
	slog.SetDefault(slog.New(multi))

	return func() {
		stdlog.SetOutput(origStdWriter)
		slog.SetDefault(origSlog)
	}
}

// LogInfo emits a structured INFO log record correlated with any active span in ctx.
func LogInfo(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

// LogWarn emits a structured WARN log record correlated with any active span in ctx.
func LogWarn(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
}

// LogError emits a structured ERROR log record correlated with any active span in ctx.
func LogError(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}
	slog.LogAttrs(ctx, slog.LevelError, msg, attrs...)
}

type otlpStdLogWriter struct {
	passthrough io.Writer
	logger      otellog.Logger
	serviceName string
	environment string
	mu          sync.Mutex
}

func (w *otlpStdLogWriter) Write(p []byte) (int, error) {
	n, err := w.passthrough.Write(p)

	msg := strings.TrimSpace(string(bytes.TrimRight(p, "\r\n")))
	if msg != "" && w.logger != nil {
		var rec otellog.Record
		now := time.Now()
		rec.SetTimestamp(now)
		rec.SetObservedTimestamp(now)
		rec.SetBody(otellog.StringValue(msg))
		sev, sevText := classifyLogSeverity(msg)
		rec.SetSeverity(sev)
		rec.SetSeverityText(sevText)

		attrs := []otellog.KeyValue{
			otellog.String("service.name", w.serviceName),
			otellog.String("deployment.environment", w.environment),
		}
		if tenant := DynamicTenantSlug(); tenant != "" {
			attrs = append(attrs, otellog.String("nembus.tenant.slug", tenant))
		}
		rec.AddAttributes(attrs...)

		// Serialize only the lightweight emit call; never log to stdlib from here.
		w.mu.Lock()
		w.logger.Emit(context.Background(), rec)
		w.mu.Unlock()
	}

	return n, err
}

func classifyLogSeverity(msg string) (otellog.Severity, string) {
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "fatal") || strings.Contains(lower, "[error]") || strings.Contains(lower, "error:") || strings.Contains(lower, "failed"):
		return otellog.SeverityError, "ERROR"
	case strings.Contains(lower, "warning") || strings.Contains(lower, "[warn]"):
		return otellog.SeverityWarn, "WARN"
	default:
		return otellog.SeverityInfo, "INFO"
	}
}

type multiSlogHandler struct {
	stdout      slog.Handler
	otelLogger  otellog.Logger
	serviceName string
	environment string
	attrs       []slog.Attr
	groups      []string
}

func (h *multiSlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.stdout.Enabled(ctx, level)
}

func (h *multiSlogHandler) Handle(ctx context.Context, record slog.Record) error {
	enriched := record.Clone()
	if tenant := TenantSlugFromContext(ctx); tenant != "" {
		enriched.AddAttrs(slog.String("nembus.tenant.slug", tenant))
	}
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		enriched.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}

	_ = h.stdout.Handle(ctx, enriched)

	if h.otelLogger != nil {
		var rec otellog.Record
		rec.SetTimestamp(record.Time)
		rec.SetObservedTimestamp(time.Now())
		rec.SetBody(otellog.StringValue(record.Message))
		sev, sevText := slogLevelToOTel(record.Level)
		rec.SetSeverity(sev)
		rec.SetSeverityText(sevText)

		kvs := []otellog.KeyValue{
			otellog.String("service.name", h.serviceName),
			otellog.String("deployment.environment", h.environment),
		}
		if tenant := TenantSlugFromContext(ctx); tenant != "" {
			kvs = append(kvs, otellog.String("nembus.tenant.slug", tenant))
		}
		if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
			kvs = append(kvs,
				otellog.String("trace_id", sc.TraceID().String()),
				otellog.String("span_id", sc.SpanID().String()),
			)
		}
		for _, a := range h.attrs {
			kvs = append(kvs, slogAttrToOTel(h.groups, a))
		}
		record.Attrs(func(a slog.Attr) bool {
			kvs = append(kvs, slogAttrToOTel(h.groups, a))
			return true
		})
		rec.AddAttributes(kvs...)
		h.otelLogger.Emit(ctx, rec)
	}
	return nil
}

func (h *multiSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	merged := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	merged = append(merged, h.attrs...)
	merged = append(merged, attrs...)
	return &multiSlogHandler{
		stdout:      h.stdout.WithAttrs(attrs),
		otelLogger:  h.otelLogger,
		serviceName: h.serviceName,
		environment: h.environment,
		attrs:       merged,
		groups:      h.groups,
	}
}

func (h *multiSlogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	groups := append(append([]string(nil), h.groups...), name)
	return &multiSlogHandler{
		stdout:      h.stdout.WithGroup(name),
		otelLogger:  h.otelLogger,
		serviceName: h.serviceName,
		environment: h.environment,
		attrs:       h.attrs,
		groups:      groups,
	}
}

func slogLevelToOTel(lvl slog.Level) (otellog.Severity, string) {
	switch {
	case lvl >= slog.LevelError:
		return otellog.SeverityError, "ERROR"
	case lvl >= slog.LevelWarn:
		return otellog.SeverityWarn, "WARN"
	case lvl >= slog.LevelInfo:
		return otellog.SeverityInfo, "INFO"
	default:
		return otellog.SeverityDebug, "DEBUG"
	}
}

func slogAttrToOTel(groups []string, a slog.Attr) otellog.KeyValue {
	key := a.Key
	if len(groups) > 0 {
		key = strings.Join(groups, ".") + "." + key
	}
	v := a.Value.Resolve()
	switch v.Kind() {
	case slog.KindInt64:
		return otellog.Int64(key, v.Int64())
	case slog.KindUint64:
		return otellog.Int64(key, int64(v.Uint64()))
	case slog.KindFloat64:
		return otellog.Float64(key, v.Float64())
	case slog.KindBool:
		return otellog.Bool(key, v.Bool())
	case slog.KindDuration:
		return otellog.Int64(key, v.Duration().Milliseconds())
	default:
		return otellog.String(key, v.String())
	}
}

var _ = os.Stdout
