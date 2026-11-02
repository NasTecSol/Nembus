package usecase

import (
	"context"
	"fmt"
	"testing"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type selectionRow struct{ scan func(...interface{}) error }

func (r selectionRow) Scan(dest ...interface{}) error { return r.scan(dest...) }

type selectionDB struct {
	repository.DBTX
	device   repository.ZatcaDeviceConfig
	count    int64
	selected int
}

func (db *selectionDB) QueryRow(_ context.Context, sql string, args ...interface{}) pgx.Row {
	if len(args) == 3 {
		return selectionRow{func(dest ...interface{}) error { *dest[0].(*int64) = db.count; return nil }}
	}
	db.selected++
	return selectionRow{func(dest ...interface{}) error { return scanOrderTestValue(db.device, dest...) }}
}

func TestDeviceSelectionRejectsAmbiguityAndSupportsExplicitDevice(t *testing.T) {
	for _, scenario := range []string{"single", "ambiguous", "explicit", "wrong-org", "invalid-env", "relabelled"} {
		t.Run(scenario, func(t *testing.T) {
			db := &selectionDB{count: 1, device: repository.ZatcaDeviceConfig{ID: 3, OrganizationID: 1, DeviceType: "pos", PosTerminalID: pgtype.Int4{Int32: 1, Valid: true}, IsActive: pgtype.Bool{Bool: true, Valid: true}, ZatcaEnv: pgtype.Text{String: "simulation", Valid: true}}}
			var metadata []byte
			switch scenario {
			case "ambiguous":
				db.count = 2
			case "explicit":
				db.count = 2
				metadata = []byte(`{"zatca_device_config_id":3}`)
			case "wrong-org":
				db.device.OrganizationID = 2
			case "invalid-env":
				db.device.ZatcaEnv.String = "invalid"
			case "relabelled":
				db.device.Metadata = []byte(`{"credential_environment":"production"}`)
			}
			uc := NewZatcaUseCase(&ZatcaConfig{Enabled: true, Env: "production"})
			uc.SetRepository(repository.New(db))
			device, err := uc.SelectZatcaDevice(context.Background(), 1, pgtype.Int4{Int32: 1, Valid: true}, metadata)
			if scenario == "single" || scenario == "explicit" {
				if err != nil || device.ID != 3 {
					t.Fatal("saved simulation device should work despite global production setting", err)
				}
			} else if err == nil {
				t.Fatal("unsafe device selection succeeded")
			}
			if scenario == "ambiguous" && db.selected != 0 {
				t.Fatal("selected a device despite ambiguity")
			}
		})
	}
}

func TestEnvironmentNamesAreValidatedWithoutFallback(t *testing.T) {
	for _, env := range []string{"", "stg", "prod", "https://example.com"} {
		if url, err := ZatcaEnvironmentBaseURL(env); err == nil {
			t.Fatal(fmt.Sprintf("%q silently became %s", env, url))
		}
	}
}
