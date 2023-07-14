package grpc

import (
	"github.com/dgraph-io/badger/v4"
	"testing"
)

func TestEventService_Delete(t *testing.T) {
	store, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("test"), []byte("test"))
	}); err != nil {
		t.Fatal(err)
	}
	currentVersion := store.MaxVersion()
	t.Logf("current version:%v", currentVersion)
	//iterator key value
	if err := store.View(func(txn *badger.Txn) error {
		options := badger.DefaultIteratorOptions
		options.SinceTs = 0
		iterator := txn.NewIterator(options)
		defer iterator.Close()
		for iterator.Rewind(); iterator.Valid(); iterator.Next() {
			item := iterator.Item()
			return item.Value(func(val []byte) error {
				t.Logf("key:%v value:%v version:%v", string(item.Key()), string(val), item.Version())
				return nil
			})
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	//delete key
	t.Logf("delete key:%v", "test")
	if err := store.Update(func(txn *badger.Txn) error {
		//return txn.Set([]byte("test"), []byte("test1"))
		return txn.Delete([]byte("test"))
	}); err != nil {
		t.Fatal(err)
	}
	t.Logf("current version:%v", store.MaxVersion())
	//iterator key value
	if err := store.View(func(txn *badger.Txn) error {
		options := badger.DefaultIteratorOptions
		options.SinceTs = 0
		options.AllVersions = true
		iterator := txn.NewIterator(options)
		defer iterator.Close()
		for iterator.Rewind(); iterator.Valid(); iterator.Next() {
			item := iterator.Item()
			return item.Value(func(val []byte) error {
				t.Logf("key:%v value:%v version:%v", string(item.Key()), string(val), item.Version())
				return nil
			})
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
