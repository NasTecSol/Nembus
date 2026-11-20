package telemetry

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type metricInstruments struct {
	httpDurationMs            metric.Float64Histogram
	httpRequestsTotal         metric.Int64Counter
	dbDurationMs              metric.Float64Histogram
	dbErrorsTotal             metric.Int64Counter
	rpcServerDurationMs       metric.Float64Histogram
	rpcServerRequestsTotal    metric.Int64Counter
	rpcClientDurationMs       metric.Float64Histogram
	rpcClientRequestsTotal    metric.Int64Counter
	syncEventsTotal           metric.Int64Counter
	syncDurationMs            metric.Float64Histogram
	posOutboxPending          metric.Int64Gauge
	posOutboxQuarantined      metric.Int64Gauge
	posSyncLagSeconds         metric.Float64Gauge
	zatcaSubmissionsTotal     metric.Int64Counter
	zatcaDurationMs           metric.Float64Histogram
	sapETLRowsTotal           metric.Int64Counter
	sapETLBatchesTotal        metric.Int64Counter
	sapETLDurationMs          metric.Float64Histogram
	biometricCapturesTotal    metric.Int64Counter
	biometricDurationMs       metric.Float64Histogram
	biometricDevicesConnected metric.Int64Gauge
	wsClientsActive           metric.Int64Gauge
	notificationsTotal        metric.Int64Counter
}

var instrumentsPtr atomic.Pointer[metricInstruments]

func init() {
	refreshInstruments()
}

func refreshInstruments() {
	m := Meter("nembus/metrics")
	inst := &metricInstruments{}

	inst.httpDurationMs, _ = m.Float64Histogram("http.server.request.duration_ms", metric.WithUnit("ms"))
	inst.httpRequestsTotal, _ = m.Int64Counter("http.server.requests_total")
	inst.dbDurationMs, _ = m.Float64Histogram("db.client.operation.duration_ms", metric.WithUnit("ms"))
	inst.dbErrorsTotal, _ = m.Int64Counter("db.client.operation.errors_total")
	inst.rpcServerDurationMs, _ = m.Float64Histogram("rpc.server.duration_ms", metric.WithUnit("ms"))
	inst.rpcServerRequestsTotal, _ = m.Int64Counter("rpc.server.requests_total")
	inst.rpcClientDurationMs, _ = m.Float64Histogram("rpc.client.duration_ms", metric.WithUnit("ms"))
	inst.rpcClientRequestsTotal, _ = m.Int64Counter("rpc.client.requests_total")

	inst.syncEventsTotal, _ = m.Int64Counter("nembus.sync.events_total")
	inst.syncDurationMs, _ = m.Float64Histogram("nembus.sync.duration_ms", metric.WithUnit("ms"))

	inst.posOutboxPending, _ = m.Int64Gauge("nembus.pos.outbox.pending_rows")
	inst.posOutboxQuarantined, _ = m.Int64Gauge("nembus.pos.outbox.quarantined_rows")
	inst.posSyncLagSeconds, _ = m.Float64Gauge("nembus.pos.sync.lag_seconds", metric.WithUnit("s"))

	inst.zatcaSubmissionsTotal, _ = m.Int64Counter("nembus.zatca.submissions_total")
	inst.zatcaDurationMs, _ = m.Float64Histogram("nembus.zatca.submission.duration_ms", metric.WithUnit("ms"))

	inst.sapETLRowsTotal, _ = m.Int64Counter("nembus.sap.etl.rows_total")
	inst.sapETLBatchesTotal, _ = m.Int64Counter("nembus.sap.etl.batches_total")
	inst.sapETLDurationMs, _ = m.Float64Histogram("nembus.sap.etl.batch.duration_ms", metric.WithUnit("ms"))

	inst.biometricCapturesTotal, _ = m.Int64Counter("nembus.biometric.captures_total")
	inst.biometricDurationMs, _ = m.Float64Histogram("nembus.biometric.capture.duration_ms", metric.WithUnit("ms"))
	inst.biometricDevicesConnected, _ = m.Int64Gauge("nembus.biometric.devices_connected")

	inst.wsClientsActive, _ = m.Int64Gauge("nembus.websocket.clients_active")
	inst.notificationsTotal, _ = m.Int64Counter("nembus.notifications.dispatched_total")

	instrumentsPtr.Store(inst)
}

func getInstruments() *metricInstruments {
	inst := instrumentsPtr.Load()
	if inst == nil {
		refreshInstruments()
		inst = instrumentsPtr.Load()
	}
	return inst
}

// RecordHTTPRequest records server HTTP request count and duration.
func RecordHTTPRequest(ctx context.Context, serviceName, method, route, tenantSlug string, statusCode int, durationMs float64) {
	inst := getInstruments()
	attrs := metric.WithAttributes(
		attribute.String("http.service", serviceName),
		attribute.String("http.request.method", method),
		attribute.String("http.route", route),
		attribute.Int("http.response.status_code", statusCode),
		attribute.Int("status_class", statusCode/100*100),
		attribute.String("nembus.tenant.slug", tenantSlug),
	)
	if inst.httpDurationMs != nil {
		inst.httpDurationMs.Record(ctx, durationMs, attrs)
	}
	if inst.httpRequestsTotal != nil {
		inst.httpRequestsTotal.Add(ctx, 1, attrs)
	}
}

// RecordSyncBatch records gRPC sync push/pull event counts and latency.
func RecordSyncBatch(ctx context.Context, direction, tenantSlug, entityType, status string, count int64, durationMs float64) {
	inst := getInstruments()
	attrs := metric.WithAttributes(
		attribute.String("direction", direction),
		attribute.String("nembus.tenant.slug", tenantSlug),
		attribute.String("entity_type", entityType),
		attribute.String("status", status),
	)
	if count > 0 && inst.syncEventsTotal != nil {
		inst.syncEventsTotal.Add(ctx, count, attrs)
	}
	if inst.syncDurationMs != nil {
		inst.syncDurationMs.Record(ctx, durationMs, attrs)
	}
}

// RecordPOSOutboxGauge records current local POS outbox queue depth and sync lag.
func RecordPOSOutboxGauge(ctx context.Context, tenantSlug string, pendingCount, quarantinedCount int64, lagSeconds float64) {
	inst := getInstruments()
	attrs := metric.WithAttributes(attribute.String("nembus.tenant.slug", tenantSlug))
	if inst.posOutboxPending != nil {
		inst.posOutboxPending.Record(ctx, pendingCount, attrs)
	}
	if inst.posOutboxQuarantined != nil {
		inst.posOutboxQuarantined.Record(ctx, quarantinedCount, attrs)
	}
	if inst.posSyncLagSeconds != nil {
		inst.posSyncLagSeconds.Record(ctx, lagSeconds, attrs)
	}
}

// RecordZatcaSubmission records automated or interactive ZATCA Phase 2 submission metrics.
func RecordZatcaSubmission(ctx context.Context, appName, tenantSlug, status string, durationMs float64) {
	inst := getInstruments()
	attrs := metric.WithAttributes(
		attribute.String("app", appName),
		attribute.String("nembus.tenant.slug", tenantSlug),
		attribute.String("status", status),
	)
	if inst.zatcaSubmissionsTotal != nil {
		inst.zatcaSubmissionsTotal.Add(ctx, 1, attrs)
	}
	if inst.zatcaDurationMs != nil {
		inst.zatcaDurationMs.Record(ctx, durationMs, attrs)
	}
}

// RecordSAPETLBatch records SAP agent downstream/upstream ETL row and batch metrics.
func RecordSAPETLBatch(ctx context.Context, tenantSlug, direction, entity, status string, rows int64, durationMs float64) {
	inst := getInstruments()
	attrs := metric.WithAttributes(
		attribute.String("nembus.tenant.slug", tenantSlug),
		attribute.String("direction", direction),
		attribute.String("entity", entity),
		attribute.String("status", status),
	)
	if rows > 0 && inst.sapETLRowsTotal != nil {
		inst.sapETLRowsTotal.Add(ctx, rows, attrs)
	}
	if inst.sapETLBatchesTotal != nil {
		inst.sapETLBatchesTotal.Add(ctx, 1, attrs)
	}
	if inst.sapETLDurationMs != nil {
		inst.sapETLDurationMs.Record(ctx, durationMs, attrs)
	}
}

// RecordBiometricCapture records biometric capture status, duration, and connected devices.
func RecordBiometricCapture(ctx context.Context, status string, durationMs float64, connectedDevices int64) {
	inst := getInstruments()
	attrs := metric.WithAttributes(attribute.String("status", status))
	if status != "" {
		if inst.biometricCapturesTotal != nil {
			inst.biometricCapturesTotal.Add(ctx, 1, attrs)
		}
		if inst.biometricDurationMs != nil {
			inst.biometricDurationMs.Record(ctx, durationMs, attrs)
		}
	}
	if connectedDevices >= 0 && inst.biometricDevicesConnected != nil {
		inst.biometricDevicesConnected.Record(ctx, connectedDevices)
	}
}

// RecordWSClients records the active WebSocket client connection count.
func RecordWSClients(ctx context.Context, activeClients int64) {
	inst := getInstruments()
	if inst.wsClientsActive != nil {
		inst.wsClientsActive.Record(ctx, activeClients)
	}
}

// RecordNotificationDispatched records outbound notification delivery counts by channel and status.
func RecordNotificationDispatched(ctx context.Context, channel, status string) {
	inst := getInstruments()
	if inst.notificationsTotal != nil {
		inst.notificationsTotal.Add(ctx, 1, metric.WithAttributes(
			attribute.String("channel", channel),
			attribute.String("status", status),
		))
	}
}
