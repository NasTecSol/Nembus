package usecase

import "testing"

func TestHashSyncTokenDeterministic(t *testing.T) {
	raw := "abcd1234"
	if HashSyncToken(raw) != HashSyncToken(raw) {
		t.Fatal("HashSyncToken must be deterministic")
	}
	if HashSyncToken(" "+raw+" ") != HashSyncToken(raw) {
		t.Fatal("HashSyncToken should trim surrounding whitespace")
	}
}

func TestGenerateSyncTokenUnique(t *testing.T) {
	r1, h1, err := GenerateSyncToken()
	if err != nil {
		t.Fatalf("GenerateSyncToken: %v", err)
	}
	r2, h2, err := GenerateSyncToken()
	if err != nil {
		t.Fatalf("GenerateSyncToken: %v", err)
	}
	if r1 == r2 || h1 == h2 {
		t.Fatal("expected unique tokens")
	}
}
