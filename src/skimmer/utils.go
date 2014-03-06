package skimmer

import (
	"math/rand"
	"time"
	"fmt"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

var rg = rand.New(rand.NewSource(time.Now().Unix()))

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

func randomByte(gradient, floor int) byte {
	if gradient == 0 {
		gradient = 1
	}
	max := int(255 / gradient)
	return byte((rg.Intn(max - floor) + floor) * gradient)
}

func RandomColor() ([3]byte){
	return [3]byte{randomByte(5, 5), randomByte(5, 5), randomByte(5, 5)}
}

func Solid16x16gifDatauri(color [3]byte) string{
	return fmt.Sprintf("data:image/gif;base64,R0lGODlhEAAQAIAA%sACH5BAQAAAAALAAAAAAQABAAAAIOhI+py+0Po5y02ouzPgUAOw==",
		base64.StdEncoding.EncodeToString([]byte{0, color[0], color[1], color[2], 0, 0}))
}

func DecodeJsonPayload(r *http.Request, v interface{}) error {
	content, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, v)
	if err != nil {
		return err
	}
	return nil
}
