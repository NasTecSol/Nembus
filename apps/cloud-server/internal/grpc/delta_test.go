package grpc

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/syncpb"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type deltaCaptureStream struct {
	grpc.ServerStream
	ctx    context.Context
	events []*syncpb.SyncEvent
}

func (s *deltaCaptureStream) Context() context.Context { return s.ctx }
func (s *deltaCaptureStream) Send(event *syncpb.SyncEvent) error {
	s.events = append(s.events, event)
	return nil
}

func TestBoundedDeltaPagesWithTimestampTies(t *testing.T) {
	dsn := os.Getenv("NEMBUS_SYNC_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set NEMBUS_SYNC_TEST_DATABASE_URL for PostgreSQL tests")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := fmt.Sprintf("delta_page_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE") }()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	_, err = pool.Exec(ctx, `CREATE TABLE promotions(id bigint PRIMARY KEY,updated_at timestamp);
		INSERT INTO promotions SELECT id,'2026-10-10'::timestamp FROM generate_series(1,450) id`)
	if err != nil {
		t.Fatal(err)
	}
	since := time.Unix(0, 0)
	var afterID int64
	server := &SyncServer{}
	for _, expectedCount := range []int{200, 200, 50, 0} {
		pageCtx := metadata.NewIncomingContext(ctx, metadata.Pairs(syncpb.PullCursorKey, strconv.FormatInt(afterID, 10)))
		stream := &deltaCaptureStream{ctx: pageCtx}
		if err := server.streamEntityDeltas(pageCtx, pool, &syncpb.PullRequest{}, "promotions", since, 200, stream); err != nil {
			t.Fatal(err)
		}
		if len(stream.events) != expectedCount {
			t.Fatalf("page length=%d want=%d", len(stream.events), expectedCount)
		}
		for _, event := range stream.events {
			if event.EntityId != afterID+1 {
				t.Fatalf("duplicate or skipped ID: got=%d after=%d", event.EntityId, afterID)
			}
			afterID = event.EntityId
			since = event.EventTime.AsTime()
		}
	}
	if afterID != 450 {
		t.Fatalf("only fetched through ID %d", afterID)
	}
}
