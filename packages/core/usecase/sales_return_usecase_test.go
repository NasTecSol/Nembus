package usecase

import (
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestGenerateReturnNumberCollisionSafe(t *testing.T) {
	now := time.Now()
	generated := make(map[string]bool, 1000)

	for i := 0; i < 1000; i++ {
		num, err := GenerateReturnNumber(now)
		if err != nil {
			t.Fatalf("unexpected error generating return number: %v", err)
		}

		// Verify prefix
		if !strings.HasPrefix(num, "RET-") {
			t.Fatalf("expected prefix 'RET-', got %s", num)
		}

		// Verify column limit compliance (VARCHAR(50))
		if len(num) > 50 {
			t.Fatalf("return number exceeds column length 50: %s (len %d)", num, len(num))
		}
		if len(num) < 20 {
			t.Fatalf("return number too short: %s (len %d)", num, len(num))
		}

		// Verify uniqueness under identical timestamp
		if generated[num] {
			t.Fatalf("collision detected for return number: %s at iteration %d", num, i)
		}
		generated[num] = true
	}
}

func TestNumericToRatExactDecimal(t *testing.T) {
	// 12.345 -> Int: 12345, Exp: -3
	n := pgtype.Numeric{
		Int:   big.NewInt(12345),
		Exp:   -3,
		Valid: true,
	}
	r := numericToRat(n)
	expected := big.NewRat(12345, 1000)
	if r.Cmp(expected) != 0 {
		t.Fatalf("expected %v, got %v", expected, r)
	}

	// Negative value
	nNeg := pgtype.Numeric{
		Int:   big.NewInt(-500),
		Exp:   -2,
		Valid: true,
	}
	rNeg := numericToRat(nNeg)
	if rNeg.Sign() >= 0 {
		t.Fatalf("expected negative rational, got %v", rNeg)
	}

	// Invalid numeric
	nInv := pgtype.Numeric{Valid: false}
	rInv := numericToRat(nInv)
	if rInv.Sign() != 0 {
		t.Fatalf("expected 0 for invalid numeric, got %v", rInv)
	}

	// Fallback parseQuantity
	rStr := parseQuantity(pgtype.Numeric{Valid: false}, "7.250")
	expectedStr := big.NewRat(725, 100)
	if rStr.Cmp(expectedStr) != 0 {
		t.Fatalf("expected %v, got %v", expectedStr, rStr)
	}
}
