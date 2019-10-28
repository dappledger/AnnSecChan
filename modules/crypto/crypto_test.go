package crypto

import (
	"encoding/hex"
	"testing"

	gcrypto "github.com/dappledger/AnnChain/gemmill/go-crypto"
)

func TestEncrypto(t *testing.T) {
	msg := []byte("12345678fghfghdfgh")
	var bytP25519 [gcrypto.PubKeyLenEd25519]byte
	byt, _ := hex.DecodeString("F2C3EB310C71FF594BD67E158E5AA4D6B860F81DC5ADDA1FCB6CED6758FC12C7")
	copy(bytP25519[:], byt)
	sbyt, err := Encrypt(gcrypto.PubKeyEd25519(bytP25519), msg)
	t.Log(string(sbyt), err)

	bytPrivkey, _ := hex.DecodeString("39BDA4399C6DC81EC7D878E74E6D12D6370FACACE45E22FC7C4C2396E4056FCFF2C3EB310C71FF594BD67E158E5AA4D6B860F81DC5ADDA1FCB6CED6758FC12C7")
	var byt25519 [gcrypto.PrivKeyLenEd25519]byte
	copy(byt25519[:], bytPrivkey)
	dbyt, err := Decrypt(gcrypto.PrivKeyEd25519(byt25519), sbyt)
	t.Log(string(dbyt), err)
}
