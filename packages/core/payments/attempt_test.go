package payments

import (
	"testing"

	"github.com/google/uuid"
)

func validAttempt() *Attempt {
	return &Attempt{
		ID: uuid.New(), IdempotencyKey: "k1", AttemptType: TypeSale, StoreID: 1,
		InvoiceRef: "INV-1", AmountMinor: 10050, Currency: "SAR", Status: StatusPending, Provider: "fake",
	}
}

func TestValidate(t *testing.T) {
	if err := validAttempt().Validate(); err != nil {
		t.Fatal(err)
	}
	mut := map[string]func(*Attempt){
		"nil id":        func(a *Attempt) { a.ID = uuid.Nil },
		"no key":        func(a *Attempt) { a.IdempotencyKey = " " },
		"zero amount":   func(a *Attempt) { a.AmountMinor = 0 },
		"negative":      func(a *Attempt) { a.AmountMinor = -5 },
		"wrong ccy":     func(a *Attempt) { a.Currency = "USD" },
		"bad status":    func(a *Attempt) { a.Status = "paid" },
		"no provider":   func(a *Attempt) { a.Provider = "" },
		"no invoice":    func(a *Attempt) { a.InvoiceRef = "" },
		"no store":      func(a *Attempt) { a.StoreID = 0 },
		"refund no ref": func(a *Attempt) { a.AttemptType = TypeRefund },
		"approved without refs": func(a *Attempt) {
			a.Status = StatusSucceeded
		},
	}
	for name, f := range mut {
		a := validAttempt()
		f(a)
		if a.Validate() == nil {
			t.Errorf("%s: expected validation error", name)
		}
	}
	a := validAttempt()
	a.Status = StatusSucceeded
	a.ProviderRefs = map[string]string{"x": "1"}
	if err := a.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestCanTransition(t *testing.T) {
	ok := [][2]Status{
		{StatusPending, StatusProcessing}, {StatusPending, StatusCancelled},
		{StatusProcessing, StatusSucceeded}, {StatusProcessing, StatusDeclined},
		{StatusProcessing, StatusUnknown}, {StatusProcessing, StatusCancelled},
		{StatusUnknown, StatusSucceeded}, {StatusUnknown, StatusDeclined}, {StatusUnknown, StatusCancelled},
	}
	for _, p := range ok {
		if !CanTransition(p[0], p[1]) {
			t.Errorf("%s -> %s should be allowed", p[0], p[1])
		}
	}
	bad := [][2]Status{
		{StatusPending, StatusSucceeded}, // cannot approve without contacting terminal
		{StatusPending, StatusUnknown},
		{StatusSucceeded, StatusDeclined}, {StatusSucceeded, StatusProcessing}, {StatusDeclined, StatusSucceeded},
		{StatusCancelled, StatusSucceeded}, {StatusUnknown, StatusProcessing}, {StatusUnknown, StatusPending},
		{StatusProcessing, StatusPending},
	}
	for _, p := range bad {
		if CanTransition(p[0], p[1]) {
			t.Errorf("%s -> %s must be rejected", p[0], p[1])
		}
	}
}

func TestMerge(t *testing.T) {
	cases := []struct {
		cur, in Status
		want    Decision
	}{
		{StatusPending, StatusProcessing, Apply},
		{StatusPending, StatusSucceeded, Apply}, // skipped states tolerated
		{StatusProcessing, StatusUnknown, Apply},
		{StatusUnknown, StatusSucceeded, Apply},
		{StatusSucceeded, StatusSucceeded, Ignore}, // duplicate
		{StatusSucceeded, StatusProcessing, Ignore}, // stale
		{StatusSucceeded, StatusUnknown, Ignore},
		{StatusProcessing, StatusPending, Ignore},
		{StatusUnknown, StatusProcessing, Ignore},
		{StatusSucceeded, StatusDeclined, Conflict}, // never regress/replace final
		{StatusDeclined, StatusSucceeded, Conflict},
		{StatusCancelled, StatusSucceeded, Conflict},
	}
	for _, c := range cases {
		if got := Merge(c.cur, c.in); got != c.want {
			t.Errorf("Merge(%s,%s)=%v want %v", c.cur, c.in, got, c.want)
		}
	}
}

func TestImmutableMatches(t *testing.T) {
	a, b := validAttempt(), validAttempt()
	b.ID = a.ID
	if err := ImmutableMatches(a, b); err != nil {
		t.Fatal(err)
	}
	for name, f := range map[string]func(*Attempt){
		"amount":   func(x *Attempt) { x.AmountMinor++ },
		"currency": func(x *Attempt) { x.Currency = "AED" },
		"invoice":  func(x *Attempt) { x.InvoiceRef = "other" },
		"store":    func(x *Attempt) { x.StoreID = 2 },
		"key":      func(x *Attempt) { x.IdempotencyKey = "k2" },
		"provider": func(x *Attempt) { x.Provider = "p" },
		"type":     func(x *Attempt) { x.AttemptType = TypeRefund },
	} {
		c := *a
		f(&c)
		if ImmutableMatches(a, &c) == nil {
			t.Errorf("%s change must be detected", name)
		}
	}
}

func TestMoney(t *testing.T) {
	good := map[string]int64{"0": 0, "1": 100, "12.5": 1250, "12.50": 1250, "100.05": 10005, "0.07": 7}
	for in, want := range good {
		got, err := DecimalToMinor(in)
		if err != nil || got != want {
			t.Errorf("DecimalToMinor(%q)=%d,%v want %d", in, got, err, want)
		}
	}
	for _, in := range []string{"", "-1", "+1", "1.", ".5", "1.234", "abc", "1,5", "1e3", "99999999999999999999"} {
		if _, err := DecimalToMinor(in); err == nil {
			t.Errorf("DecimalToMinor(%q) should fail", in)
		}
	}
	for minor, want := range map[int64]string{0: "0.00", 7: "0.07", 10050: "100.50", 100: "1.00"} {
		if got := MinorToDecimal(minor); got != want {
			t.Errorf("MinorToDecimal(%d)=%s want %s", minor, got, want)
		}
	}
	// the classic float trap: 19.99 * 100 = 1998.9999...
	if got, _ := DecimalToMinor("19.99"); got != 1999 {
		t.Errorf("19.99 -> %d", got)
	}
}

func TestPolicy(t *testing.T) {
	p := DefaultPolicy()
	if !p.RequiresAttempt("card", "GEIDEA") || p.RequiresAttempt("cash", "") || p.RequiresAttempt("card", "stripe") {
		t.Fatal("default policy wrong")
	}
	p.EnforcedMethods = []string{"card", "mada"}
	if !p.RequiresAttempt("MADA", "") || p.RequiresAttempt("cash", "") {
		t.Fatal("method enforcement wrong")
	}
	if len(ParseList(" a, ,b ")) != 2 {
		t.Fatal("ParseList")
	}
}
