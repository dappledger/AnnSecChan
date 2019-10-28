package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	gcrypto "github.com/dappledger/AnnChain/gemmill/go-crypto"
	"golang.org/x/crypto/curve25519"
)

func Encrypt(pubKey gcrypto.PubKey, msg []byte) ([]byte, error) {
	var (
		secret [32]byte
	)
	randEdPriv := gcrypto.GenPrivKeyEd25519()
	curve25519.ScalarMult(&secret, randEdPriv.ToCurve25519(), pubKey.(gcrypto.PubKeyEd25519).ToCurve25519())
	if err := aesEncode(secret, &msg); err != nil {
		return nil, err
	}
	randEdPub := randEdPriv.PubKey().Bytes()
	fmt.Println(len(randEdPub))
	randEdPub = append(randEdPub, msg...)
	return randEdPub, nil
}

func Decrypt(privKey gcrypto.PrivKey, msg []byte) ([]byte, error) {
	pubKey, err := gcrypto.PubKeyFromBytes(msg[:33])
	if err != nil {
		return nil, err
	}
	var secret [32]byte
	curve25519.ScalarMult(&secret, privKey.(gcrypto.PrivKeyEd25519).ToCurve25519(), pubKey.(gcrypto.PubKeyEd25519).ToCurve25519())
	xmsg := make([]byte, len(msg[33:]))
	copy(xmsg, msg[33:])
	if err := aesDecode(secret, &xmsg); err != nil {
		return nil, err
	}
	return xmsg, nil
}

func aesDecode(secret [32]byte, data *[]byte) error {
	var key []byte
	key = make([]byte, 32)
	copy(key[:], secret[:32])
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("aesEncode Error: %v", err)
	}
	stream := cipher.NewCFBDecrypter(block, secret[:block.BlockSize()])
	stream.XORKeyStream(*data, *data)
	return nil
}

func aesEncode(secret [32]byte, data *[]byte) error {
	var key []byte
	key = make([]byte, 32)
	copy(key[:], secret[:32])
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("aesEncode Error: %v", err)
	}
	stream := cipher.NewCFBEncrypter(block, secret[:block.BlockSize()])
	stream.XORKeyStream(*data, *data)
	return nil
}
