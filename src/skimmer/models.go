package skimmer

import "time"

var rs = NewRandomString("0123456789abcdefghijklmnopqrstuvwxyz")

type Bin struct {
	Name         string `json:"name"`
	Created      int64  `json:"created"`
	Updated      int64  `json:"updated"`
	RequestCount int    `json:"requestCount"`
}

func NewBin() *Bin {
	now := time.Now().Unix()
	bin := Bin{
		Created:      now,
		Updated:	  now,
		Name:         rs.Generate(6),
	}
	return &bin
}
