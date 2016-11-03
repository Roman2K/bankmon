package trace_test

import (
	"os"
	"testing"
	"time"

	"bankmon/bank/account"
	"bankmon/bank/account/trace"
	"bankmon/testutil"
)

func TestDiff(t *testing.T) {
	testDiff(t, false)
	testDiff(t, true)
}

func testDiff(t *testing.T, lastSnapshot bool) {
	db, err := testutil.TempBolt()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(db.Path())

	acc := testAcc{
		iban:         "acc1",
		balance:      1,
		availBalance: 1,
	}
	now := time.Now()

	tracer := trace.NewTracer(db, []account.Account{acc})
	tracer.SetSnapTime(now.Add(-10 * time.Second))
	err = tracer.Snapshot()
	if err != nil {
		panic(err)
	}

	acc.balance -= 9
	tracer = trace.NewTracer(db, []account.Account{acc})
	tracer.SetSnapTime(now.Add(-5 * time.Second))
	if lastSnapshot {
		err := tracer.Snapshot()
		if err != nil {
			panic(err)
		}
	}
	diffs, err := tracer.Diff()
	if err != nil {
		panic(err)
	}
	if want, got := 1, len(diffs); got != want {
		t.Fatalf("diffs: want %d, got %d", want, got)
	}
	d := diffs[0]
	if want, got := float64(0.0), d.AvailableBalance(); got != want {
		t.Fatalf("AvailableBalance: want %f, got %f", want, got)
	}
	if want, got := float64(-9.0), d.Balance(); got != want {
		t.Fatalf("Balance: want %f, got %f", want, got)
	}
}

type testAcc struct {
	iban         string
	balance      float64
	availBalance float64
}

func (a testAcc) IBAN() string              { return a.iban }
func (a testAcc) Name() string              { return "Test" }
func (a testAcc) Currency() string          { return "EUR" }
func (a testAcc) Balance() float64          { return a.balance }
func (a testAcc) AvailableBalance() float64 { return a.availBalance }
