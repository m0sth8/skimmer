package skimmer

import (
	"testing"
	"github.com/stretchr/testify/assert"

)


func TestRandomString(t *testing.T) {
	pool := "abcdefg12345"
	rs := NewRandomString(pool)
	val := rs.Generate(1000)
	for _, alpha := range(val) {
		assert.Contains(t, pool, string(alpha))
	}
	assert.Empty(t, rs.Generate(-1))
}

func TestRandomByte(t *testing.T) {
//	assert.Equal(t, int(randomByte(100, 5)), 0)
}
