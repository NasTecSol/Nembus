package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/handler"
	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	router "github.com/NasTecSol/nembus-core/routing"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const zatcaSyncTestDDL = `
CREATE TABLE organizations(id int PRIMARY KEY);
CREATE TABLE stores(id int PRIMARY KEY,organization_id int REFERENCES organizations);
CREATE TABLE pos_terminals(id int PRIMARY KEY,store_id int REFERENCES stores);
CREATE TYPE zatca_doc_status AS ENUM('pending','cleared','reported','warning','rejected','failed');
CREATE TABLE zatca_device_configs(id serial PRIMARY KEY,organization_id int NOT NULL REFERENCES organizations,store_id int REFERENCES stores,pos_terminal_id int REFERENCES pos_terminals,
 device_serial text NOT NULL,device_type text NOT NULL,csr_pem text,private_key_pem text NOT NULL,compliance_csid text,production_csid text,csid_expiry timestamptz,
 zatca_env text,is_active boolean DEFAULT true,is_revoked boolean DEFAULT false,revoked_at timestamptz,revoked_reason text,metadata jsonb DEFAULT '{}',created_at timestamp DEFAULT NOW(),updated_at timestamp DEFAULT NOW(),UNIQUE(organization_id,device_serial));
CREATE TABLE zatca_document_chain(id bigserial PRIMARY KEY,entity_type text NOT NULL,entity_id text NOT NULL,device_config_id int NOT NULL REFERENCES zatca_device_configs,organization_id int NOT NULL REFERENCES organizations,
 zatca_uuid uuid NOT NULL,icv bigint NOT NULL,pih text NOT NULL,xml_hash text NOT NULL,zatca_status zatca_doc_status DEFAULT 'pending',zatca_response jsonb DEFAULT '{}',qr_code_base64 text,signed_xml text,
 submitted_at timestamptz,cleared_at timestamptz,created_at timestamp DEFAULT NOW(),updated_at timestamp DEFAULT NOW(),UNIQUE(device_config_id,icv));
CREATE TABLE sync_queue(id bigserial PRIMARY KEY,entity_type text,entity_id bigint,action text,payload jsonb,status text DEFAULT 'pending',priority int DEFAULT 0,correlation_id text,max_retries int DEFAULT 5,retry_count int DEFAULT 0,last_error text,created_at timestamptz DEFAULT NOW(),synced_at timestamptz);
CREATE TABLE local_device_config(id serial PRIMARY KEY,store_id int,pos_terminal_id int);
INSERT INTO organizations VALUES(1),(2);
INSERT INTO stores VALUES(1,1),(2,2);
INSERT INTO pos_terminals VALUES(8,1),(9,1),(20,2);
INSERT INTO local_device_config(store_id,pos_terminal_id) VALUES(1,8);
`

func newZatcaSyncTestPool(t *testing.T, ctx context.Context, dsn, label string) *pgxpool.Pool {
	t.Helper()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("zatca_sync_%s_%d", label, time.Now().UnixNano())
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize()); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		_, _ = admin.Exec(context.Background(), "DROP SCHEMA "+pgx.Identifier{schema}.Sanitize()+" CASCADE")
		admin.Close()
	})
	if _, err := pool.Exec(ctx, zatcaSyncTestDDL); err != nil {
		t.Fatal(err)
	}
	return pool
}

func insertZatcaSyncTestDevice(t *testing.T, ctx context.Context, pool *pgxpool.Pool, serial, key string, terminal int32) int32 {
	t.Helper()
	var id int32
	err := pool.QueryRow(ctx, `INSERT INTO zatca_device_configs(organization_id,store_id,pos_terminal_id,device_serial,device_type,private_key_pem,compliance_csid,production_csid,csid_expiry,zatca_env,metadata)
 VALUES(1,1,$1,$2,'pos',$3,'compliance-certificate','reporting-certificate',NOW()+interval '1 year','simulation','{"onboarding_stage":"production","production_secret":"test-secret","credential_environment":"simulation"}') RETURNING id`, terminal, serial, key).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// Dedicated PostgreSQL only. Fixtures use isolated schemas and mock certificates;
// the HTTPS loopback server is the actual authenticated cloud handler.
func TestZatcaOutboxCloudAndCloneRecoveryPostgres(t *testing.T) {
	dsn := os.Getenv("NEMBUS_ZATCA_SYNC_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set NEMBUS_ZATCA_SYNC_TEST_DATABASE_URL for isolated ZATCA sync integration tests")
	}
	t.Chdir(t.TempDir())
	t.Setenv("JWT_SECRET", "zatca-sync-test-only-signing-secret")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	source := newZatcaSyncTestPool(t, ctx, dsn, "source")
	cloudPool := newZatcaSyncTestPool(t, ctx, dsn, "cloud")
	restored := newZatcaSyncTestPool(t, ctx, dsn, "restored")
	const serial = "PHYSICAL-POS-08"
	sourceID := insertZatcaSyncTestDevice(t, ctx, source, serial, "PRIVATE-LOCAL-A", 8)
	otherID := insertZatcaSyncTestDevice(t, ctx, cloudPool, "OTHER-POS", "PRIVATE-OTHER", 9)
	if sourceID != otherID {
		t.Fatal("fixture must reproduce numeric device-ID collision")
	}
	insertZatcaSyncTestDevice(t, ctx, restored, "OTHER-RESTORED-1", "PRIVATE-OTHER-1", 9)
	insertZatcaSyncTestDevice(t, ctx, restored, "OTHER-RESTORED-2", "PRIVATE-OTHER-2", 9)
	for icv := 1; icv <= 23; icv++ {
		pih := "GENESIS"
		if icv > 1 {
			pih = fmt.Sprintf("hash-%d", icv-1)
		}
		if _, err := source.Exec(ctx, `INSERT INTO zatca_document_chain(entity_type,entity_id,device_config_id,organization_id,zatca_uuid,icv,pih,xml_hash,zatca_status,signed_xml,qr_code_base64) VALUES('sales_order_v2',$1,$2,1,$3,$4,$5,$6,'reported',$7,'qr')`, uuid.NewString(), sourceID, uuid.New(), icv, pih, fmt.Sprintf("hash-%d", icv), fmt.Sprintf("<Invoice>%d</Invoice>", icv)); err != nil {
			t.Fatal(err)
		}
	}
	for _, pool := range []*pgxpool.Pool{source, restored} {
		if _, err := pool.Exec(ctx, ZatcaOutboxDDL); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := source.Exec(ctx, backfillZatcaOutboxSQL); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := source.QueryRow(ctx, `SELECT count(*) FROM sync_queue WHERE status='pending'`).Scan(&count); err != nil || count != 24 {
		t.Fatalf("expected complete registration and chain backfill, count=%d err=%v", count, err)
	}
	if err := source.QueryRow(ctx, `SELECT count(*) FROM sync_queue WHERE payload ?| ARRAY['private_key_pem','metadata','signed_xml','compliance_csid','production_csid']`).Scan(&count); err != nil || count != 0 {
		t.Fatal("outbox copied credentials or signed XML instead of keeping lightweight references")
	}
	var rows pgx.Rows
	var err error
	rows, err = source.Query(ctx, pendingOutboxSQL)
	if err != nil {
		t.Fatal(err)
	}
	if rows.Next() {
		t.Fatal("credential payload was eligible for plaintext gRPC")
	}
	rows.Close()
	token, err := middleware.GenerateM2MToken("pos-sync-test", "POS sync test", "test-tenant", nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	engine := gin.New()
	engine.GET("/api/m2m/tokens", middleware.M2MAPIKeyMiddleware(), func(c *gin.Context) {
		c.JSON(200, []middleware.M2MClient{{ClientID: "pos-sync-test", TenantID: "test-tenant", Token: token, IsActive: true}})
	})
	api := engine.Group("/api")
	api.Use(middleware.JWTAuthMiddleware())
	api.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), middleware.RepoKey, repository.New(cloudPool)))
	})
	router.RegisterZatcaRoutes(api, handler.NewZatcaHandler(nil))
	server := httptest.NewServer(engine)
	defer server.Close()
	service := NewSyncService(ctx, source, "localhost:1", "test-tenant")
	defer service.Stop()
	service.ConfigureZatcaSync(server.URL)
	service.drainZatcaOutbox()
	if err := source.QueryRow(ctx, `SELECT count(*) FROM sync_queue WHERE status<>'synced'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("ZATCA queue was not acknowledged, remaining=%d err=%v", count, err)
	}
	// Subsequent credential/report changes are queued transactionally as well.
	if _, err := source.Exec(ctx, `UPDATE zatca_device_configs SET updated_at=NOW() WHERE id=$1`, sourceID); err != nil {
		t.Fatal(err)
	}
	if _, err := source.Exec(ctx, `UPDATE zatca_document_chain SET zatca_response='{"reportingStatus":"REPORTED"}',updated_at=NOW() WHERE device_config_id=$1 AND icv=1`, sourceID); err != nil {
		t.Fatal(err)
	}
	if err := source.QueryRow(ctx, `SELECT count(*) FROM sync_queue WHERE status='pending'`).Scan(&count); err != nil || count != 2 {
		t.Fatal("device/document updates did not enqueue both sync events")
	}
	if err := source.QueryRow(ctx, `SELECT count(*) FROM sync_queue WHERE payload ?| ARRAY['private_key_pem','metadata','signed_xml','compliance_csid','production_csid']`).Scan(&count); err != nil || count != 0 {
		t.Fatal("update triggers copied credential payloads")
	}
	service.drainZatcaOutbox()
	cloudRepo := repository.New(cloudPool)
	t.Run("legacy cloud registration cannot activate a second signer", func(t *testing.T) {
		if _, err := cloudRepo.ZatcaRecoveryPage(ctx, "OTHER-POS", 1, 0, -1, 20); err == nil {
			t.Fatal("recovery allowed a device whose signing ownership has not been established")
		}
	})
	device, err := cloudRepo.GetZatcaDeviceConfigBySerial(ctx, repository.GetZatcaDeviceConfigBySerialParams{OrganizationID: 1, DeviceSerial: serial})
	if err != nil {
		t.Fatal(err)
	}
	var ownership struct {
		Owner string `json:"signing_owner"`
	}
	if json.Unmarshal(device.Metadata, &ownership) != nil {
		t.Fatal("invalid cloud device metadata")
	}
	if device.ID == sourceID || device.PrivateKeyPem != "PRIVATE-LOCAL-A" || ownership.Owner != "pos" {
		t.Fatal("cloud did not map and preserve the owning POS registration")
	}
	var otherKey string
	if err := cloudPool.QueryRow(ctx, `SELECT private_key_pem FROM zatca_device_configs WHERE id=$1`, otherID).Scan(&otherKey); err != nil || otherKey != "PRIVATE-OTHER" {
		t.Fatal("unrelated cloud device was overwritten")
	}
	if err := cloudPool.QueryRow(ctx, `SELECT count(*) FROM zatca_document_chain WHERE device_config_id=$1`, device.ID).Scan(&count); err != nil || count != 23 {
		t.Fatalf("cloud chain count=%d err=%v", count, err)
	}
	// Replaying the same queued snapshot cannot duplicate either identity.
	if _, err := source.Exec(ctx, `UPDATE sync_queue SET status='pending'`); err != nil {
		t.Fatal(err)
	}
	service.drainZatcaOutbox()
	if err := cloudPool.QueryRow(ctx, `SELECT count(*) FROM zatca_document_chain`).Scan(&count); err != nil || count != 23 {
		t.Fatal("retry duplicated the device chain")
	}
	var deviceJSON, documentJSON json.RawMessage
	if err := source.QueryRow(ctx, `SELECT row_to_json(d),row_to_json(c) FROM zatca_device_configs d JOIN zatca_document_chain c ON c.device_config_id=d.id WHERE d.id=$1 AND c.icv=1`, sourceID).Scan(&deviceJSON, &documentJSON); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"different key", "foreign store", "forked document", "missing predecessor", "wrong PIH"} {
		t.Run(name, func(t *testing.T) {
			var d, c map[string]any
			_ = json.Unmarshal(deviceJSON, &d)
			_ = json.Unmarshal(documentJSON, &c)
			snapshot := repository.ZatcaSyncSnapshot{Device: deviceJSON}
			switch name {
			case "different key":
				d["private_key_pem"] = "DIFFERENT-PRIVATE-KEY"
			case "foreign store":
				d["organization_id"] = 2
			case "forked document":
				c["xml_hash"] = "forked-hash"
				snapshot.Document, _ = json.Marshal(c)
			case "missing predecessor":
				c["icv"] = 25
				c["zatca_uuid"] = uuid.NewString()
				snapshot.Document, _ = json.Marshal(c)
			case "wrong PIH":
				c["icv"] = 24
				c["pih"] = "wrong-hash"
				c["zatca_uuid"] = uuid.NewString()
				snapshot.Document, _ = json.Marshal(c)
			}
			snapshot.Device, _ = json.Marshal(d)
			if err := cloudRepo.ApplyZatcaSnapshot(ctx, 1, 0, snapshot); err == nil {
				t.Fatal("unsafe ZATCA snapshot was accepted")
			}
			var head int64
			if err := cloudPool.QueryRow(ctx, `SELECT MAX(icv) FROM zatca_document_chain WHERE device_config_id=$1`, device.ID).Scan(&head); err != nil || head != 23 {
				t.Fatal("rejected snapshot changed the device chain")
			}
		})
	}
	t.Run("terminal status cannot regress", func(t *testing.T) {
		var c map[string]any
		_ = json.Unmarshal(documentJSON, &c)
		c["zatca_status"] = "pending"
		raw, _ := json.Marshal(c)
		if err := cloudRepo.ApplyZatcaSnapshot(ctx, 1, 1, repository.ZatcaSyncSnapshot{Device: deviceJSON, Document: raw}); err != nil {
			t.Fatal(err)
		}
		var status string
		if err := cloudPool.QueryRow(ctx, `SELECT zatca_status FROM zatca_document_chain WHERE device_config_id=$1 AND icv=1`, device.ID).Scan(&status); err != nil || status != "reported" {
			t.Fatal("reported document regressed")
		}
	})
	t.Run("ready registration cannot regress", func(t *testing.T) {
		var d map[string]any
		_ = json.Unmarshal(deviceJSON, &d)
		d["production_csid"] = nil
		d["metadata"] = map[string]any{"onboarding_stage": "compliance"}
		raw, _ := json.Marshal(d)
		if err := cloudRepo.ApplyZatcaSnapshot(ctx, 1, 1, repository.ZatcaSyncSnapshot{Device: raw}); err != nil {
			t.Fatal(err)
		}
		current, err := cloudRepo.GetZatcaDeviceConfigBySerial(ctx, repository.GetZatcaDeviceConfigBySerialParams{OrganizationID: 1, DeviceSerial: serial})
		if err != nil || current.ProductionCsid.String != "reporting-certificate" {
			t.Fatal("stale snapshot erased reporting credentials")
		}
	})
	if err := GuardZatcaClone(ctx, source, server.URL, "test-tenant", serial); err != nil {
		t.Fatalf("fully backed-up device should be cloneable: %v", err)
	}
	if _, err := source.Exec(ctx, `UPDATE sync_queue SET status='pending' WHERE entity_type='zatca_device_configs'`); err != nil {
		t.Fatal(err)
	}
	if err := GuardZatcaClone(ctx, source, server.URL, "test-tenant", serial); err == nil {
		t.Fatal("clone accepted an unacknowledged device backup")
	}
	service.drainZatcaOutbox()
	if err := RecoverZatcaDevice(ctx, repository.New(restored), server.URL, "test-tenant", serial, 1); err != nil {
		t.Fatalf("recover registration and paginated chain: %v", err)
	}
	recovered, err := repository.New(restored).GetZatcaDeviceConfigBySerial(ctx, repository.GetZatcaDeviceConfigBySerialParams{OrganizationID: 1, DeviceSerial: serial})
	if err != nil {
		t.Fatal(err)
	}
	if recovered.ID == device.ID || recovered.PrivateKeyPem != device.PrivateKeyPem || recovered.ProductionCsid != device.ProductionCsid {
		t.Fatal("recovery did not preserve credentials or map the receiving IDs")
	}
	uc := usecase.NewZatcaUseCase(&usecase.ZatcaConfig{LocalPOS: true})
	uc.SetRepository(repository.New(restored))
	next, pih, err := uc.GetNextICVAndPIH(ctx, recovered.ID)
	if err != nil || next != 24 || pih != "hash-23" {
		t.Fatalf("recovery reset the invoice chain: next=%d pih=%q err=%v", next, pih, err)
	}
	if err := restored.QueryRow(ctx, `SELECT count(*) FROM sync_queue WHERE status<>'synced'`).Scan(&count); err != nil || count != 0 {
		t.Fatal("restore generated echo uploads")
	}
	if err := usecase.RequireZatcaSigningOwner(&recovered, &usecase.ZatcaConfig{}); err == nil {
		t.Fatal("cloud process was permitted to sign with the POS-owned key")
	}
	if err := usecase.RequireZatcaSigningOwner(&recovered, &usecase.ZatcaConfig{LocalPOS: true}); err != nil {
		t.Fatal(err)
	}
	t.Run("interrupted recovery rolls back all pages", func(t *testing.T) {
		interrupted := newZatcaSyncTestPool(t, ctx, dsn, "interrupted")
		client, err := NewZatcaCloud(ctx, server.URL, "test-tenant")
		if err != nil {
			t.Fatal(err)
		}
		err = repository.New(interrupted).RestoreZatcaDevice(ctx, serial, 1, func(after, through int64) (repository.ZatcaRestorePage, error) {
			if after > 0 {
				return repository.ZatcaRestorePage{}, fmt.Errorf("simulated interrupted recovery")
			}
			return client.RecoveryPage(ctx, serial, 1, after, through)
		})
		if err == nil {
			t.Fatal("interrupted recovery was accepted")
		}
		if err := interrupted.QueryRow(ctx, `SELECT count(*) FROM zatca_device_configs`).Scan(&count); err != nil || count != 0 {
			t.Fatal("partial device was committed")
		}
		if err := interrupted.QueryRow(ctx, `SELECT count(*) FROM zatca_document_chain`).Scan(&count); err != nil || count != 0 {
			t.Fatal("partial chain was committed")
		}
	})
	t.Run("recovery cannot roll back a newer local counter", func(t *testing.T) {
		if _, err := restored.Exec(ctx, `INSERT INTO zatca_document_chain(entity_type,entity_id,device_config_id,organization_id,zatca_uuid,icv,pih,xml_hash,zatca_status,signed_xml) VALUES('sales_order_v2',$1,$2,1,$3,24,'hash-23','hash-24','reported','<Invoice>24</Invoice>')`, uuid.NewString(), recovered.ID, uuid.New()); err != nil {
			t.Fatal(err)
		}
		if err := RecoverZatcaDevice(ctx, repository.New(restored), server.URL, "test-tenant", serial, 1); err == nil {
			t.Fatal("newer local chain was overwritten")
		}
		var head int64
		if err := restored.QueryRow(ctx, `SELECT MAX(icv) FROM zatca_document_chain WHERE device_config_id=$1`, recovered.ID).Scan(&head); err != nil || head != 24 {
			t.Fatal("local high-water mark was lost")
		}
	})
}
