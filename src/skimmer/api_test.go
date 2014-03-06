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
	api := GetApi()
	req, err := http.NewRequest("POST", "/api/v1/bins/", nil)
	if assert.Nil(t, err) {
		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)
		assert.Equal(t, res.Code, 201)
		data := map[string] interface {}{}
		err = json.Unmarshal([]byte(res.Body.String()), &data)
		if assert.Nil(t, err) {
			assert.Equal(t, len(data["name"].(string)), 6)
		}
	}
	mockedStorage := &MockedStorage{}
	api.MapTo(mockedStorage, (*Storage)(nil))
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
		api := GetApi()
		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)
		assert.Equal(t, res.Code, 200)
		data := []map[string] interface {}{}
		err = json.Unmarshal([]byte(res.Body.String()), &data)
		if assert.Nil(t, err) {
			assert.Equal(t, len(data), 0)
		}

		bin1, err := createBin(api)
		if assert.Nil(t, err) {
			bin2, err := createBin(api)
			if assert.Nil(t, err) {
				assert.NotEqual(t, bin1, bin2)
				res := httptest.NewRecorder()
				api.ServeHTTP(res, req)
				assert.Equal(t, res.Code, 200)
				data := []map[string] interface {}{}
				err = json.Unmarshal([]byte(res.Body.String()), &data)
				if assert.Nil(t, err) {
					if assert.Equal(t, len(data), 2) {
						assert.Equal(t, data[0], bin1)
						assert.Equal(t, data[1], bin2)
					}
				}
			}
		}
		req, _ := http.NewRequest("GET", "/api/v1/bins/", nil)
		api = GetApi()
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
	api := GetApi()
	req, err := http.NewRequest("GET", "/api/v1/bins/name22", nil)
	if assert.Nil(t, err) {
		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)
		assert.Equal(t, res.Code, 404)
	}
	bin, err := createBin(api)
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
	req, err = http.NewRequest("GET", "/api/v1/bins/name22", nil)
	if assert.Nil(t, err) {
		api = GetApi()
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
	api := GetApi()
	bin, err := createBin(api)
	if assert.Nil(t, err) {
		req, err := http.NewRequest("GET", fmt.Sprintf("/bins/%s", bin["name"]), bytes.NewBuffer(nil))
		if assert.Nil(t, err) {
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

	api = GetApi()
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
	api := GetApi()

	if bin, err := createBin(api); assert.Nil(t, err) {
		url := fmt.Sprintf("/api/v1/bins/%s/requests/", bin["name"])
		if req, err := http.NewRequest("GET", url, nil); assert.Nil(t, err){
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
						if err = json.Unmarshal([]byte(res.Body.String()), &requests); assert.Nil(t, err) {
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

	api = GetApi()
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
	api := GetApi()

	if bin, err := createBin(api); assert.Nil(t, err) {
		url := fmt.Sprintf("/api/v1/bins/%s/requests/name", bin["name"])
		if req, err := http.NewRequest("GET", url, nil); assert.Nil(t, err){
			res := httptest.NewRecorder()
			api.ServeHTTP(res, req)
			assert.Equal(t, res.Code, 404)
		}
		if request, err := createRequest(api, bin["name"].(string)); assert.Nil(t, err) {
			url = fmt.Sprintf("/api/v1/bins/%s/requests/%s", bin["name"], request["id"])
			if req, err := http.NewRequest("GET", url, nil); assert.Nil(t, err) {
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

	api = GetApi()
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

func createBin(handler http.Handler) (bin map[string]interface {}, err error){
	if req, err := http.NewRequest("POST", "/api/v1/bins/", nil); err == nil {
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		err = json.Unmarshal([]byte(res.Body.String()), &bin)
	}
	return
}

func createRequest(handler http.Handler, binName string) (request map[string]interface {}, err error){
	if req, err := http.NewRequest("POST", fmt.Sprintf("/bins/%s", binName), bytes.NewBuffer(nil)); err == nil {
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		err = json.Unmarshal([]byte(res.Body.String()), &request)
	}
	return
}

