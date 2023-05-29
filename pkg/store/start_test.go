package store

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/y"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"testing"
	"time"
)

func TestIter(t *testing.T) {
	options := badger.DefaultOptions("").WithInMemory(true)
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)

	options.Logger = l
	db, err := badger.Open(options)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(strconv.FormatInt(1, 16)), []byte("1"))
	})

	db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(strconv.FormatInt(2, 16)), []byte("2"))
	})

	db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(strconv.FormatInt(12, 16)), []byte("12"))
	})
	db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(strconv.FormatInt(22, 16)), []byte("22"))
	})

	db.View(func(txn *badger.Txn) error {
		iteratorOptions := badger.DefaultIteratorOptions
		//iteratorOptions.Reverse = true
		iterator := txn.NewIterator(iteratorOptions)
		defer iterator.Close()
		for iterator.Rewind(); iterator.Valid(); iterator.Next() {
			item := iterator.Item()
			if err := item.Value(func(val []byte) error {
				fmt.Printf("========%s\n", string(val))
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		}
		return nil
	})

}

func TestIterForSequence(t *testing.T) {
	options := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(options)
	if err != nil {
		panic(err)
	}
	defer func() { y.Check(db.Close()) }()

	bkey := func(i int) []byte {
		return []byte(fmt.Sprintf("%09d", i))
	}
	bval := func(i int) []byte {
		return []byte(fmt.Sprintf("%025d", i))
	}

	txn := db.NewTransaction(true)

	// Fill in 1000 items
	n := 1000
	for i := 0; i < n; i++ {
		err := txn.SetEntry(badger.NewEntry(bkey(i), bval(i)))
		if err != nil {
			panic(err)
		}
	}

	if err := txn.Commit(); err != nil {
		panic(err)
	}

	opt := badger.DefaultIteratorOptions
	opt.PrefetchSize = 10

	// Iterate over 1000 items
	var count int
	err = db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(opt)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			fmt.Printf("%s \n", string(it.Item().Key()))
			it.Item().Value(func(val []byte) error {
				fmt.Printf("value:%s \n", string(val))
				return nil
			})
			count++
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Counted %d elements", count)

}

func TestIterForTimestamp(t *testing.T) {
	options := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(options)
	if err != nil {
		panic(err)
	}
	defer func() { y.Check(db.Close()) }()

	txn := db.NewTransaction(true)
	err = txn.SetEntry(badger.NewEntry([]byte(`CLOUDEVENT/abc`), []byte("干扰选项")))
	if err != nil {
		panic(err)
	}

	// Fill in 1000 items
	n := 1000
	for i := 0; i < n; i++ {
		now := timestamppb.Now()
		key := fmt.Sprintf(`%s/%s/%025d`, "CLOUDEVENT", "abcd", now.AsTime().UnixNano())
		fmt.Println("key:", key)
		err := txn.SetEntry(badger.NewEntry([]byte(key), []byte(key)))
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Millisecond)
	}

	if err := txn.Commit(); err != nil {
		panic(err)
	}

	opt := badger.DefaultIteratorOptions
	opt.PrefetchSize = 10
	prefix := []byte(fmt.Sprintf("%s/%s", "CLOUDEVENT", "abcd"))
	opt.Prefix = prefix

	// Iterate over 1000 items
	var count int
	err = db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(opt)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			fmt.Printf("%s \n", string(it.Item().Key()))
			//it.Item().Value(func(val []byte) error {
			//	fmt.Printf("value:%s \n", string(val))
			//	return nil
			//})
			count++
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Counted %d elements", count)

}
