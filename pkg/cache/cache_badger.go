package cache

import (
	"log"

	"github.com/dgraph-io/badger"
)

type BadgerDBCache struct {
	db *badger.DB
}

func (b *BadgerDBCache) Set(k []byte, v []byte) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(k, v)
	})
	return err
}

func (b *BadgerDBCache) Get(k []byte) ([]byte, error) {
	var v []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(k)
		if err != nil {
			return err
		}
		item.Value(func(value []byte) error {
			v = value
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (b *BadgerDBCache) Delete(k []byte) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(k)
	})
	return err
}

func (b *BadgerDBCache) Close() error {
	return b.db.Close()
}

func newBadgerDBCache() (Cacher, error) {
	opts := badger.DefaultOptions
	opts.Dir = "./cache"
	opts.ValueDir = "./cache"
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}

	return &BadgerDBCache{db: db}, nil
}
