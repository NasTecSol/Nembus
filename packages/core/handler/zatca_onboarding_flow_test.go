package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type onboardingHandlerStub struct {
	device repository.ZatcaDeviceConfig
	err    error
}

func TestOnboardRejectsAnotherOrganizationBeforeGateway(t *testing.T) {
	stub := onboardingHandlerStub{device: repository.ZatcaDeviceConfig{ID: 3}}
	handler := NewZatcaHandler(nil)
	handler.SetOnboarder(stub)
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(string(middleware.OrgIDKey), int32(2)) })
	engine.POST("/api/zatca/onboard", handler.OnboardDevice)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/api/zatca/onboard", strings.NewReader(`{"organization_id":1,"device_serial":"POS-01","otp":"123456","zatca_env":"simulation"}`))
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, request)
	if recorder.Code != 403 {
		t.Fatalf("HTTP %d: %s", recorder.Code, recorder.Body)
	}
}

func (s onboardingHandlerStub) OnboardDevice(context.Context, int32, *int32, *int32, string, string, string, string) (*repository.ZatcaDeviceConfig, error) {
	return &s.device, s.err
}

type onboardingProgressError struct{}

func (onboardingProgressError) Error() string { return "sample check failed" }
func (onboardingProgressError) OnboardingFailureDetails() json.RawMessage {
	return []byte(`{"device_id":3,"id":3,"reporting_ready":false,"onboarding_stage":"compliance","completion_endpoint":"/api/zatca/devices/3/complete-onboarding"}`)
}

type onboardingOTPError struct{}

func (onboardingOTPError) Error() string             { return "compliance API error: Invalid-OTP" }
func (onboardingOTPError) OnboardingHTTPStatus() int { return 400 }
func (onboardingOTPError) OnboardingFailureDetails() json.RawMessage {
	return []byte(`{"onboarding_stage":"compliance_csid","zatca_env":"simulation","gateway_http_status":400,"csr_organization_identifier":"310123456700003","device_saved":false}`)
}

func TestOnboardHTTPReturnsSuccessOnlyForReportingReadyDevice(t *testing.T) {
	for _, scenario := range []struct {
		name string
		code int
	}{
		{"ready", 200}, {"partial-error", 502}, {"invalid-otp", 400}, {"compliance-only", 502}, {"no-gateway", 501}, {"organization-disabled", 409}, {"missing-env", 400}, {"invalid-env", 400},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			handler := NewZatcaHandler(nil)
			if scenario.name != "no-gateway" {
				stub := onboardingHandlerStub{device: repository.ZatcaDeviceConfig{ID: 3, ComplianceCsid: pgtype.Text{String: "ccsid", Valid: true}, Metadata: []byte(`{"production_secret":"must-not-leak"}`)}}
				if scenario.name == "ready" {
					stub.device.ProductionCsid = pgtype.Text{String: "pcsid", Valid: true}
				}
				if scenario.name == "organization-disabled" {
					stub.err = usecase.ErrZatcaDisabled
				}
				if scenario.name == "partial-error" {
					stub.err = onboardingProgressError{}
				}
				if scenario.name == "invalid-otp" {
					stub.err = onboardingOTPError{}
				}
				handler.SetOnboarder(stub)
			}
			engine := gin.New()
			engine.POST("/api/zatca/onboard", handler.OnboardDevice)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest("POST", "/api/zatca/onboard", strings.NewReader(`{"organization_id":1,"device_serial":"NEW-DEVICE","device_type":"pos","otp":"test-otp","zatca_env":"simulation"}`))
			if scenario.name == "missing-env" {
				request = httptest.NewRequest("POST", "/api/zatca/onboard", strings.NewReader(`{"organization_id":1,"device_serial":"NEW-DEVICE","otp":"test-otp"}`))
			}
			if scenario.name == "invalid-env" {
				request = httptest.NewRequest("POST", "/api/zatca/onboard", strings.NewReader(`{"organization_id":1,"device_serial":"NEW-DEVICE","otp":"test-otp","zatca_env":"staging"}`))
			}
			request.Header.Set("Content-Type", "application/json")
			engine.ServeHTTP(recorder, request)
			if recorder.Code != scenario.code {
				t.Fatalf("HTTP %d: %s", recorder.Code, recorder.Body)
			}
			if strings.Contains(recorder.Body.String(), "must-not-leak") || strings.Contains(recorder.Body.String(), "SIMULATED_") {
				t.Fatal("fake credentials or secrets exposed")
			}
			if scenario.name == "ready" && !strings.Contains(recorder.Body.String(), `"reporting_ready":true`) {
				t.Fatal("completion missing")
			}
			if scenario.name == "partial-error" && !strings.Contains(recorder.Body.String(), `"device_id":3`) {
				t.Fatal("recovery device missing")
			}
			if scenario.name == "invalid-otp" && (!strings.Contains(recorder.Body.String(), `"csr_organization_identifier":"310123456700003"`) || !strings.Contains(recorder.Body.String(), `"device_saved":false`)) {
				t.Fatal("compliance request diagnostics missing")
			}
		})
	}
}
