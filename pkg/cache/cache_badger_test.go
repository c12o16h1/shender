package cache

import (
	"testing"
	"bytes"
)

func TestNew(t *testing.T) {
	c, err := newBadgerDBCache()
	if err != nil {
		t.Fatalf("Can't create BadgerDB cache")
	}

	key := []byte("'/~`';testkey")
	val := []byte("testvalue")

	if err := c.Set(key, val); err != nil {
		t.Fatalf("Can't set key:val")
	}

	v, err := c.Get(key)
	if err != nil {
		t.Fatalf("Can't get val")
	}
	if !bytes.Equal(v, val) {
		t.Fatalf("Received value is wrong")
	}

	if err := c.Delete(key); err != nil {
		t.Fatalf("Can't delete key")
	}

	if err := c.Close(); err != nil {
		t.Fatalf("Can't close DB connection")
	}
}