package skimmer

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"net/http"
	"bytes"
	"time"
)


func getMemoryStorage() *MemoryStorage {
	return NewMemoryStorage(REQUEST_BODY_SIZE, BIN_LIFETIME)
}

func TestNewMemoryStorage(t *testing.T) {
	maxRequests := 20
	storage := NewMemoryStorage(maxRequests, BIN_LIFETIME)

	assert.Equal(t, storage.maxRequests, maxRequests)
	assert.NotNil(t, storage.binRecords)
}

func TestCreateBin(t *testing.T){
	storage := getMemoryStorage()
	bin := NewBin()
	err := storage.CreateBin(bin)
	if assert.Nil(t, err){
		assert.Equal(t, storage.binRecords[bin.Name].bin, bin)
	}
}

func TestUpdateBin(t *testing.T){
	storage := getMemoryStorage()
	bin := NewBin()
	if err := storage.CreateBin(bin); assert.Nil(t, err) {
		if err := storage.UpdateBin(bin); assert.Nil(t, err) {
			assert.Equal(t, storage.binRecords[bin.Name].bin, bin)
		}
	}
}

func TestGetBinRecord(t *testing.T) {
	storage := getMemoryStorage()

	binRecord, err := storage.getBinRecord("test")
	assert.NotNil(t, err)
	assert.Nil(t, binRecord)

	testBin := NewBin()
	testBin.Name = "test"
	err = storage.CreateBin(testBin)
	if assert.Nil(t, err){
		binRecord, err = storage.getBinRecord("test")
		if assert.Nil(t, err) {
			assert.Equal(t, binRecord.bin, testBin)
		}
	}
}

func TestLookupBin(t *testing.T) {
	storage := getMemoryStorage()

	_, err := storage.LookupBin("test")
	assert.NotNil(t, err)

	testBin := NewBin()
	testBin.Name = "test"
	err = storage.CreateBin(testBin)
	if assert.Nil(t, err){
		bin, err := storage.LookupBin("test")
		if assert.Nil(t, err) {
			assert.Equal(t, bin, testBin)
		}
	}
}

func TestLookupBins(t *testing.T) {
	storage := getMemoryStorage()

	bins, err := storage.LookupBins([]string{"test1", "test2"})
	if assert.Nil(t, err) {
		assert.Empty(t, bins)
	}

	testBin1 := NewBin()
	testBin1.Name = "test1"
	err = storage.CreateBin(testBin1)
	if assert.Nil(t, err){
		testBin2 := NewBin()
		testBin2.Name = "test2"
		err = storage.CreateBin(testBin2)
		if assert.Nil(t, err){
			bins, err = storage.LookupBins([]string{"test2", "test1"})
			if assert.Nil(t, err) {
				assert.Equal(t, bins, []*Bin{testBin2, testBin1})
			}
			bins, err = storage.LookupBins([]string{"test1"})
			if assert.Nil(t, err) {
				assert.Equal(t, bins, []*Bin{testBin1})
			}
		}
	}
}

func TestCreateRequest(t *testing.T) {
	storage := NewMemoryStorage(2, BIN_LIFETIME)
	bin := NewBin()
	storage.CreateBin(bin)
	httpRequest, _ := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("body")))
	request1 := NewRequest(httpRequest, REQUEST_BODY_SIZE)
	err := storage.CreateRequest(bin, request1)
	if assert.Nil(t, err) {

		binRecord, _ := storage.getBinRecord(bin.Name)
		assert.Equal(t, len(binRecord.requests), 1)
		assert.Equal(t, binRecord.requests[0], request1)
		assert.Equal(t, len(binRecord.requestMap), 1)
		assert.Equal(t, binRecord.requestMap[request1.Id], request1)
		assert.Equal(t, binRecord.bin.RequestCount, 1)

		request2 := NewRequest(httpRequest, REQUEST_BODY_SIZE)
		err = storage.CreateRequest(bin, request2)
		if assert.Nil(t, err) {
			assert.NotEqual(t, request1, request2)
			assert.Equal(t, len(binRecord.requests), 2)
			assert.Equal(t, binRecord.requests[1], request2)
			assert.Equal(t, len(binRecord.requestMap), 2)
			assert.Equal(t, binRecord.requestMap[request2.Id], request2)
			assert.Equal(t, binRecord.bin.RequestCount, 2)

			// shrinking
			request3 := NewRequest(httpRequest, REQUEST_BODY_SIZE)
			err = storage.CreateRequest(bin, request3)
			if assert.Nil(t, err) {
				assert.Equal(t, len(binRecord.requests), 2)
				assert.Equal(t, binRecord.requests[1], request3)
				assert.Equal(t, len(binRecord.requestMap), 2)
				assert.Equal(t, binRecord.requestMap[request3.Id], request3)
				assert.Equal(t, binRecord.bin.RequestCount, 2)
				_, err = storage.LookupRequest(bin.Name, request1.Id)
				assert.NotNil(t, err)
			}
		}
	}
	err = storage.CreateRequest(&Bin{Name:"wrong_name"}, request1)
	assert.NotNil(t, err)
}

func TestLookupRequest(t *testing.T) {
	storage := getMemoryStorage()
	bin := NewBin()
	storage.CreateBin(bin)
	httpRequest, _ := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("body")))
	request := NewRequest(httpRequest, REQUEST_BODY_SIZE)
	err := storage.CreateRequest(bin, request)
	if assert.Nil(t, err) {
		tmpRequest, err := storage.LookupRequest(bin.Name, request.Id)
		if assert.Nil(t, err) {
			assert.Equal(t, request, tmpRequest)
		}
	}
	_, err = storage.LookupRequest(bin.Name, "bad id")
	assert.NotNil(t, err)

	_, err = storage.LookupRequest("bad name", request.Id)
	assert.NotNil(t, err)
}

func TestLookupRequests(t *testing.T) {
	storage := getMemoryStorage()
	_, err := storage.LookupRequests("bad name", 0, 2)
	assert.NotNil(t, err)
	bin := NewBin()
	storage.CreateBin(bin)
	requests, err := storage.LookupRequests(bin.Name, 0, 3)
	if assert.Nil(t, err) {
		assert.Empty(t, requests)
	}

	httpRequest, _ := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("body")))
	request1 := NewRequest(httpRequest, REQUEST_BODY_SIZE)
	request2 := NewRequest(httpRequest, REQUEST_BODY_SIZE)
	request3 := NewRequest(httpRequest, REQUEST_BODY_SIZE)
	storage.CreateRequest(bin, request1)
	storage.CreateRequest(bin, request2)
	storage.CreateRequest(bin, request3)
	requests, err = storage.LookupRequests(bin.Name, 1, 3)
	if assert.Nil(t, err) {
		assert.Equal(t, requests[0], request2)
		assert.Equal(t, requests[1], request1)
	}
	requests, err = storage.LookupRequests(bin.Name, 0, 1)
	if assert.Nil(t, err) {
		assert.Equal(t, requests[0], request3)
	}
	requests, err = storage.LookupRequests(bin.Name, -1, 100)
	if assert.Nil(t, err) {
		assert.Equal(t, len(requests), 3)
	}
	requests, err = storage.LookupRequests(bin.Name, 1, -2)
	if assert.Nil(t, err) {
		assert.Equal(t, len(requests), 0)
	}

}

func TestMemoryClean(t *testing.T) {
	storage := NewMemoryStorage(2, -1)
	bin := NewBin()
	storage.CreateBin(bin)
	assert.Equal(t, storage.binRecords[bin.Name].bin, bin)
	storage.clean()
	assert.Equal(t, len(storage.binRecords), 0)

	storage.CreateBin(bin)
	assert.Equal(t, storage.binRecords[bin.Name].bin, bin)
	storage.StartCleaning(0)
	assert.Equal(t, len(storage.binRecords), 0)
	storage.CreateBin(bin)
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, len(storage.binRecords), 0)
	storage.StopCleaning()
}
