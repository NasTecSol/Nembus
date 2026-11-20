package telemetry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type pgxTraceCtxKey struct{}

type pgxTraceState struct {
	start  time.Time
	op     string
	tenant string
}

// PGXTracer implements pgx.QueryTracer to emit PostgreSQL child spans and
// sanitized query latency/error metrics without logging parameter values.
type PGXTracer struct {
	dbRole      string
	defaultSlug string
}

// NewPGXTracer creates a pgx.QueryTracer tagged with the given dbRole (e.g. "master", "tenant:acme", "pos-local").
func NewPGXTracer(dbRole string) *PGXTracer {
	var slug string
	if strings.HasPrefix(dbRole, "tenant:") {
		slug = strings.TrimPrefix(dbRole, "tenant:")
	}
	return &PGXTracer{
		dbRole:      dbRole,
		defaultSlug: slug,
	}
}

func (t *PGXTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	op := parseSQLOperation(data.SQL)
	tenant := TenantSlugFromContext(ctx)
	if tenant == "" {
		tenant = t.defaultSlug
	}

	attrs := []attribute.KeyValue{
		attribute.String("db.system.name", "postgresql"),
		attribute.String("db.operation.name", op),
		attribute.String("nembus.db.role", t.dbRole),
	}
	if tenant != "" {
		attrs = append(attrs, attribute.String("nembus.tenant.slug", tenant))
	}

	ctx, _ = Tracer("nembus/pgx").Start(
		ctx,
		fmt.Sprintf("postgresql.%s", op),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	)

	return context.WithValue(ctx, pgxTraceCtxKey{}, pgxTraceState{
		start:  time.Now(),
		op:     op,
		tenant: tenant,
	})
}

func (t *PGXTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	state, _ := ctx.Value(pgxTraceCtxKey{}).(pgxTraceState)
	durationMs := float64(time.Since(state.start).Microseconds()) / 1000.0
	if state.start.IsZero() {
		durationMs = 0
	}

	isErr := data.Err != nil && !errors.Is(data.Err, pgx.ErrNoRows)
	if isErr {
		span.RecordError(data.Err)
		span.SetStatus(codes.Error, data.Err.Error())
	}

	inst := getInstruments()
	attrs := metric.WithAttributes(
		attribute.String("db.operation.name", state.op),
		attribute.String("nembus.db.role", t.dbRole),
		attribute.String("nembus.tenant.slug", state.tenant),
	)
	if inst.dbDurationMs != nil {
		inst.dbDurationMs.Record(ctx, durationMs, attrs)
	}
	if isErr && inst.dbErrorsTotal != nil {
		inst.dbErrorsTotal.Add(ctx, 1, attrs)
	}
}

// AttachPoolTracer configures a pgxpool.Config with PGXTracer.
func AttachPoolTracer(cfg *pgxpool.Config, dbRole string) {
	if cfg == nil || cfg.ConnConfig == nil {
		return
	}
	cfg.ConnConfig.Tracer = NewPGXTracer(dbRole)
}

// RegisterPoolMetrics registers observable gauges tracking pgxpool usage, idle,
// and max connection counts labeled by pool_name and nembus.tenant.slug.
func RegisterPoolMetrics(pool *pgxpool.Pool, poolName string) {
	if pool == nil {
		return
	}
	tenantSlug := ""
	if strings.HasPrefix(poolName, "tenant:") {
		tenantSlug = strings.TrimPrefix(poolName, "tenant:")
	}

	meter := Meter("nembus/pgx-pool")
	usageGauge, err1 := meter.Int64ObservableGauge(
		"db.client.connections.usage",
		metric.WithDescription("Active acquired connections in the PostgreSQL pool"),
	)
	idleGauge, err2 := meter.Int64ObservableGauge(
		"db.client.connections.idle",
		metric.WithDescription("Idle connections in the PostgreSQL pool"),
	)
	maxGauge, err3 := meter.Int64ObservableGauge(
		"db.client.connections.max",
		metric.WithDescription("Configured maximum connections in the PostgreSQL pool"),
	)
	if err1 != nil || err2 != nil || err3 != nil {
		return
	}

	commonAttrs := metric.WithAttributes(
		attribute.String("pool.name", poolName),
		attribute.String("nembus.tenant.slug", tenantSlug),
	)
	_, _ = meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		st := pool.Stat()
		if st == nil {
			return nil
		}
		o.ObserveInt64(usageGauge, int64(st.AcquiredConns()), commonAttrs)
		o.ObserveInt64(idleGauge, int64(st.IdleConns()), commonAttrs)
		o.ObserveInt64(maxGauge, int64(st.MaxConns()), commonAttrs)
		return nil
	}, usageGauge, idleGauge, maxGauge)
}

func parseSQLOperation(sql string) string {
	trimmed := strings.TrimSpace(sql)
	for strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "/*") {
		if strings.HasPrefix(trimmed, "--") {
			idx := strings.IndexByte(trimmed, '\n')
			if idx < 0 {
				return "SQL"
			}
			trimmed = strings.TrimSpace(trimmed[idx+1:])
			continue
		}
		end := strings.Index(trimmed, "*/")
		if end < 0 {
			return "SQL"
		}
		trimmed = strings.TrimSpace(trimmed[end+2:])
	}
	if trimmed == "" {
		return "SQL"
	}
	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return "SQL"
	}
	verb := strings.ToUpper(fields[0])
	switch verb {
	case "SELECT", "INSERT", "UPDATE", "DELETE", "WITH", "BEGIN", "COMMIT", "ROLLBACK", "CREATE", "ALTER", "DROP", "CALL":
		return verb
	default:
		return "SQL"
	}
}
