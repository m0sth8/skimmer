package skimmer

import (
	"errors"
	"sync"
	"time"
)

type MemoryStorage struct {
	BaseStorage
	sync.RWMutex
	binRecords map[string]*BinRecord
	cleanTimer *time.Timer
}

type BinRecord struct {
	bin        *Bin
	requests   []*Request
	requestMap map[string]*Request
}

func (binRecord *BinRecord) ShrinkRequests(size int) {
	if size > 0 && len(binRecord.requests) > size {
		requests := binRecord.requests
		lenDiff := len(requests) - size
		removed := requests[:lenDiff]
		for _, removedReq := range removed {
			delete(binRecord.requestMap, removedReq.Id)
		}
		requests = requests[lenDiff:]
		binRecord.requests = requests
	}
}

func NewMemoryStorage(maxRequests int, binLifetime int64) *MemoryStorage {
	storage := &MemoryStorage{
		BaseStorage{
			maxRequests:        maxRequests,
			binLifetime:		binLifetime,
		},
		sync.RWMutex{},
		map[string]*BinRecord{},
		&time.Timer{},
	}
	return storage
}

func (storage *MemoryStorage) StartCleaning(timeout int) {
	defer func(){
		storage.cleanTimer = time.AfterFunc(time.Duration(timeout) * time.Second, func(){storage.StartCleaning(timeout)})
	}()
	storage.clean()
}

func (storage *MemoryStorage) StopCleaning() {
	if storage.cleanTimer != nil {
		storage.cleanTimer.Stop()
	}
}

func (storage *MemoryStorage) clean() {
	storage.Lock()
	defer storage.Unlock()
	now := time.Now().Unix()
	for name, binRecord := range storage.binRecords {
		if binRecord.bin.Updated < (now - storage.binLifetime) {
			delete(storage.binRecords, name)
		}
	}
}

func (storage *MemoryStorage) getBinRecord(name string) (*BinRecord, error) {
	storage.RLock()
	defer storage.RUnlock()
	if binRecord, ok := storage.binRecords[name]; ok {
		return binRecord, nil
	}
	return nil, errors.New("Bin not found")
}

func (storage *MemoryStorage) LookupBin(name string) (*Bin, error) {
	if binRecord, err := storage.getBinRecord(name); err == nil {
		return binRecord.bin, nil
	} else {
		return nil, err
	}
}

func (storage *MemoryStorage) LookupBins(names []string) ([]*Bin, error) {
	bins := []*Bin{}
	for _, name := range names {
		if binRecord, err := storage.getBinRecord(name); err == nil {
			bins = append(bins, binRecord.bin)
		}
	}
	return bins, nil
}

func (storage *MemoryStorage) CreateBin(bin *Bin) error {
	storage.Lock()
	defer storage.Unlock()
	binRec := BinRecord{bin, []*Request{}, map[string]*Request{}}
	storage.binRecords[bin.Name] = &binRec
	return nil
}

func (storage *MemoryStorage) UpdateBin(_ *Bin) error {
	return nil
}

func (storage *MemoryStorage) LookupRequest(binName, id string) (*Request, error) {
	if binRecord, err := storage.getBinRecord(binName); err == nil {
		if request, ok := binRecord.requestMap[id]; ok {
			return request, nil
		} else {
			return nil, errors.New("Request not found")
		}
	} else {
		return nil, err
	}
}

func (storage *MemoryStorage) LookupRequests(binName string, from int, to int) ([]*Request, error) {
	if binRecord, err := storage.getBinRecord(binName); err == nil {
		requestLen := len(binRecord.requests)
		if to >= requestLen {
			to = requestLen
		}
		if to < 0 {
			to = 0
		}
		if from < 0 {
			from = 0
		}
		if from > to {
			from = to
		}
		reversedLen := len(binRecord.requests)
		reversed := make([]*Request, reversedLen)
		for i, request := range binRecord.requests {
			reversed[reversedLen-i-1] = request
		}
		return reversed[from:to], nil
	} else {
		return nil, err
	}
}

func (storage *MemoryStorage) CreateRequest(bin *Bin, req *Request) error {
	if binRecord, err := storage.getBinRecord(bin.Name); err == nil {
		storage.Lock()
		defer storage.Unlock()
		binRecord.requests = append(binRecord.requests, req)
		binRecord.requestMap[req.Id] = req
		binRecord.ShrinkRequests(storage.maxRequests)
		binRecord.bin.RequestCount = len(binRecord.requests)
		binRecord.bin.Updated = time.Now().Unix()
		return nil
	} else {
		return err
	}
}
