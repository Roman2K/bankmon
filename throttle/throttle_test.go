package throttle_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"bankmon/testutil"
	"bankmon/throttle"
)

func TestThrottling(t *testing.T) {
	db, err := testutil.TempBolt()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(db.Path())

	throt := throttle.New(db)
	curTime := time.Now()
	var lim throttle.Limit

	start := func() (r throttle.Run, ok bool) {
		r, ok, err := throt.StartAt(lim, curTime)
		if err != nil {
			t.Fatal(err)
		}
		return
	}

	end := func(r throttle.Run) {
		if err := r.EndAt(curTime); err != nil {
			t.Fatal(err)
		}
	}

	assertStart := func() (r throttle.Run) {
		r, ok := start()
		if !ok {
			t.Fatal("should have complied with limit")
		}
		return
	}

	assertNoStart := func() (r throttle.Run) {
		r, ok := start()
		if ok {
			t.Fatal("should not have complied with limit")
		}
		return
	}

	// Not started

	lim = throttle.Limit{
		Job: "job0",
		Comply: func(cur, prev time.Time) bool {
			return true
		},
	}
	assertStart()

	// Already started

	lim = throttle.Limit{
		Job: "job1",
		Comply: func(cur, prev time.Time) bool {
			return true
		},
	}
	assertStart()
	assertNoStart()

	// Consecutive runs

	lim = throttle.Limit{
		Job: "job2",
		Comply: func(cur, prev time.Time) bool {
			return true
		},
	}

	run := assertStart()
	end(run)
	assertStart()

	// Once per second

	lim = throttle.Limit{
		Job: "job3",
		Comply: func(cur, prev time.Time) bool {
			return cur.Unix() > prev.Unix()
		},
	}

	run = assertStart()
	end(run)
	assertNoStart()
	assertNoStart()
	curTime = curTime.Add(1 * time.Second)
	run = assertStart()
	end(run)
	assertNoStart()

	// Already ended

	lim = throttle.Limit{
		Job: "job4",
		Comply: func(cur, prev time.Time) bool {
			return cur.Unix() > prev.Unix()
		},
	}

	run = assertStart()
	end(run)
	err = run.End()
	if err == nil || !strings.Contains(err.Error(), "already ended") {
		t.Fatal(`expected to return "already ended" error`)
	}
}
