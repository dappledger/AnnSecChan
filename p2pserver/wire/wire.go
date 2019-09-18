package wire

import (
	"encoding/json"

	"github.com/dappledger/AnnSecChan/modules/rlp"
)

type Wire interface {
	Encode(interface{}) ([]byte, error)
	Decode([]byte, interface{}) error
}

type WireRlp struct{}

func (w *WireRlp) Encode(v interface{}) ([]byte, error) {
	return rlp.EncodeToBytes(v)
}

func (w *WireRlp) Decode(b []byte, v interface{}) error {
	return rlp.DecodeBytes(b, v)
}

type WireJson struct{}

func (w *WireJson) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (w *WireJson) Decode(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}
