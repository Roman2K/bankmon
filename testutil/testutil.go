package testutil

import (
	"io/ioutil"

	"github.com/boltdb/bolt"
)

func TempBolt() (db *bolt.DB, err error) {
	f, err := ioutil.TempFile("", "TempBolt")
	if err != nil {
		return
	}
	f.Close()
	return bolt.Open(f.Name(), 0600, nil)
}
