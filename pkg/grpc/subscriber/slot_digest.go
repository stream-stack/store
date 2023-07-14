package subscriber

import (
	"context"
	"github.com/dgraph-io/badger/v4"
	"github.com/linvon/cuckoo-filter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	v1 "github.com/stream-stack/common/cloudevents.io/genproto/v1"
	"github.com/stream-stack/common/util"
	"github.com/stream-stack/store/pkg/grpc"
	"github.com/stream-stack/store/pkg/store"
	"strings"
	"time"
)

const LocalSlotDigestOffsetKey = "LocalSlotDigestOffset"
const SlotDigestPrefix = "SlotDigest/"

var slotDigestMap = make(map[string]*slotDigester)

type slotDigester struct {
	slot   string
	filter *cuckoo.Filter
	op     chan func(f *cuckoo.Filter)
}

func (d *slotDigester) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case f := <-d.op:
			f(d.filter)
		}
	}
}

func (d *slotDigester) store() {
	if d.filter == nil {
		return
	}
	encode, err := d.filter.Encode()
	if err != nil {
		logrus.Errorf("[slotDigester]slot:%v encode err:%v", d.slot, err)
	}
	if err := store.KvStore.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(SlotDigestPrefix+d.slot), encode)
	}); err != nil {
		logrus.Errorf("[slotDigester]slot:%v store err:%v", d.slot, err)
	} else {
		logrus.Debugf("[slotDigester]slot:%v store success", d.slot)
	}
}

type slotDigestSubscriber struct {
	ctx               context.Context
	kvsvc             *grpc.KeyValueService
	evsvc             *grpc.EventService
	slotDigestMapOpCh chan func(m map[string]*slotDigester)
	offset            uint64
}

func (s *slotDigestSubscriber) Start(ctx context.Context) error {
	s.ctx = ctx
	offset, err := s.startOffsetManager(ctx)
	if err != nil {
		return err
	}

	s.startSlotDigestMapOp(ctx)
	//每一个槽位,一个布谷鸟过滤器
	//key: SlotDigest/slotId
	err = s.loadExistSlot()
	if err != nil {
		return err
	}
	//subscribe event
	go func() {
		err := s.evsvc.SubscribeWithHandler(s, func(options badger.IteratorOptions) badger.IteratorOptions {
			//offset, []byte(common.StoreTypeValuePrefix)
			options.Prefix = []byte(SlotDigestPrefix)
			options.SinceTs = offset + 1
			options.AllVersions = true
			return options
		}, nil)
		if err != nil {
			logrus.Errorf("[slot]subscribe err %v", err)
		}
	}()
	return nil
}

func (s *slotDigestSubscriber) loadExistSlot() error {
	if err := store.KvStore.View(func(txn *badger.Txn) error {
		options := badger.DefaultIteratorOptions
		options.Prefix = []byte(SlotDigestPrefix)
		iterator := txn.NewIterator(options)
		defer iterator.Close()
		for iterator.Rewind(); iterator.Valid(); iterator.Next() {
			item := iterator.Item()
			key := item.Key()
			slot, found := strings.CutPrefix(string(key), SlotDigestPrefix)
			if !found {
				logrus.Warnf("key %s not found prefix SlotDigest/", string(key))
				continue
			}
			return item.Value(func(val []byte) error {
				decode, err := cuckoo.Decode(val)
				if err != nil {
					logrus.Errorf("[slot]decode slot %s digest err %v", slot, err)
					return err
				}
				s.slotDigestMapOpCh <- func(m map[string]*slotDigester) {
					if _, found := m[slot]; !found {
						digester := newSlotDigester(slot, decode)
						m[slot] = digester
						digester.Start(s.ctx)
					}
				}
				return nil
			})
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func newSlotDigester(slot string, decode *cuckoo.Filter) *slotDigester {
	return &slotDigester{
		slot:   slot,
		filter: decode,
		op:     make(chan func(f *cuckoo.Filter), 1),
	}
}

func newSlotDigestSubscriber(evsvc *grpc.EventService, kvsvc *grpc.KeyValueService) *slotDigestSubscriber {
	return &slotDigestSubscriber{kvsvc: kvsvc, evsvc: evsvc}
}

func (s *slotDigestSubscriber) Context() context.Context {
	return s.ctx
}

func (s *slotDigestSubscriber) Handler(response *v1.CloudEventResponse) error {
	event := response.GetEvent()
	slot := response.Slot
	key := response.Key
	digester, ok := slotDigestMap[slot]
	if !ok {
		s.slotDigestMapOpCh <- func(m map[string]*slotDigester) {
			if exists, found := m[slot]; !found {
				digester = newSlotDigester(slot, nil)
				m[slot] = digester
				digester.Start(s.ctx)
			} else {
				digester = exists
			}
		}
	}
	var f func(f *cuckoo.Filter)
	if event == nil || len(event.GetId()) == 0 {
		f = func(f *cuckoo.Filter) {
			f.Add(key)
		}
	} else {
		f = func(f *cuckoo.Filter) {
			f.Delete(key)
		}
	}

	digester.op <- f
	return nil
}

func (s *slotDigestSubscriber) startSlotDigestMapOp(ctx context.Context) {
	s.slotDigestMapOpCh = make(chan func(m map[string]*slotDigester))
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case op := <-s.slotDigestMapOpCh:
				op(slotDigestMap)
			}
		}
	}()
}

func (s *slotDigestSubscriber) startOffsetManager(ctx context.Context) (uint64, error) {
	var offset uint64
	val, err := s.kvsvc.Get(ctx, &v1.GetRequest{Key: []byte(LocalSlotDigestOffsetKey)})
	if err != nil {
		logrus.Errorf("[slot]get LocalSlotDigestOffsetKey err %v", err)
		return 0, err
	}
	if val.GetValue() == nil {
		offset = 0
	} else {
		offset = util.BytesToUint64(val.GetValue())
	}
	interval := viper.GetDuration("SlotDigestOffsetStoreInterval")
	go func() {
		ticker := time.NewTicker(interval)
		prevOffset := s.offset
		go func() {
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if prevOffset == s.offset {
						logrus.Debugf("[grpc][digest]offset not change,skip save offset to store")
						continue
					}
					for _, digester := range slotDigestMap {
						digester.store()
					}
					//save offset to store
					_, err := s.kvsvc.Put(ctx, &v1.PutRequest{
						Key:   []byte(LocalSlotDigestOffsetKey),
						Value: util.Uint64ToBytes(s.offset),
					})
					if err != nil {
						logrus.Errorf("[grpc][digest]save offset to store error: %s", err.Error())
					} else {
						logrus.Debugf("[grpc][digest]save offset to store success,offset:%d", s.offset)
					}
					prevOffset = s.offset
				}
			}
		}()
	}()
	return offset, nil
}
