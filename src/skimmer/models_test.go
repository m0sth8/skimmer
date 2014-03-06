package skimmer

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"bytes"
	"strings"
	"net/http"
	"fmt"
)

func TestRs(t *testing.T) {
	pool := "0123456789abcdefghijklmnopqrstuvwxyz"
	val := rs.Generate(1000)
	for _, alpha := range val {
		assert.Contains(t, pool, string(alpha))
	}
}

func TestNewBin(t *testing.T) {
	now := time.Now().Unix()
	bin := NewBin()
	if assert.NotNil(t, bin) {
		assert.Equal(t, len(bin.Name), 6)
		assert.Equal(t, bin.RequestCount, 0)
		assert.Equal(t, bin.Created, bin.Updated)
		assert.True(t, bin.Created < (now+1))
		assert.True(t, bin.Created > (now-1))
	}
}

func TestNewRequest(t *testing.T) {
	body := []byte("arg1=value1&arg2=value2")
	httpRequest, err := http.NewRequest("POST", "http://www.google.com/path/", bytes.NewBuffer(body))
	if assert.Nil(t, err){
		httpRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req := NewRequest(httpRequest, 100)
		if assert.NotNil(t, req) {
			assert.Equal(t, len(req.Id), 12)
			assert.Equal(t, req.Method, "POST")
			assert.Equal(t, req.Proto, "HTTP/1.1")
			assert.Equal(t, req.Host, "www.google.com")
			assert.Equal(t, req.ContentLength, len(body))
			assert.Equal(t, req.RequestURI, "/path/")
			assert.Equal(t, req.FormValue, map[string][]string{"arg1": {"value1"}, "arg2": {"value2"}})
			assert.Equal(t, req.FormFile, []string(nil))
			assert.Equal(t, req.Body, string(body))
		}
	}
	httpRequest, err = http.NewRequest("POST", "http://www.google.com/path/", bytes.NewBuffer(body))
	if assert.Nil(t, err){
		httpRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		// test truncating
		req := NewRequest(httpRequest, 10)
		if assert.NotNil(t, req) {
			assert.Equal(t, req.ContentLength, len(body))
			assert.Equal(t, req.FormValue, map[string][]string{"arg1": {"value1"}, "arg2": {"value2"}})
			assert.NotEqual(t, req.Body, string(body))
			assert.True(t, strings.HasSuffix(string(req.Body), fmt.Sprintf("\n<<<TRUNCATED , %d of %d", 10, len(body))))
			assert.Equal(t, strings.Index(string(req.Body), "\n<<<TRUNCATED"), 10)
		}
	}

	multipartBody := []byte("\r\n--B1897D34-1FF0-40CD-BDAA-61A5662758C1\r\nContent-Disposition: form-data; name=\"todo.sublime-project\"; filename=\"todo.sublime-project\"\r\nContent-Transfer-Encoding: binary\r\nContent-Type: (null)\r\n\r\n{\n}\n\r\n--B1897D34-1FF0-40CD-BDAA-61A5662758C1--\r\n")
	httpRequest, err = http.NewRequest("POST", "http://www.google.com/path/", bytes.NewBuffer(multipartBody))
	if assert.Nil(t, err) {
		httpRequest.Header.Add("Content-Type", "multipart/form-data; boundary=B1897D34-1FF0-40CD-BDAA-61A5662758C1")
		req := NewRequest(httpRequest, 1000)
		if assert.NotNil(t, req) {
			assert.NotEqual(t, req.ContentLength, len(body))
			assert.Equal(t, req.FormValue, map[string][]string{})
			assert.Equal(t, req.Body, string(multipartBody))
			assert.Equal(t, req.FormFile, []string{"todo.sublime-project"})
		}
	}

}
