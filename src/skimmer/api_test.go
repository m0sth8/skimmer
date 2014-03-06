package skimmer


import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"encoding/json"
	"errors"
	"fmt"
	"bytes"
)

type MockedStorage struct{
	mock.Mock
}

func (s *MockedStorage) CreateBin(_ *Bin) error {
	args := s.Mock.Called()
	return args.Error(0)
}

func (s *MockedStorage) UpdateBin(bin *Bin) error {
	args := s.Mock.Called(bin)
	return args.Error(0)
}

func (s *MockedStorage) LookupBin(name string) (*Bin, error) {
	args := s.Mock.Called(name)
	return args.Get(0).(*Bin), args.Error(1)
}

func (s *MockedStorage) LookupBins(names []string) ([]*Bin, error) {
	args := s.Mock.Called(names)
	return args.Get(0).([]*Bin), args.Error(1)
}

func (s *MockedStorage) LookupRequest(binName, id string) (*Request, error) {
	args := s.Mock.Called(binName, id)
	return args.Get(0).(*Request), args.Error(1)
}

func (s *MockedStorage) CreateRequest(bin *Bin, req *Request) error {
	args := s.Mock.Called(bin)
	return args.Error(0)
}

func (s *MockedStorage) LookupRequests(binName string, from, to int) ([]*Request, error) {
	args := s.Mock.Called(binName, from, to)
	return args.Get(0).([]*Request), args.Error(1)
}


func TestBinsPost(t *testing.T) {
	api := GetApi(&Config{SessionSecret: "123"})
	req, err := http.NewRequest("POST", "/api/v1/bins/", bytes.NewBuffer([]byte("{}")))
	if assert.Nil(t, err) {
		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)
		assert.Equal(t, res.Code, 201)
		data := map[string] interface {}{}
		err = json.Unmarshal([]byte(res.Body.String()), &data)
		if assert.Nil(t, err) {
			assert.Equal(t, len(data["name"].(string)), 6)
			assert.False(t, data["private"].(bool))
		}
	}
	req, err = http.NewRequest("POST", "/api/v1/bins/", bytes.NewBuffer([]byte("{\"private\":true}")))
	if assert.Nil(t, err) {
		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)
		assert.Equal(t, res.Code, 201)
		data := map[string] interface {}{}
		err = json.Unmarshal(res.Body.Bytes(), &data)
		if assert.Nil(t, err) {
			assert.Equal(t, len(data["name"].(string)), 6)
			fmt.Println(data)
			assert.True(t, data["private"].(bool))
		}
	}
	// decoding payload error
	if req, err = http.NewRequest("POST", "/api/v1/bins/", bytes.NewBuffer([]byte(""))); assert.Nil(t, err){
		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)
		assert.Equal(t, res.Code, 400)
	}

	mockedStorage := &MockedStorage{}
	api.MapTo(mockedStorage, (*Storage)(nil))
	req, err = http.NewRequest("POST", "/api/v1/bins/", bytes.NewBuffer([]byte("{}")))
	if assert.Nil(t, err) {
		res := httptest.NewRecorder()
		mockedStorage.On("CreateBin").Return(errors.New("Storage error"))
		api.ServeHTTP(res, req)
		mockedStorage.AssertExpectations(t)
		assert.Equal(t, res.Code, 500)
		assert.Contains(t, res.Body.String(), "Storage error")
	}
}

func TestBinsGet(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/bins/", nil)
	if assert.Nil(t, err) {
		api := GetApi(&Config{SessionSecret: "123"})
		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)
		assert.Equal(t, res.Code, 200)
		data := []map[string] interface {}{}
		err = json.Unmarshal([]byte(res.Body.String()), &data)
		if assert.Nil(t, err) {
			assert.Equal(t, len(data), 0)
		}
		bin1, cookie, err := createBin(api, "{}", "")
		if assert.Nil(t, err) {
			bin2, cookie, err := createBin(api, "{}", cookie)
			if assert.Nil(t, err) {
				assert.NotEqual(t, bin1, bin2)
				res := httptest.NewRecorder()
				req.Header.Set("Cookie", cookie)
				api.ServeHTTP(res, req)
				req.Header.Del("Cookie")
				assert.Equal(t, res.Code, 200)
				data := []map[string] interface {}{}
				err = json.Unmarshal([]byte(res.Body.String()), &data)
				if assert.Nil(t, err) {
					if assert.Equal(t, len(data), 2) {
						assert.Equal(t, data[1], bin1)
						assert.Equal(t, data[0], bin2)
					}
				}
			}
		}

		api = GetApi(&Config{SessionSecret: "123"})
		mockedStorage := &MockedStorage{}
		api.MapTo(mockedStorage, (*Storage)(nil))
		res = httptest.NewRecorder()
		mockedStorage.On("LookupBins", []string{}).Return([]*Bin(nil), errors.New("Storage error"))
		api.ServeHTTP(res, req)
		mockedStorage.AssertExpectations(t)
		if assert.Equal(t, res.Code, 500) {
			assert.Contains(t, res.Body.String(), "Storage error")
		}
	}

}

func TestBinGet(t *testing.T) {
	api := GetApi(&Config{SessionSecret: "123"})
	req, err := http.NewRequest("GET", "/api/v1/bins/name22", nil)
	if assert.Nil(t, err) {
		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)
		assert.Equal(t, res.Code, 404)
	}
	bin, cookie, err := createBin(api, "{}", "")
	if assert.Nil(t, err) {
		req, err = http.NewRequest("GET", fmt.Sprintf("/api/v1/bins/%s", bin["name"]), nil)
		if assert.Nil(t, err) {
			res := httptest.NewRecorder()
			api.ServeHTTP(res, req)
			if assert.Equal(t, res.Code, 200) {
				tBin := map[string]interface {}{}
				err = json.Unmarshal([]byte(res.Body.String()), &tBin)
				if assert.Nil(t, err) {
					assert.Equal(t, tBin, bin)
				}
			}
		}
	}
	// test private bin
	bin, cookie, err = createBin(api, "{\"private\":true}", "")
	if assert.Nil(t, err) {
		req, err = http.NewRequest("GET", fmt.Sprintf("/api/v1/bins/%s", bin["name"]), nil)
		if assert.Nil(t, err) {
			req.Header.Set("Cookie", cookie)
			res := httptest.NewRecorder()
			api.ServeHTTP(res, req)
			if assert.Equal(t, res.Code, 200) {
				tBin := map[string]interface {}{}
				err = json.Unmarshal(res.Body.Bytes(), &tBin)
				if assert.Nil(t, err) {
					assert.Equal(t, tBin, bin)
				}
			}
			req.Header.Del("Cookie")
			res = httptest.NewRecorder()
			api.ServeHTTP(res, req)
			assert.Equal(t, res.Code, 403)
		}
	}

	req, err = http.NewRequest("GET", "/api/v1/bins/name22", nil)
	if assert.Nil(t, err) {
		api = GetApi(&Config{SessionSecret: "123"})
		mockedStorage := &MockedStorage{}
		api.MapTo(mockedStorage, (*Storage)(nil))
		res := httptest.NewRecorder()
		mockedStorage.On("LookupBin", "name22").Return(new(Bin), errors.New("Storage error"))
		api.ServeHTTP(res, req)
		mockedStorage.AssertExpectations(t)
		assert.Equal(t, res.Code, 404)
	}

}

func TestRequestCreate(t *testing.T) {
	api := GetApi(&Config{SessionSecret: "123"})
	bin, cookie, err := createBin(api, "{}", "")
	if assert.Nil(t, err) {
		req, err := http.NewRequest("GET", fmt.Sprintf("/bins/%s", bin["name"]), bytes.NewBuffer(nil))
		if assert.Nil(t, err) {
			req.Header.Set("Cookie", cookie)
			res := httptest.NewRecorder()
			api.ServeHTTP(res, req)
			if assert.Equal(t, res.Code, 200) {
				request := map[string]interface {}{}
				err = json.Unmarshal([]byte(res.Body.String()), &request)
				if assert.Nil(t, err) {
					assert.Equal(t, request["method"].(string), "GET")
				}
			}
		}
		req, err = http.NewRequest("GET", fmt.Sprintf("/bins/%s", "name22"), bytes.NewBuffer(nil))
		if assert.Nil(t, err) {
			res := httptest.NewRecorder()
			api.ServeHTTP(res, req)
			assert.Equal(t, res.Code, 404)
		}
	}

	api = GetApi(&Config{SessionSecret: "123"})
	mockedStorage := &MockedStorage{}
	api.MapTo(mockedStorage, (*Storage)(nil))
	res := httptest.NewRecorder()
	realBin := NewBin()
	req, err := http.NewRequest("GET", fmt.Sprintf("/bins/%s", realBin.Name), bytes.NewBuffer(nil))
	if assert.Nil(t, err){
		mockedStorage.On("LookupBin", realBin.Name).Return(realBin, nil)
		mockedStorage.On("CreateRequest", realBin).Return(errors.New("Storage error"))
		api.ServeHTTP(res, req)
		mockedStorage.AssertExpectations(t)
		if assert.Equal(t, res.Code, 500) {
			assert.Contains(t, res.Body.String(), "Storage error")
		}
	}

}

func TestRequestsGet(t *testing.T) {
	api := GetApi(&Config{SessionSecret: "123"})

	if bin, cookie, err := createBin(api, "{}", ""); assert.Nil(t, err) {
		url := fmt.Sprintf("/api/v1/bins/%s/requests/", bin["name"])
		if req, err := http.NewRequest("GET", url, bytes.NewBuffer(nil)); assert.Nil(t, err){
			req.Header.Set("Cookie", cookie)
			res := httptest.NewRecorder()
			api.ServeHTTP(res, req)
			if assert.Equal(t, res.Code, 200) {
				requests := []map[string]interface {}{}
				if err = json.Unmarshal([]byte(res.Body.String()), &requests); assert.Nil(t, err) {
					assert.Equal(t, len(requests), 0)
				}
			}
		}
		if request1, err := createRequest(api, bin["name"].(string)); assert.Nil(t, err) {
			if request2, err := createRequest(api, bin["name"].(string)); assert.Nil(t, err) {
				url = fmt.Sprintf("/api/v1/bins/%s/requests/", bin["name"])
				if req, err := http.NewRequest("GET", url, nil); assert.Nil(t, err) {
					res := httptest.NewRecorder()
					api.ServeHTTP(res, req)
					if assert.Equal(t, res.Code, 200) {
						requests := []map[string]interface {}{}
						if err = json.Unmarshal(res.Body.Bytes(), &requests); assert.Nil(t, err) {
							assert.Equal(t, len(requests), 2)
							assert.Equal(t, requests, []map[string]interface {}{request2, request1})
						}
					}
				}
				url = fmt.Sprintf("/api/v1/bins/%s/requests/?from=1&to=2", bin["name"])

				if req, err := http.NewRequest("GET", url, nil); assert.Nil(t, err) {
					res := httptest.NewRecorder()
					api.ServeHTTP(res, req)
					if assert.Equal(t, res.Code, 200) {
						requests := []map[string]interface {}{}
						if err = json.Unmarshal([]byte(res.Body.String()), &requests); assert.Nil(t, err) {
							assert.Equal(t, len(requests), 1)
							assert.Equal(t, requests, []map[string]interface {}{request1})
						}
					}
				}
			}
		}
	}

	api = GetApi(&Config{SessionSecret: "123"})
	mockedStorage := &MockedStorage{}
	api.MapTo(mockedStorage, (*Storage)(nil))
	realBin := NewBin()
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/bins/%s/requests/", realBin.Name), bytes.NewBuffer(nil))
	if assert.Nil(t, err){
		res := httptest.NewRecorder()
		mockedStorage.On("LookupBin", realBin.Name).Return(realBin, errors.New("Storage error")).Once()
		api.ServeHTTP(res, req)
		mockedStorage.AssertExpectations(t)
		assert.Equal(t, res.Code, 404)

		res = httptest.NewRecorder()
		mockedStorage.On("LookupBin", realBin.Name).Return(realBin, nil)
		mockedStorage.On("LookupRequests", realBin.Name, 0, 20).Return([]*Request{}, errors.New("Storage error"))
		api.ServeHTTP(res, req)
		mockedStorage.AssertExpectations(t)
		if assert.Equal(t, res.Code, 500) {
			assert.Contains(t, res.Body.String(), "Storage error")
		}
	}
}

func TestRequestGet(t *testing.T){
	api := GetApi(&Config{SessionSecret: "123"})

	if bin, cookie, err := createBin(api, "{}", ""); assert.Nil(t, err) {
		url := fmt.Sprintf("/api/v1/bins/%s/requests/name", bin["name"])
		if req, err := http.NewRequest("GET", url, nil); assert.Nil(t, err){
			req.Header.Set("Cookie", cookie)
			res := httptest.NewRecorder()
			api.ServeHTTP(res, req)
			assert.Equal(t, res.Code, 404)
		}
		if request, err := createRequest(api, bin["name"].(string)); assert.Nil(t, err) {
			url = fmt.Sprintf("/api/v1/bins/%s/requests/%s", bin["name"], request["id"])
			if req, err := http.NewRequest("GET", url, nil); assert.Nil(t, err) {
				req.Header.Set("Cookie", cookie)
				res := httptest.NewRecorder()
				api.ServeHTTP(res, req)
				if assert.Equal(t, res.Code, 200) {
					tRequest := map[string]interface {}{}
					if err = json.Unmarshal([]byte(res.Body.String()), &tRequest); assert.Nil(t, err) {
						assert.Equal(t, tRequest, request)
					}
				}
			}
		}
	}

	api = GetApi(&Config{SessionSecret: "123"})
	mockedStorage := &MockedStorage{}
	api.MapTo(mockedStorage, (*Storage)(nil))
	res := httptest.NewRecorder()
	realBin := NewBin()
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/bins/%s/requests/id", realBin.Name), bytes.NewBuffer(nil))
	if assert.Nil(t, err){
		mockedStorage.On("LookupRequest", realBin.Name, "id").Return(new(Request), errors.New("Storage error"))
		api.ServeHTTP(res, req)
		mockedStorage.AssertExpectations(t)
		assert.Equal(t, res.Code, 404)
	}
}

func createBin(handler http.Handler, body string, cookie string) (bin map[string]interface {}, setCookie string, err error){
	if req, err := http.NewRequest("POST", "/api/v1/bins/", bytes.NewBuffer([]byte(body))); err == nil {
		req.Header.Set("Cookie", cookie)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		setCookie = res.Header().Get("Set-Cookie")
		err = json.Unmarshal([]byte(res.Body.String()), &bin)
	}
	return
}

func createRequest(handler http.Handler, binName string) (request map[string]interface {}, err error){
	if req, err := http.NewRequest("POST", fmt.Sprintf("/bins/%s", binName), bytes.NewBuffer(nil)); err == nil {
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		err = json.Unmarshal(res.Body.Bytes(), &request)
	}
	return
}

