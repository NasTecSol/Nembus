package middleware

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/telemetry"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// LoggerMiddleware provides structured request logging and OpenTelemetry tracing.
func LoggerMiddleware() gin.HandlerFunc {
	return OTelGinMiddleware("nembus-http")
}

// OTelGinMiddleware extracts W3C trace context, starts a server span per HTTP
// request, propagates trace/tenant attributes to downstream handlers and logs,
// sets X-Trace-ID on responses, and records HTTP request metrics.
func OTelGinMiddleware(serviceName string) gin.HandlerFunc {
	if strings.TrimSpace(serviceName) == "" {
		serviceName = "nembus-http"
	}

	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		rawPath := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		if tenantHeader := strings.TrimSpace(c.GetHeader("x-tenant-id")); tenantHeader != "" {
			ctx = telemetry.WithTenantSlug(ctx, tenantHeader)
		}

		ctx, span := telemetry.Tracer("nembus/http-server").Start(
			ctx,
			fmt.Sprintf("%s %s", method, rawPath),
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.service", serviceName),
				attribute.String("http.request.method", method),
				attribute.String("url.path", rawPath),
			),
		)
		defer span.End()

		if sc := span.SpanContext(); sc.IsValid() {
			c.Writer.Header().Set("X-Trace-ID", sc.TraceID().String())
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		latency := time.Since(start)
		latencyMs := float64(latency.Microseconds()) / 1000.0
		route := c.FullPath()
		if route == "" {
			route = rawPath
		} else {
			span.SetName(fmt.Sprintf("%s %s", method, route))
		}

		reqCtx := c.Request.Context()
		tenantSlug := telemetry.TenantSlugFromContext(reqCtx)
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()
		clientIP := c.ClientIP()

		spanAttrs := []attribute.KeyValue{
			attribute.String("http.route", route),
			attribute.Int("http.response.status_code", statusCode),
		}
		if tenantSlug != "" {
			spanAttrs = append(spanAttrs, attribute.String("nembus.tenant.slug", tenantSlug))
		}
		span.SetAttributes(spanAttrs...)

		if statusCode >= 500 || errorMessage != "" {
			span.SetStatus(codes.Error, errorMessage)
		}

		telemetry.RecordHTTPRequest(reqCtx, serviceName, method, route, tenantSlug, statusCode, latencyMs)

		fullPath := rawPath
		if rawQuery != "" {
			fullPath = rawPath + "?" + rawQuery
		}

		logAttrs := []slog.Attr{
			slog.String("http.service", serviceName),
			slog.String("http.method", method),
			slog.String("http.route", route),
			slog.String("http.path", fullPath),
			slog.Int("http.status_code", statusCode),
			slog.Float64("duration_ms", latencyMs),
			slog.String("client_ip", clientIP),
		}
		if tenantSlug != "" {
			logAttrs = append(logAttrs, slog.String("nembus.tenant.slug", tenantSlug))
		}

		if statusCode >= 500 || errorMessage != "" {
			if errorMessage != "" {
				logAttrs = append(logAttrs, slog.String("error", errorMessage))
			}
			telemetry.LogError(reqCtx, "http request failed", nil, logAttrs...)
		} else {
			telemetry.LogInfo(reqCtx, "http request completed", logAttrs...)
		}
	}
}
