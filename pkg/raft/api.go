package raft

import (
	"github.com/dgraph-io/badger/v3"
	hraft "github.com/hashicorp/raft"
)

func GetLogByKey(key []byte) ([]byte, uint64, error) {
	//Store.conn.View(func(txn *badger.Txn) error {
	//	iterator := txn.NewIterator(badger.IteratorOptions{})
	//	defer iterator.Close()
	//	for iterator.Rewind(); iterator.Valid(); iterator.Next() {
	//		item := iterator.Item()
	//		fmt.Printf("key:%s|", item.Key())
	//		item.Value(func(val []byte) error {
	//			fmt.Printf(" value:%s \n", val)
	//			return nil
	//		})
	//	}
	//	return nil
	//})
	var logKey uint64
	if err := Store.conn.View(func(txn *badger.Txn) error {
		var value []byte
		item, err := txn.Get(key)
		if err != nil {
			switch err {
			case badger.ErrKeyNotFound:
				return ErrKeyNotFound
			default:
				return err
			}
		}
		value, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		logKey = bytesToUint64(value)
		return nil
	}); err != nil {
		return nil, 0, err
	}
	log := &hraft.Log{}
	if err := Store.GetLog(logKey, log); err != nil {
		return nil, 0, err
	}
	return log.Data, log.Index, nil
}

func Iterator(prefix []byte, slice uint64,
	filter func(log *hraft.Log) (bool, error),
	handler func(key, value []byte, version uint64) error) error {
	return Store.conn.View(func(txn *badger.Txn) error {
		options := badger.IteratorOptions{
			SinceTs: slice,
		}
		if len(prefix) > 0 {
			options.Prefix = prefix
		}
		iterator := txn.NewIterator(options)
		defer iterator.Close()
		for iterator.Rewind(); iterator.Valid(); iterator.Next() {
			item := iterator.Item()
			version := item.Version()
			log := &hraft.Log{}
			log.Extensions = item.Key()
			b, err := filter(log)
			if err != nil {
				return err
			}
			if !b {
				continue
			}
			key := item.Key()
			offset, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			item, err = txn.Get(append(prefixLogs, offset...))
			if err != nil {
				return err
			}
			fn := func(val []byte) error {
				return decodeMsgPack(val, log)
			}
			if err = item.Value(fn); err != nil {
				return err
			}

			if err = handler(key, log.Data, version); err != nil {
				return err
			}
		}
		return nil
	})
}
