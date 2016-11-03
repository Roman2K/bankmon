package trace

import (
	"encoding/json"
	"time"

	"github.com/boltdb/bolt"

	"bankmon/bank/account"
)

const (
	snapKeyFmt = "20060102150405"
)

var (
	bucketBalances = []byte("balances")
)

type Tracer struct {
	db       *bolt.DB
	accs     []account.Account
	snapTime time.Time
}

func NewTracer(db *bolt.DB, accs []account.Account) *Tracer {
	return &Tracer{
		db:       db,
		accs:     accs,
		snapTime: time.Now(),
	}
}

func (t *Tracer) SetSnapTime(time time.Time) {
	t.snapTime = time
}

func (t *Tracer) Snapshot() error {
	return t.db.Update(func(tx *bolt.Tx) error {
		balbkt, err := tx.CreateBucketIfNotExists(bucketBalances)
		if err != nil {
			return err
		}
		for _, acc := range t.accs {
			if bkt, err := balbkt.CreateBucketIfNotExists(accKey(acc)); err != nil {
				return err
			} else if buf, err := json.Marshal(account.Marshaler(acc)); err != nil {
				return err
			} else if err := bkt.Put(snapKey(t.snapTime), buf); err != nil {
				return err
			}
		}
		return nil
	})
}

func accKey(acc account.Account) []byte {
	return []byte(acc.IBAN())
}

func snapKey(t time.Time) []byte {
	return []byte(t.UTC().Format(snapKeyFmt))
}

func unmarshalAcc(buf []byte) (account.Account, error) {
	acc := account.Marshal{}
	err := json.Unmarshal(buf, &acc)
	return acc, err
}
