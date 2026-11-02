package config

import "testing"

// Exercise the actual application config loader: the HTTP client also has URL
// defaults, but the app supplies its own URL and must not silently use sandbox.
func TestZatcaEnvironmentGateway(t *testing.T) {
	for _, test := range []struct{ env, suffix string }{
		{"sandbox", "developer-portal"}, {"simulation", "simulation"}, {"production", "core"},
	} {
		t.Run(test.env, func(t *testing.T) {
			t.Setenv("ZATCA_ENV", test.env)
			cfg := LoadConfig("stg")
			want := "https://gw-fatoora.zatca.gov.sa/e-invoicing/" + test.suffix
			if cfg.ZatcaEnv != test.env || cfg.ZatcaBaseURL != want {
				t.Fatalf("environment %s resolved to %s", cfg.ZatcaEnv, cfg.ZatcaBaseURL)
			}
		})
	}
}
