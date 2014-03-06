package skimmer

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/ugorji/go/codec"
	"strings"
	"time"
)

type RedisStorage struct {
	BaseStorage
	pool       *redis.Pool
	prefix     string
	cleanTimer *time.Timer
}

const (
	KEY_SEPARATOR    = "|"
	BIN_KEY          = "bins"
	REQUESTS_KEY     = "rq"
	REQUEST_HASH_KEY = "rhsh"
	CLEANING_SET	 = "cln"
	CLEANING_FACTOR  = 3
)

func getPool(server string, password string) (pool *redis.Pool) {
	pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, _ time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	return pool
}

func NewRedisStorage(server, password, prefix string, maxRequests int, binLifetime int64) *RedisStorage {
	return &RedisStorage{
		BaseStorage{maxRequests, binLifetime},
		getPool(server, password),
		prefix,
		&time.Timer{},
	}
}

func (storage *RedisStorage) StartCleaning(timeout int) {
	defer func(){
		storage.cleanTimer = time.AfterFunc(time.Duration(timeout) * time.Second, func(){storage.StartCleaning(timeout)})
	}()
	storage.clean()
}

func (storage *RedisStorage) StopCleaning() {
	if storage.cleanTimer != nil {
		storage.cleanTimer.Stop()
	}
}

func (storage *RedisStorage) clean() {
	for {
		conn := storage.pool.Get()
		defer conn.Close()
		binName, err := redis.String(conn.Do("SPOP", storage.getKey(CLEANING_SET)))
		if err != nil {
			break
		}
		conn.Send("LRANGE", storage.getKey(REQUESTS_KEY, binName), storage.maxRequests, -1)
		conn.Send("LTRIM", storage.getKey(REQUESTS_KEY, binName), 0, storage.maxRequests-1)
		conn.Flush()
		if values, error := redis.Values(conn.Receive()); error == nil {
			ids := []string{}
			if err := redis.ScanSlice(values, &ids); err != nil {
				continue
			}
			if len(ids) > 0 {
				args := redis.Args{}.Add(storage.getKey(REQUEST_HASH_KEY, binName)).AddFlat(ids)
				conn.Do("HDEL", args...)
			}
		}
	}
}

func (storage *RedisStorage) getKey(keys ...string) string {
	return fmt.Sprintf("%s%s%s", storage.prefix, KEY_SEPARATOR, strings.Join(keys, KEY_SEPARATOR))
}

func (storage *RedisStorage) LookupBin(name string) (bin *Bin, err error) {
	conn := storage.pool.Get()
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", storage.getKey(BIN_KEY, name)))
	if err != nil {
		if err == redis.ErrNil {
			err = errors.New("Bin was not found")
		}
		return
	}
	err = storage.Load(reply, &bin)
	return
}

func (storage *RedisStorage) LookupBins(names []string) ([]*Bin, error) {
	bins := []*Bin{}
	if len(names) == 0 {
		return bins, nil
	}
	args := redis.Args{}
	for _, name := range names {
		args = args.Add(storage.getKey(BIN_KEY, name))
	}
	conn := storage.pool.Get()
	defer conn.Close()
	if values, err := redis.Values(conn.Do("MGET", args...)); err == nil {
		bytes := [][]byte{}
		if err = redis.ScanSlice(values, &bytes); err != nil {
			return nil, err
		}
		for _, rawbin := range bytes {
			if len(rawbin) > 0 {
				bin := &Bin{}
				if err := storage.Load(rawbin, bin); err == nil {
					bins = append(bins, bin)
				}
			}
		}
		return bins, nil
	} else {
		return nil, err
	}
}

func (storage *RedisStorage) UpdateBin(bin *Bin) (err error) {
	dumpedBin, err := storage.Dump(bin)
	if err != nil {
		return
	}
	conn := storage.pool.Get()
	defer conn.Close()
	key := storage.getKey(BIN_KEY, bin.Name)
	err = conn.Send("SET", key, dumpedBin)
	if err != nil {
		return
	}
	err = conn.Send("EXPIRE", key, storage.binLifetime)
	conn.Flush()
	return err
}

func (storage *RedisStorage) LookupRequest(binName string, id string) (req *Request, err error) {
	conn := storage.pool.Get()
	defer conn.Close()
	data, err := redis.Bytes(conn.Do("HGET", storage.getKey(REQUEST_HASH_KEY, binName), id))
	if err != nil {
		if err == redis.ErrNil {
			err = errors.New("Request was not found")
		}
		return
	}
	req = &Request{}
	err = storage.Load(data, req)
	return
}

func (storage *RedisStorage) LookupRequests(binName string, from int, to int) ([]*Request, error) {
	conn := storage.pool.Get()
	requests := []*Request{}
	defer conn.Close()
	if from < 0 {
		return nil, errors.New("from argument should be more then 0")
	}
	if to < 0 {
		return nil, errors.New("to argument should be more then 0")
	}
	if to > storage.maxRequests {
		to = storage.maxRequests
	}
	if from > to {
		return nil, errors.New("value of agrument to should be less then value of argument from")
	}
	if from == to {
		return requests, nil
	}
	if values, err := redis.Values(conn.Do("LRANGE", storage.getKey(REQUESTS_KEY, binName), from, to-1)); err == nil {
		ids := []string{}
		if error := redis.ScanSlice(values, &ids); error != nil {
			return nil, error
		}
		if len(ids) > 0 {
			args := redis.Args{}.Add(storage.getKey(REQUEST_HASH_KEY, binName)).AddFlat(ids)
			if values, err := redis.Values(conn.Do("HMGET", args...)); err == nil {
				bytes := [][]byte{}
				if error := redis.ScanSlice(values, &bytes); error != nil {
					return nil, error
				}
				for _, raw := range bytes {
					if len(raw) > 0 {
						request := &Request{}
						if err := storage.Load(raw, request); err == nil {
							requests = append(requests, request)
						}
					}
				}
			} else {
				return nil, err
			}
		}
		return requests, nil
	} else {
		return nil, err
	}

}

func (storage *RedisStorage) CreateBin(bin *Bin) error {
	if err := storage.UpdateBin(bin); err != nil {
		return err
	}
	return nil
}

func (storage *RedisStorage) CreateRequest(bin *Bin, req *Request) (err error) {
	data, err := storage.Dump(req)
	if err != nil {
		return
	}
	conn := storage.pool.Get()
	defer conn.Close()
	key := storage.getKey(REQUESTS_KEY, bin.Name)
	err = conn.Send("LPUSH", key, req.Id)
	if err != nil {
		return
	}
	err = conn.Send("EXPIRE", key, storage.binLifetime)
	if err != nil {
		return
	}
	key = storage.getKey(REQUEST_HASH_KEY, bin.Name)
	err = conn.Send("HSET", key, req.Id, data)
	if err != nil {
		return
	}
	err = conn.Send("EXPIRE", key, storage.binLifetime)
	if err != nil {
		return
	}
	conn.Flush()
	requestCount, err := redis.Int(conn.Receive())
	if err != nil {
		return
	}
	if requestCount < storage.maxRequests {
		bin.RequestCount = requestCount
	} else {
		bin.RequestCount = storage.maxRequests
	}
	bin.Updated = time.Now().Unix()
	if _, err = redis.Int(conn.Receive()); err != nil {
		return
	}
	if requestCount > storage.maxRequests * CLEANING_FACTOR {
		conn.Do("SADD", storage.getKey(CLEANING_SET), bin.Name)
	}
	if err = storage.UpdateBin(bin); err != nil {
		return
	}
	return
}

func (storage *RedisStorage) Dump(v interface{}) (data []byte, err error) {
	var (
		mh codec.MsgpackHandle
		h  = &mh
	)
	err = codec.NewEncoderBytes(&data, h).Encode(v)
	return
}

func (storage *RedisStorage) Load(data []byte, v interface{}) error {
	var (
		mh codec.MsgpackHandle
		h  = &mh
	)
	return codec.NewDecoderBytes(data, h).Decode(v)
}
