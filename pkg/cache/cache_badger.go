package cache

import (
	"log"
	"time"

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

func (b *BadgerDBCache) Setex(k []byte, ttl time.Duration, v []byte) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		return txn.SetWithTTL(k, v, ttl)
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

func (b *BadgerDBCache) Spop(prefix []byte, amount uint) ([][]byte, error) {
	var results [][]byte
	err := b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); len(results) <= int(amount) && it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			if len(k) > 0 {
				results = append(results, k)
			}

		}
		return nil
	})

	return results, err
}

func (b *BadgerDBCache) Delete(k []byte) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(k)
	})
	return err
}

func (b *BadgerDBCache) Close() {
	b.db.Close()
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
