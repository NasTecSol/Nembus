package telemetry

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

type grpcRPCStateKey struct{}

type grpcRPCState struct {
	start      time.Time
	fullMethod string
	tenantSlug string
}

// GRPCServerStatsHandler returns a gRPC stats.Handler that extracts W3C trace
// context from incoming gRPC metadata, creates server spans, and records RPC metrics.
func GRPCServerStatsHandler() stats.Handler {
	return &grpcStatsHandler{isServer: true}
}

// GRPCClientStatsHandler returns a gRPC stats.Handler that starts client spans,
// injects W3C trace context into outgoing gRPC metadata, and records RPC metrics.
func GRPCClientStatsHandler() stats.Handler {
	return &grpcStatsHandler{isServer: false}
}

type grpcStatsHandler struct {
	isServer bool
}

func (h *grpcStatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	method := "unknown"
	if info != nil && info.FullMethodName != "" {
		method = info.FullMethodName
	}

	var tenantSlug string
	if h.isServer {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			ctx = otel.GetTextMapPropagator().Extract(ctx, metadataCarrier(md))
			if vals := md.Get("x-tenant-id"); len(vals) > 0 && strings.TrimSpace(vals[0]) != "" {
				tenantSlug = strings.TrimSpace(vals[0])
				ctx = WithTenantSlug(ctx, tenantSlug)
			}
		}
		attrs := []attribute.KeyValue{
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.method", method),
		}
		if tenantSlug != "" {
			attrs = append(attrs, attribute.String("nembus.tenant.slug", tenantSlug))
		}
		ctx, _ = Tracer("nembus/grpc-server").Start(
			ctx,
			method,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(attrs...),
		)
	} else {
		tenantSlug = TenantSlugFromContext(ctx)
		if md, ok := metadata.FromOutgoingContext(ctx); ok {
			if tenantSlug == "" {
				if vals := md.Get("x-tenant-id"); len(vals) > 0 {
					tenantSlug = strings.TrimSpace(vals[0])
				}
			}
		}
		attrs := []attribute.KeyValue{
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.method", method),
		}
		if tenantSlug != "" {
			attrs = append(attrs, attribute.String("nembus.tenant.slug", tenantSlug))
		}
		ctx, _ = Tracer("nembus/grpc-client").Start(
			ctx,
			method,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(attrs...),
		)

		outMD, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			outMD = metadata.MD{}
		} else {
			outMD = outMD.Copy()
		}
		otel.GetTextMapPropagator().Inject(ctx, metadataCarrier(outMD))
		ctx = metadata.NewOutgoingContext(ctx, outMD)
	}

	return context.WithValue(ctx, grpcRPCStateKey{}, grpcRPCState{
		start:      time.Now(),
		fullMethod: method,
		tenantSlug: tenantSlug,
	})
}

func (h *grpcStatsHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	end, ok := rs.(*stats.End)
	if !ok {
		return
	}

	span := trace.SpanFromContext(ctx)
	defer span.End()

	state, _ := ctx.Value(grpcRPCStateKey{}).(grpcRPCState)
	durationMs := float64(time.Since(state.start).Microseconds()) / 1000.0
	if state.start.IsZero() {
		durationMs = 0
	}

	grpcCode := codes.OK
	if end.Error != nil && !errors.Is(end.Error, io.EOF) {
		grpcCode = status.Code(end.Error)
		if grpcCode != codes.OK && grpcCode != codes.Canceled {
			span.RecordError(end.Error)
			span.SetStatus(otelcodes.Error, end.Error.Error())
		}
	}
	span.SetAttributes(attribute.String("rpc.grpc.status_code", grpcCode.String()))

	inst := getInstruments()
	attrs := metric.WithAttributes(
		attribute.String("rpc.method", state.fullMethod),
		attribute.String("rpc.grpc.status_code", grpcCode.String()),
		attribute.String("nembus.tenant.slug", state.tenantSlug),
	)
	if h.isServer {
		if inst.rpcServerDurationMs != nil {
			inst.rpcServerDurationMs.Record(ctx, durationMs, attrs)
		}
		if inst.rpcServerRequestsTotal != nil {
			inst.rpcServerRequestsTotal.Add(ctx, 1, attrs)
		}
	} else {
		if inst.rpcClientDurationMs != nil {
			inst.rpcClientDurationMs.Record(ctx, durationMs, attrs)
		}
		if inst.rpcClientRequestsTotal != nil {
			inst.rpcClientRequestsTotal.Add(ctx, 1, attrs)
		}
	}
}

func (h *grpcStatsHandler) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return ctx
}

func (h *grpcStatsHandler) HandleConn(context.Context, stats.ConnStats) {}

type metadataCarrier metadata.MD

func (m metadataCarrier) Get(key string) string {
	vals := metadata.MD(m).Get(strings.ToLower(key))
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

func (m metadataCarrier) Set(key, value string) {
	metadata.MD(m).Set(strings.ToLower(key), value)
}

func (m metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
