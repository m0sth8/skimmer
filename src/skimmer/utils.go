package skimmer

import (
	"math/rand"
	"time"
)

type RandomString struct {
	pool string
	rg   *rand.Rand
}

func NewRandomString(pool string) *RandomString {
	return &RandomString{
		pool,
		rand.New(rand.NewSource(time.Now().Unix())),
	}
}

func (rs *RandomString) Generate(length int) (r string) {
	if length < 1 {
		return
	}
	b := make([]byte, length)
	for i, _ := range b {
		b[i] = rs.pool[rs.rg.Intn(len(rs.pool))]
	}
	r = string(b)
	return
}
