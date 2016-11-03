package throttle

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

var (
	bucketJobs = []byte("jobs")
)

type Throttle struct {
	db *bolt.DB
}

type Limit struct {
	Job    string
	Comply func(cur, prev time.Time) bool
}

func New(db *bolt.DB) Throttle {
	return Throttle{db: db}
}

func (t Throttle) Start(lim Limit) (Run, bool, error) {
	return t.StartAt(lim, time.Now())
}

func (t Throttle) StartAt(lim Limit, start time.Time) (run Run, ok bool, err error) {
	err = t.db.Update(func(tx *bolt.Tx) error {
		jobBkt, err := tx.CreateBucketIfNotExists(bucketJobs)
		if err != nil {
			return err
		}

		bktName := []byte(lim.Job)
		bkt, err := jobBkt.CreateBucketIfNotExists(bktName)
		if err != nil {
			return err
		}

		runk, err := runKey(bkt, start)
		if err != nil {
			return err
		}

		run = Run{
			db:  t.db,
			bkt: bktName,
			k:   runk,
		}

		can, err := canStart(bkt, lim, start)
		if !can || err != nil {
			ok = can
			return err
		}

		j := job{Start: start, Done: false}
		buf, err := json.Marshal(j)
		if err != nil {
			return err
		}

		err = bkt.Put(runk, buf)
		if err != nil {
			return err
		}

		ok = true
		return nil
	})
	return
}

func canStart(bkt *bolt.Bucket, lim Limit, start time.Time) (ok bool, err error) {
	k, buf := bkt.Cursor().Last()
	if k == nil {
		ok = true
		return
	}

	j := job{}
	err = json.Unmarshal(buf, &j)
	if err != nil {
		return
	}

	if !j.Done {
		ok = false
		return
	}

	ok = lim.Comply(start, j.End)
	return
}

func runKey(bkt *bolt.Bucket, t time.Time) (k []byte, err error) {
	n, err := bkt.NextSequence()
	if err != nil {
		return
	}
	k = []byte(fmt.Sprintf("%d-%d", t.UnixNano(), n))
	return
}

type job struct {
	Start time.Time
	End   time.Time
	Done  bool
}

type Run struct {
	db  *bolt.DB
	bkt []byte
	k   []byte
}

func (r Run) End() error {
	return r.EndAt(time.Now())
}

func (r Run) EndAt(end time.Time) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketJobs).Bucket(r.bkt)
		buf := bkt.Get(r.k)
		if buf == nil {
			return fmt.Errorf("no such job: %s", r.k)
		}

		j := job{}
		err := json.Unmarshal(buf, &j)
		if err != nil {
			return err
		}

		if j.Done {
			return fmt.Errorf("job %s already ended (at %s)", r.k, j.End)
		}

		j.End = end
		j.Done = true
		buf, err = json.Marshal(j)
		if err != nil {
			return err
		}

		return bkt.Put(r.k, buf)
	})
}
