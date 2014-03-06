package skimmer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var rs = NewRandomString("0123456789abcdefghijklmnopqrstuvwxyz")

type Bin struct {
	Name         string  `json:"name"`
	Created      int64   `json:"created"`
	Updated      int64   `json:"updated"`
	RequestCount int     `json:"requestCount"`
	Color		 [3]byte `json:"color"`
	Favicon      string  `json:"favicon"`
	Private      bool    `json:"private"`
	SecretKey    string  `json:"-"`
}

func (bin *Bin) SetPrivate() {
	bin.Private = true
	bin.SecretKey = rs.Generate(32)
}

func NewBin() *Bin {
	color:= RandomColor()
	now := time.Now().Unix()
	bin := Bin{
		Created:      now,
		Updated:	  now,
		Name:         rs.Generate(6),
		Color:		  color,
		Favicon:      Solid16x16gifDatauri(color),
		Private:	  false,
	}
	return &bin
}

type Request struct {
	Id      string `json:"id"`
	Created int64  `json:"created"`

	Method        string              `json:"method"` // GET, POST, PUT, etc.
	Proto         string              `json:"proto"`  // "HTTP/1.0"
	Header        http.Header         `json:"header"`
	ContentLength int64               `json:"contentLength"`
	RemoteAddr    string              `json:"remoteAddr"`
	Host          string              `json:"host"`
	RequestURI    string              `json:"requestURI"`
	Body          string              `json:"body"`
	FormValue     map[string][]string `json:"formValue"`
	FormFile      []string            `json:"formFile"`
}

func NewRequest(httpRequest *http.Request, maxBodySize int) *Request {
	var (
		bodyValue string
		formValue map[string][]string
		formFile  []string
	)
	if body, err := ioutil.ReadAll(httpRequest.Body); err == nil {
		if len(body) > 0 && maxBodySize != 0 {
			if maxBodySize == -1 || httpRequest.ContentLength < int64(maxBodySize) {
				bodyValue = string(body)
			} else {
				bodyValue = fmt.Sprintf("%s\n<<<TRUNCATED , %d of %d", string(body[0:maxBodySize]),
					maxBodySize, httpRequest.ContentLength)
			}
		}
		httpRequest.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		defer httpRequest.Body.Close()
	}
	httpRequest.ParseMultipartForm(0)
	if httpRequest.MultipartForm != nil {
		formValue = httpRequest.MultipartForm.Value
		for key := range httpRequest.MultipartForm.File {
			formFile = append(formFile, key)
		}
	} else {
		formValue = httpRequest.PostForm
	}
	if realIp := httpRequest.Header.Get("X-Real-Ip"); realIp != "" {
		httpRequest.Header.Del("X-Real-Ip")
		httpRequest.RemoteAddr = realIp
	}
	request := Request{
		Id:            rs.Generate(12),
		Created:       time.Now().Unix(),
		Method:        httpRequest.Method,
		Proto:         httpRequest.Proto,
		Host:          httpRequest.Host,
		Header:        httpRequest.Header,
		ContentLength: httpRequest.ContentLength,
		RemoteAddr:    httpRequest.RemoteAddr,
		RequestURI:    httpRequest.URL.RequestURI(),
		FormValue:     formValue,
		FormFile:      formFile,
		Body:          bodyValue,
	}
	return &request
}
