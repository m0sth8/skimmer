package skimmer

import (
	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"

//	"github.com/davecgh/go-spew/spew"
	"bytes"
	"net/http"
	"errors"
	"fmt"
	"time"
)

type MockedRedisConn struct {
	mock.Mock
}

func (c *MockedRedisConn) Close() error {
	args := c.Mock.Called()
	return args.Error(0)
}

func (c *MockedRedisConn) Err() error {
	args := c.Mock.Called()
	return args.Error(0)
}

func (c *MockedRedisConn) Do(commandName string, params ...interface{}) (interface{}, error) {
	//	spew.Dump(commandName, params)
	args := c.Mock.Called(commandName, params)
	return args.Get(0), args.Error(1)
}

func (c *MockedRedisConn) Send(commandName string, params ...interface{}) error {
//	spew.Dump(commandName, params)
	args := c.Mock.Called(commandName, params)
	return args.Error(0)
}

func (c *MockedRedisConn) Flush() error {
	args := c.Mock.Called()
	return args.Error(0)
}

func (c *MockedRedisConn) Receive() (interface{}, error) {
	args := c.Mock.Called()
	return args.Get(0), args.Error(1)
}

func getRedisStorage(conn redis.Conn) *RedisStorage {
	pool := &redis.Pool{
		MaxIdle: 3,
		Dial: func() (redis.Conn, error) {
			return conn, nil
		},
	}

	return &RedisStorage{
		BaseStorage{MAX_REQUEST_COUNT, BIN_LIFETIME},
		pool,
		"skimmer",
		&time.Timer{},
	}
}

func TestRedisCreateBin(t *testing.T) {
	conn := &MockedRedisConn{}
	storage := getRedisStorage(conn)
	bin := NewBin()
	if binDump, err := storage.Dump(bin); assert.Nil(t, err) {
		conn.On("Send", "SET", []interface{}{
			storage.getKey(BIN_KEY, bin.Name),
			binDump,
		},
		).Return(nil).Once()
		conn.On("Send", "EXPIRE", []interface{}{
			storage.getKey(BIN_KEY, bin.Name),
			BIN_LIFETIME,
		},
		).Return(nil).Once()
		conn.On("Flush").Return(nil).Once()
		conn.On("Do", "", []interface{}(nil)).Return(0, nil).Once()
		conn.On("Err").Return(nil).Once()
		err := storage.CreateBin(bin)
		conn.Mock.AssertExpectations(t)
		assert.Nil(t, err)
	}
}

func TestRedisUpdateBin(t *testing.T) {
	conn := &MockedRedisConn{}
	storage := getRedisStorage(conn)
	bin := NewBin()
	if binDump, err := storage.Dump(bin); assert.Nil(t, err) {
		conn.On("Send", "SET", []interface{}{
			storage.getKey(BIN_KEY, bin.Name),
			binDump,
		},
		).Return(nil).Once()
		conn.On("Send", "EXPIRE", []interface{}{
			storage.getKey(BIN_KEY, bin.Name),
			BIN_LIFETIME,
		},
		).Return(nil).Once()
		conn.On("Flush").Return(nil).Once()
		conn.On("Do", "", []interface{}(nil)).Return(0, nil).Once()
		conn.On("Err").Return(nil).Once()
		err := storage.UpdateBin(bin)
		conn.Mock.AssertExpectations(t)
		assert.Nil(t, err)

		// error
		conn.On("Send", "SET", []interface{}{
			storage.getKey(BIN_KEY, bin.Name),
			binDump,
		},
		).Return(errors.New("error")).Once()
		conn.On("Do", "", []interface{}(nil)).Return(0, nil).Once()
		conn.On("Err").Return(nil).Once()
		err = storage.UpdateBin(bin)
		conn.Mock.AssertExpectations(t)
		assert.NotNil(t, err)
	}

}

func TestRedisLookupBin(t *testing.T) {
	conn := &MockedRedisConn{}
	storage := getRedisStorage(conn)
	bin := NewBin()
	if binDump, err := storage.Dump(bin); assert.Nil(t, err) {
		conn.On("Do", "GET", []interface{}{storage.getKey(BIN_KEY, bin.Name)}).Return(binDump, nil).Once()
		conn.On("Do", "", []interface{}(nil)).Return(0, nil).Once()
		conn.On("Err").Return(nil).Once()
		bin2, err := storage.LookupBin(bin.Name)
		conn.Mock.AssertExpectations(t)
		if assert.Nil(t, err) {
			assert.Equal(t, bin2, bin)
		}

		// error
		conn.On("Do", "GET", []interface{}{storage.getKey(BIN_KEY, bin.Name)}).Return(nil, errors.New("error")).Once()
		conn.On("Do", "", []interface{}(nil)).Return(0, nil).Once()
		conn.On("Err").Return(nil).Once()
		bin2, err = storage.LookupBin(bin.Name)
		conn.Mock.AssertExpectations(t)
		assert.NotNil(t, err)
	}
}

func TestRedisLookupBins(t *testing.T) {
	conn := &MockedRedisConn{}
	storage := getRedisStorage(conn)
	bins, err := storage.LookupBins([]string{})
	if assert.Nil(t, err) {
		assert.Empty(t, bins)
	}
	bin1 := NewBin()
	bin2 := NewBin()
	if binDump1, err := storage.Dump(bin1); assert.Nil(t, err) {
		if binDump2, err := storage.Dump(bin2); assert.Nil(t, err) {
			conn.On("Do", "MGET", []interface{}{
				storage.getKey(BIN_KEY, bin1.Name),
				storage.getKey(BIN_KEY, bin2.Name),
			}).Return([]interface{}{binDump1, binDump2}, nil).Once()
			conn.On("Do", "", []interface{}(nil)).Return(0, nil).Once()
			conn.On("Err").Return(nil).Once()

			bins, err = storage.LookupBins([]string{bin1.Name, bin2.Name})
			conn.Mock.AssertExpectations(t)
			if assert.Nil(t, err) {
				assert.Equal(t, len(bins), 2)
				assert.Equal(t, bins[0], bin1)
				assert.Equal(t, bins[1], bin2)
			}

			// mget error
			conn.On("Do", "MGET", []interface{}{
				storage.getKey(BIN_KEY, bin1.Name),
			}).Return(nil, errors.New("error")).Once()
			conn.On("Do", "", []interface{}(nil)).Return(0, nil).Once()
			conn.On("Err").Return(nil).Once()

			bins, err = storage.LookupBins([]string{bin1.Name})
			conn.Mock.AssertExpectations(t)
			assert.NotNil(t, err)

			// slice error
			conn.On("Do", "MGET", []interface{}{
				storage.getKey(BIN_KEY, bin1.Name),
			}).Return([]interface {}{string(binDump1)},nil).Once()
			conn.On("Do", "", []interface{}(nil)).Return(0, nil).Once()
			conn.On("Err").Return(nil).Once()

			bins, err = storage.LookupBins([]string{bin1.Name})
			conn.Mock.AssertExpectations(t)
			assert.NotNil(t, err)
		}
	}
}

func TestRedisCreateRequest(t *testing.T) {
	conn := &MockedRedisConn{}
	storage := getRedisStorage(conn)
	bin := NewBin()
	bin.RequestCount = 1
	httpRequest, _ := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("body")))
	req := NewRequest(httpRequest, REQUEST_BODY_SIZE)
	binDump, _ := storage.Dump(bin)
	if reqDump, err := storage.Dump(req); assert.Nil(t, err) {
		conn.On("Send", "LPUSH", []interface{}{
			storage.getKey(REQUESTS_KEY, bin.Name),
			req.Id,
		},
		).Return(nil).Once()
		conn.On("Send", "EXPIRE", []interface{}{
			storage.getKey(REQUESTS_KEY, bin.Name),
			BIN_LIFETIME,
		},
		).Return(nil).Once()
		conn.On("Send", "HSET", []interface{}{
			storage.getKey(REQUEST_HASH_KEY, bin.Name),
			req.Id,
			reqDump,
		},
		).Return(nil).Once()
		conn.On("Send", "EXPIRE", []interface{}{
			storage.getKey(REQUEST_HASH_KEY, bin.Name),
			BIN_LIFETIME,
		},
		).Return(nil).Once()
		conn.On("Flush").Return(nil).Once()
		conn.On("Receive").Return(int64(1), nil).Twice()

		// update bin
		conn.On("Send", "SET", []interface{}{
			storage.getKey(BIN_KEY, bin.Name),
			binDump,
		},
		).Return(nil).Once()
		conn.On("Send", "EXPIRE", []interface{}{
			storage.getKey(BIN_KEY, bin.Name),
			BIN_LIFETIME,
		},
		).Return(nil).Once()

		conn.On("Flush").Return(nil).Once()

		conn.On("Do", "", []interface{}(nil)).Return(0, nil).Twice()
		conn.On("Err").Return(nil).Times(2)
		err = storage.CreateRequest(bin, req)
		conn.Mock.AssertExpectations(t)
		assert.Nil(t, err)
	}
}

func TestRedisLookupRequest(t *testing.T) {
	conn := &MockedRedisConn{}
	storage := getRedisStorage(conn)
	bin := NewBin()
	httpRequest, _ := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("body")))
	req := NewRequest(httpRequest, REQUEST_BODY_SIZE)
	if reqDump, err := storage.Dump(req); assert.Nil(t, err) {
		conn.On("Do", "HGET", []interface{}{
			storage.getKey(REQUEST_HASH_KEY, bin.Name),
			req.Id,
		}).Return(reqDump, nil).Once()
		conn.On("Do", "", []interface{}(nil)).Return(0, nil).Once()
		conn.On("Err").Return(nil).Once()
		req2, err := storage.LookupRequest(bin.Name, req.Id)
		conn.Mock.AssertExpectations(t)
		if assert.Nil(t, err) {
			assert.Equal(t, req2, req)
		}

		// not found
		conn.On("Do", "HGET", []interface{}{
			storage.getKey(REQUEST_HASH_KEY, bin.Name),
			"id",
		}).Return(nil, nil).Once()
		conn.On("Do", "", []interface{}(nil)).Return(0, nil).Once()
		conn.On("Err").Return(nil).Once()
		req2, err = storage.LookupRequest(bin.Name, "id")
		fmt.Println(req2, err)
		conn.Mock.AssertExpectations(t)
		assert.NotNil(t, err)
	}
}

func TestRedisLookupRequests(t *testing.T) {
	conn := &MockedRedisConn{}
	storage := getRedisStorage(conn)
	bin := NewBin()
	httpRequest, _ := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("body")))
	req1 := NewRequest(httpRequest, REQUEST_BODY_SIZE)
	req2 := NewRequest(httpRequest, REQUEST_BODY_SIZE)
	req3 := NewRequest(httpRequest, REQUEST_BODY_SIZE)
	req1Dump, err1 := storage.Dump(req1)
	req2Dump, err2 := storage.Dump(req2)
	req3Dump, err3 := storage.Dump(req3)
	if assert.Nil(t, err1) && assert.Nil(t, err2) && assert.Nil(t, err3) {
		conn.On("Do", "LRANGE", []interface{}{
			storage.getKey(REQUESTS_KEY, bin.Name),
			0, 2 - 1,
		}).Return([]interface{}{[]byte(req1.Id), []byte(req2.Id), []byte(req3.Id)}, nil).Once()

		conn.On("Do", "HMGET", []interface {}{
			storage.getKey(REQUEST_HASH_KEY, bin.Name),
			req1.Id, req2.Id, req3.Id,
		}).Return([]interface{}{req1Dump, req2Dump, req3Dump}, nil).Once()

		conn.On("Do", "", []interface{}(nil)).Return(0, nil).Once()
		conn.On("Err").Return(nil).Once()
		requests, err := storage.LookupRequests(bin.Name, 0, 2)
		conn.Mock.AssertExpectations(t)
		if assert.Nil(t, err) {
			assert.Equal(t, len(requests), 3)
			assert.Equal(t, requests[0], req1)
			assert.Equal(t, requests[1], req2)
			assert.Equal(t, requests[2], req3)
		}
	}
}
