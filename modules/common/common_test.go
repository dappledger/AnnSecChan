package common

import (
	"encoding/base64"
	"testing"
)

func TestParseListenAddress(t *testing.T) {
	t.Log(ParseListenAddress("tcp://127.1:8080"))
}

func TestBench(t *testing.T) {

	byt := []byte("12345678")

	t.Log(base64.StdEncoding.EncodeToString(byt))

	h := Hash(byt)

	t.Log(base64.StdEncoding.EncodeToString(h))

	t.Log(Bytes2Hex(h))

	t.Log(IsHash(Bytes2Hex(h)))
}
