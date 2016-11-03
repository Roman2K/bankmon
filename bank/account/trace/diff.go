package trace

import (
	"bytes"
	"time"

	"github.com/boltdb/bolt"

	"bankmon/bank/account"
)

func (t *Tracer) Diff() ([]*Diff, error) {
	return t.diff(func(c *bolt.Cursor, acc account.Account) (*Diff, error) {
		return doAccDiffPrev(c, acc, t.snapTime)
	})
}

func (t *Tracer) diff(do func(*bolt.Cursor, account.Account) (*Diff, error)) (diffs []*Diff, err error) {
	err = t.db.View(func(tx *bolt.Tx) error {
		balbkt := tx.Bucket(bucketBalances)
		for _, acc := range t.accs {
			c := balbkt.Bucket(accKey(acc)).Cursor()
			diff, err := do(c, acc)
			if err != nil {
				return err
			}
			if diff == nil || !diff.Any() {
				continue
			}
			diffs = append(diffs, diff)
		}
		return nil
	})
	return
}

func doAccDiffPrev(c *bolt.Cursor, acc account.Account, cur time.Time) (
	diff *Diff, err error,
) {
	curk := snapKey(cur)
	k, buf := c.Seek(curk)
	if k == nil {
		k, buf = c.Last()
	} else {
		// k >= curk
		k, buf = c.Prev()
	}
	if k == nil || bytes.Compare(k, curk) >= 0 {
		return
	}
	return newDiff(acc, k, buf)
}

func newDiff(acc account.Account, k, buf []byte) (diff *Diff, err error) {
	fromAcc, err := unmarshalAcc(buf)
	if err != nil {
		return
	}

	fromTime, err := time.ParseInLocation(snapKeyFmt, string(k), time.UTC)
	if err != nil {
		return
	}

	diff = &Diff{
		ATime: fromTime,
		A:     fromAcc,
		B:     acc,
	}
	return
}

type Diff struct {
	ATime time.Time
	A, B  account.Account
}

func (d Diff) Any() bool {
	if d.A == nil || d.B == nil {
		return false
	}
	return d.Balance() != 0 || d.AvailableBalance() != 0
}

func (d Diff) Balance() float64 {
	return d.B.Balance() - d.A.Balance()
}

func (d Diff) AvailableBalance() float64 {
	return d.B.AvailableBalance() - d.A.AvailableBalance()
}
