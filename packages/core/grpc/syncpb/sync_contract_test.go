package syncpb

import (
	"bytes"
	"math"
	"testing"
	"time"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestIsUUIDEntity(t *testing.T) {
	migratedTables := []string{
		"user_biometric_templates",
		"cashier_sessions",
		"stock_movements",
		"stock_counts",
		"stock_count_lines",
		"customers",
		"pos_transactions",
		"pos_transaction_lines",
		"pos_payments",
		"restaurant_orders",
		"restaurant_order_items",
		"waste_logs",
		"kiosk_sessions",
		"sales_returns",
		"sales_return_lines",
	}
	for _, tbl := range migratedTables {
		if !IsUUIDEntity(tbl) {
			t.Errorf("IsUUIDEntity(%q) = false, want true (M1-migrated table)", tbl)
		}
	}

	alreadyUUIDSyncTables := []string{
		"carts",
		"cart_items",
		"sales_orders_v2",
		"sales_order_lines_v2",
		"draft_cart_templates",
		"notification_events",
	}
	for _, tbl := range alreadyUUIDSyncTables {
		if !IsUUIDEntity(tbl) {
			t.Errorf("IsUUIDEntity(%q) = false, want true (already-UUID sync table)", tbl)
		}
	}

	schemaUUIDDomainTables := []string{
		"draft_cart_template_items",
		"restaurant_carts",
		"restaurant_cart_items",
		"order_fulfillments",
		"order_fulfillment_items",
		"invoices",
		"invoice_lines",
		"invoice_payments",
		"quotes",
		"quote_lines",
		"user_notifications",
		"notification_deliveries",
		"customer_payments",
		"vendor_payments",
		"rfqs",
		"rfq_items",
		"rfq_responses",
		"rfq_response_items",
	}
	for _, tbl := range schemaUUIDDomainTables {
		if !IsUUIDEntity(tbl) {
			t.Errorf("IsUUIDEntity(%q) = false, want true (schema UUID domain table)", tbl)
		}
	}

	// Exact 5-of-32 split for PullEntityTypes()
	expectedUUIDPull := map[string]bool{
		"user_biometric_templates": true,
		"customers":                true,
		"stock_counts":             true,
		"stock_count_lines":        true,
		"stock_movements":          true,
	}
	uuidPullCount := 0
	intPullCount := 0
	for _, entity := range PullEntityTypes() {
		got := IsUUIDEntity(entity)
		want := expectedUUIDPull[entity]
		if got != want {
			t.Errorf("PullEntityTypes entry %q: IsUUIDEntity = %v, want %v", entity, got, want)
		}
		if got {
			uuidPullCount++
		} else {
			intPullCount++
		}
	}
	if uuidPullCount != 5 || intPullCount != 27 {
		t.Fatalf("PullEntityTypes UUID/int split = %d/%d, want 5/27", uuidPullCount, intPullCount)
	}

	// Exact 1-of-7 split for LeafDeleteEntityTypes()
	uuidLeafCount := 0
	intLeafCount := 0
	for _, leaf := range LeafDeleteEntityTypes() {
		got := IsUUIDEntity(leaf)
		want := leaf == "user_biometric_templates"
		if got != want {
			t.Errorf("LeafDeleteEntityTypes entry %q: IsUUIDEntity = %v, want %v", leaf, got, want)
		}
		if got {
			uuidLeafCount++
		} else {
			intLeafCount++
		}
	}
	if uuidLeafCount != 1 || intLeafCount != 6 {
		t.Fatalf("LeafDeleteEntityTypes UUID/int split = %d/%d, want 1/6", uuidLeafCount, intLeafCount)
	}

	// Non-UUID outbox/invalid entities
	nonUUID := []string{
		"zatca_device_configs",
		"zatca_document_chain",
		"",
		"CUSTOMERS",
		"unknown_table",
	}
	for _, tbl := range nonUUID {
		if IsUUIDEntity(tbl) {
			t.Errorf("IsUUIDEntity(%q) = true, want false", tbl)
		}
	}
}

func TestResolveEntityKey(t *testing.T) {
	tests := []struct {
		name      string
		entityKey string
		entityID  int64
		want      string
	}{
		{
			name:      "uuid string without entityID",
			entityKey: "550e8400-e29b-41d4-a716-446655440000",
			entityID:  0,
			want:      "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:      "whitespace-padded uuid string",
			entityKey: "  550e8400-e29b-41d4-a716-446655440000 \t\n",
			entityID:  0,
			want:      "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:      "entityKey overrides positive entityID",
			entityKey: "550e8400-e29b-41d4-a716-446655440000",
			entityID:  42,
			want:      "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:      "matching numeric entityKey and entityID",
			entityKey: "42",
			entityID:  42,
			want:      "42",
		},
		{
			name:      "empty entityKey falls back to positive entityID",
			entityKey: "",
			entityID:  42,
			want:      "42",
		},
		{
			name:      "whitespace-only entityKey falls back to positive entityID",
			entityKey: "  \t\n ",
			entityID:  42,
			want:      "42",
		},
		{
			name:      "empty entityKey and zero entityID",
			entityKey: "",
			entityID:  0,
			want:      "",
		},
		{
			name:      "whitespace entityKey and zero entityID",
			entityKey: "   ",
			entityID:  0,
			want:      "",
		},
		{
			name:      "empty entityKey and negative entityID",
			entityKey: "",
			entityID:  -1,
			want:      "",
		},
		{
			name:      "empty entityKey and MaxInt64 entityID",
			entityKey: "",
			entityID:  math.MaxInt64,
			want:      "9223372036854775807",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveEntityKey(tc.entityKey, tc.entityID); got != tc.want {
				t.Errorf("ResolveEntityKey(%q, %d) = %q, want %q", tc.entityKey, tc.entityID, got, tc.want)
			}
		})
	}
}

func TestParseNumericEntityID(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		wantID int64
		wantOK bool
	}{
		{name: "simple positive", key: "42", wantID: 42, wantOK: true},
		{name: "whitespace padded", key: "  42 \n", wantID: 42, wantOK: true},
		{name: "MaxInt64", key: "9223372036854775807", wantID: math.MaxInt64, wantOK: true},
		{name: "int64 overflow", key: "9223372036854775808", wantID: 0, wantOK: false},
		{name: "int64 underflow", key: "-9223372036854775809", wantID: 0, wantOK: false},
		{name: "zero", key: "0", wantID: 0, wantOK: false},
		{name: "negative", key: "-1", wantID: 0, wantOK: false},
		{name: "leading plus rejected", key: "+42", wantID: 0, wantOK: false},
		{name: "empty", key: "", wantID: 0, wantOK: false},
		{name: "whitespace only", key: "   ", wantID: 0, wantOK: false},
		{name: "uuid", key: "550e8400-e29b-41d4-a716-446655440000", wantID: 0, wantOK: false},
		{name: "alphanumeric", key: "12abc", wantID: 0, wantOK: false},
		{name: "decimal", key: "12.34", wantID: 0, wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotID, gotOK := ParseNumericEntityID(tc.key)
			if gotID != tc.wantID || gotOK != tc.wantOK {
				t.Errorf("ParseNumericEntityID(%q) = (%d, %v), want (%d, %v)", tc.key, gotID, gotOK, tc.wantID, tc.wantOK)
			}
		})
	}
}

func TestSyncProtoWireRoundTripAndBackwardCompat(t *testing.T) {
	if PullCursorAfterKey != "nembus-pull-after-key" {
		t.Fatalf("PullCursorAfterKey = %q, want %q", PullCursorAfterKey, "nembus-pull-after-key")
	}

	const uuidA = "550e8400-e29b-41d4-a716-446655440000"
	const uuidB = "7c9e6679-7425-40de-944b-e07fc1f90ae7"
	ts := timestamppb.New(time.Date(2026, 11, 21, 9, 30, 0, 0, time.UTC))

	t.Run("UUID round-trip across all contract messages", func(t *testing.T) {
		evIn := &SyncEvent{
			Id:            101,
			EntityType:    "pos_transactions",
			EntityId:      0,
			EntityKey:     uuidA,
			Action:        "INSERT",
			PayloadJson:   []byte(`{"id":"` + uuidA + `"}`),
			StoreId:       5,
			TenantSlug:    "tenant-1",
			EventTime:     ts,
			Sha256:        "deadbeef",
			IsLastChunk:   true,
			CorrelationId: "corr-1",
			OutboxUuid:    uuidB,
			DeviceId:      "POS-T01",
			BatchId:       "batch-1",
			BatchSize:     1,
			SeqInBatch:    1,
		}
		evBytes, err := proto.Marshal(evIn)
		if err != nil {
			t.Fatalf("Marshal SyncEvent: %v", err)
		}
		var evOut SyncEvent
		if err := proto.Unmarshal(evBytes, &evOut); err != nil {
			t.Fatalf("Unmarshal SyncEvent: %v", err)
		}
		if !proto.Equal(evIn, &evOut) {
			t.Fatalf("SyncEvent mismatch:\ngot:  %+v\nwant: %+v", &evOut, evIn)
		}
		if got := ResolveEntityKey(evOut.GetEntityKey(), evOut.GetEntityId()); got != uuidA {
			t.Fatalf("ResolveEntityKey(SyncEvent) = %q, want %q", got, uuidA)
		}

		ackIn := &SyncAck{
			Id:          101,
			EntityType:  "pos_transactions",
			EntityId:    0,
			EntityKey:   uuidA,
			Success:     true,
			Sha256:      "deadbeef",
			ProcessedAt: ts,
			BatchId:     "batch-1",
			OutboxUuid:  uuidB,
		}
		ackBytes, err := proto.Marshal(ackIn)
		if err != nil {
			t.Fatalf("Marshal SyncAck: %v", err)
		}
		var ackOut SyncAck
		if err := proto.Unmarshal(ackBytes, &ackOut); err != nil {
			t.Fatalf("Unmarshal SyncAck: %v", err)
		}
		if !proto.Equal(ackIn, &ackOut) {
			t.Fatalf("SyncAck mismatch")
		}
		if got := ResolveEntityKey(ackOut.GetEntityKey(), ackOut.GetEntityId()); got != uuidA {
			t.Fatalf("ResolveEntityKey(SyncAck) = %q, want %q", got, uuidA)
		}

		chgIn := &ChangeEvent{
			Seq:            9001,
			EntityType:     "customers",
			EntityId:       0,
			EntityKey:      uuidA,
			Op:             "UPDATE",
			PayloadJson:    []byte(`{"id":"` + uuidA + `"}`),
			Sha256:         "cafebabe",
			ChangedAt:      ts,
			IsSnapshotMode: true,
			SnapshotMaxSeq: 9100,
		}
		chgBytes, err := proto.Marshal(chgIn)
		if err != nil {
			t.Fatalf("Marshal ChangeEvent: %v", err)
		}
		var chgOut ChangeEvent
		if err := proto.Unmarshal(chgBytes, &chgOut); err != nil {
			t.Fatalf("Unmarshal ChangeEvent: %v", err)
		}
		if !proto.Equal(chgIn, &chgOut) {
			t.Fatalf("ChangeEvent mismatch")
		}
		if got := ResolveEntityKey(chgOut.GetEntityKey(), chgOut.GetEntityId()); got != uuidA {
			t.Fatalf("ResolveEntityKey(ChangeEvent) = %q, want %q", got, uuidA)
		}

		pullIn := &PullRequest{
			TenantSlug:  "tenant-1",
			StoreId:     5,
			Since:       ts,
			EntityTypes: []string{"customers"},
			Limit:       PullPageLimit,
			AfterKey:    uuidA,
		}
		pullBytes, err := proto.Marshal(pullIn)
		if err != nil {
			t.Fatalf("Marshal PullRequest: %v", err)
		}
		var pullOut PullRequest
		if err := proto.Unmarshal(pullBytes, &pullOut); err != nil {
			t.Fatalf("Unmarshal PullRequest: %v", err)
		}
		if !proto.Equal(pullIn, &pullOut) || pullOut.GetAfterKey() != uuidA {
			t.Fatalf("PullRequest mismatch: got AfterKey=%q", pullOut.GetAfterKey())
		}

		scIn := &StreamChangesRequest{
			TenantSlug: "tenant-1",
			StoreId:    5,
			FromSeq:    8999,
			Limit:      100,
			DeviceId:   "POS-T01",
		}
		scBytes, err := proto.Marshal(scIn)
		if err != nil {
			t.Fatalf("Marshal StreamChangesRequest: %v", err)
		}
		var scOut StreamChangesRequest
		if err := proto.Unmarshal(scBytes, &scOut); err != nil {
			t.Fatalf("Unmarshal StreamChangesRequest: %v", err)
		}
		if !proto.Equal(scIn, &scOut) {
			t.Fatalf("StreamChangesRequest mismatch")
		}

		cliIn := &SessionClientMsg{
			Msg: &SessionClientMsg_PushEvent{PushEvent: evIn},
		}
		cliBytes, err := proto.Marshal(cliIn)
		if err != nil {
			t.Fatalf("Marshal SessionClientMsg: %v", err)
		}
		var cliOut SessionClientMsg
		if err := proto.Unmarshal(cliBytes, &cliOut); err != nil {
			t.Fatalf("Unmarshal SessionClientMsg: %v", err)
		}
		if !proto.Equal(cliIn, &cliOut) || cliOut.GetPushEvent().GetEntityKey() != uuidA {
			t.Fatalf("SessionClientMsg mismatch")
		}

		srvIn := &SessionServerMsg{
			Msg: &SessionServerMsg_PushAck{PushAck: ackIn},
		}
		srvBytes, err := proto.Marshal(srvIn)
		if err != nil {
			t.Fatalf("Marshal SessionServerMsg: %v", err)
		}
		var srvOut SessionServerMsg
		if err := proto.Unmarshal(srvBytes, &srvOut); err != nil {
			t.Fatalf("Unmarshal SessionServerMsg: %v", err)
		}
		if !proto.Equal(srvIn, &srvOut) || srvOut.GetPushAck().GetEntityKey() != uuidA {
			t.Fatalf("SessionServerMsg mismatch")
		}
	})

	t.Run("mixed rolling upgrade integer EntityId plus EntityKey", func(t *testing.T) {
		ev := &SyncEvent{Id: 10, EntityType: "products", EntityId: 123, EntityKey: "123"}
		ack := &SyncAck{Id: 10, EntityType: "products", EntityId: 123, EntityKey: "123", Success: true}
		chg := &ChangeEvent{Seq: 55, EntityType: "products", EntityId: 123, EntityKey: "123", Op: "UPDATE"}

		b1, _ := proto.Marshal(ev)
		b2, _ := proto.Marshal(ack)
		b3, _ := proto.Marshal(chg)

		var ev2 SyncEvent
		var ack2 SyncAck
		var chg2 ChangeEvent
		if err := proto.Unmarshal(b1, &ev2); err != nil {
			t.Fatal(err)
		}
		if err := proto.Unmarshal(b2, &ack2); err != nil {
			t.Fatal(err)
		}
		if err := proto.Unmarshal(b3, &chg2); err != nil {
			t.Fatal(err)
		}

		for _, pair := range []struct {
			key string
			id  int64
		}{
			{ev2.GetEntityKey(), ev2.GetEntityId()},
			{ack2.GetEntityKey(), ack2.GetEntityId()},
			{chg2.GetEntityKey(), chg2.GetEntityId()},
		} {
			resolved := ResolveEntityKey(pair.key, pair.id)
			if resolved != "123" {
				t.Fatalf("resolved = %q, want \"123\"", resolved)
			}
			parsedID, ok := ParseNumericEntityID(resolved)
			if !ok || parsedID != 123 {
				t.Fatalf("ParseNumericEntityID(%q) = (%d, %v), want (123, true)", resolved, parsedID, ok)
			}
		}
	})

	t.Run("old-binary compatibility with empty EntityKey", func(t *testing.T) {
		evBytes, _ := proto.Marshal(&SyncEvent{Id: 1, EntityType: "products", EntityId: 77})
		ackBytes, _ := proto.Marshal(&SyncAck{Id: 1, EntityType: "products", EntityId: 77, Success: true})
		chgBytes, _ := proto.Marshal(&ChangeEvent{Seq: 9, EntityType: "products", EntityId: 77, Op: "INSERT"})

		var ev SyncEvent
		var ack SyncAck
		var chg ChangeEvent
		if err := proto.Unmarshal(evBytes, &ev); err != nil {
			t.Fatal(err)
		}
		if err := proto.Unmarshal(ackBytes, &ack); err != nil {
			t.Fatal(err)
		}
		if err := proto.Unmarshal(chgBytes, &chg); err != nil {
			t.Fatal(err)
		}

		if ev.GetEntityKey() != "" || ev.GetEntityId() != 77 || ResolveEntityKey(ev.GetEntityKey(), ev.GetEntityId()) != "77" {
			t.Fatalf("unexpected SyncEvent compat result: %+v", &ev)
		}
		if ack.GetEntityKey() != "" || ack.GetEntityId() != 77 || ResolveEntityKey(ack.GetEntityKey(), ack.GetEntityId()) != "77" {
			t.Fatalf("unexpected SyncAck compat result: %+v", &ack)
		}
		if chg.GetEntityKey() != "" || chg.GetEntityId() != 77 || ResolveEntityKey(chg.GetEntityKey(), chg.GetEntityId()) != "77" {
			t.Fatalf("unexpected ChangeEvent compat result: %+v", &chg)
		}
	})

	t.Run("protowire field tag verification and unknown-field stripping", func(t *testing.T) {
		verifyTag := func(name string, raw []byte, wantTag protowire.Number, wantVal string) []byte {
			t.Helper()
			found := false
			var stripped []byte
			rest := raw
			for len(rest) > 0 {
				num, typ, n := protowire.ConsumeTag(rest)
				if n < 0 {
					t.Fatalf("%s: ConsumeTag error: %v", name, protowire.ParseError(n))
				}
				valStart := n
				m := protowire.ConsumeFieldValue(num, typ, rest[n:])
				if m < 0 {
					t.Fatalf("%s: ConsumeFieldValue error: %v", name, protowire.ParseError(m))
				}
				fieldTotal := valStart + m
				if num == wantTag {
					if typ != protowire.BytesType {
						t.Fatalf("%s: tag %d wire type = %v, want BytesType", name, wantTag, typ)
					}
					s, sn := protowire.ConsumeString(rest[valStart:fieldTotal])
					if sn < 0 || s != wantVal {
						t.Fatalf("%s: tag %d string = %q (n=%d), want %q", name, wantTag, s, sn, wantVal)
					}
					found = true
				} else {
					stripped = append(stripped, rest[:fieldTotal]...)
				}
				rest = rest[fieldTotal:]
			}
			if !found {
				t.Fatalf("%s: expected wire field tag %d not found", name, wantTag)
			}
			return stripped
		}

		// SyncEvent tag 18
		evRaw, err := proto.Marshal(&SyncEvent{
			Id: 11, EntityType: "customers", EntityId: 77, OutboxUuid: uuidB, EntityKey: uuidA,
		})
		if err != nil {
			t.Fatal(err)
		}
		evStripped := verifyTag("SyncEvent", evRaw, 18, uuidA)
		var evLegacy SyncEvent
		if err := proto.Unmarshal(evStripped, &evLegacy); err != nil {
			t.Fatalf("Unmarshal stripped SyncEvent: %v", err)
		}
		if evLegacy.GetId() != 11 || evLegacy.GetEntityId() != 77 || evLegacy.GetOutboxUuid() != uuidB || evLegacy.GetEntityKey() != "" {
			t.Fatalf("stripped SyncEvent legacy fields mismatch: %+v", &evLegacy)
		}

		// SyncAck tag 12
		ackRaw, err := proto.Marshal(&SyncAck{
			Id: 11, EntityType: "customers", EntityId: 77, Success: true, OutboxUuid: uuidB, EntityKey: uuidA,
		})
		if err != nil {
			t.Fatal(err)
		}
		ackStripped := verifyTag("SyncAck", ackRaw, 12, uuidA)
		var ackLegacy SyncAck
		if err := proto.Unmarshal(ackStripped, &ackLegacy); err != nil {
			t.Fatalf("Unmarshal stripped SyncAck: %v", err)
		}
		if ackLegacy.GetId() != 11 || ackLegacy.GetEntityId() != 77 || !ackLegacy.GetSuccess() || ackLegacy.GetOutboxUuid() != uuidB || ackLegacy.GetEntityKey() != "" {
			t.Fatalf("stripped SyncAck legacy fields mismatch: %+v", &ackLegacy)
		}

		// PullRequest tag 6
		pullRaw, err := proto.Marshal(&PullRequest{
			TenantSlug: "tenant-1", StoreId: 3, Limit: 200, AfterKey: uuidA,
		})
		if err != nil {
			t.Fatal(err)
		}
		pullStripped := verifyTag("PullRequest", pullRaw, 6, uuidA)
		var pullLegacy PullRequest
		if err := proto.Unmarshal(pullStripped, &pullLegacy); err != nil {
			t.Fatalf("Unmarshal stripped PullRequest: %v", err)
		}
		if pullLegacy.GetTenantSlug() != "tenant-1" || pullLegacy.GetStoreId() != 3 || pullLegacy.GetLimit() != 200 || pullLegacy.GetAfterKey() != "" {
			t.Fatalf("stripped PullRequest legacy fields mismatch: %+v", &pullLegacy)
		}

		// ChangeEvent tag 10
		chgRaw, err := proto.Marshal(&ChangeEvent{
			Seq: 500, EntityType: "customers", EntityId: 77, Op: "INSERT", EntityKey: uuidA,
		})
		if err != nil {
			t.Fatal(err)
		}
		chgStripped := verifyTag("ChangeEvent", chgRaw, 10, uuidA)
		var chgLegacy ChangeEvent
		if err := proto.Unmarshal(chgStripped, &chgLegacy); err != nil {
			t.Fatalf("Unmarshal stripped ChangeEvent: %v", err)
		}
		if chgLegacy.GetSeq() != 500 || chgLegacy.GetEntityId() != 77 || chgLegacy.GetOp() != "INSERT" || chgLegacy.GetEntityKey() != "" {
			t.Fatalf("stripped ChangeEvent legacy fields mismatch: %+v", &chgLegacy)
		}

		// Raw protowire construction decoding into new structs
		var manual []byte
		manual = protowire.AppendTag(manual, 18, protowire.BytesType)
		manual = protowire.AppendString(manual, uuidA)
		var evManual SyncEvent
		if err := proto.Unmarshal(manual, &evManual); err != nil || evManual.GetEntityKey() != uuidA {
			t.Fatalf("manual tag 18 decode failed: err=%v key=%q", err, evManual.GetEntityKey())
		}
		if bytes.Equal(evRaw, evStripped) {
			t.Fatal("expected stripped SyncEvent bytes to differ from full SyncEvent bytes")
		}
	})
}
