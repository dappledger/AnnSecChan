package common

import (
	"encoding/hex"
	"strings"

	"github.com/dappledger/AnnSecChan/modules/rlp"
	"golang.org/x/crypto/sha3"
)

const HASH_LEN = 32

func HexRemovePrefix(shex string) string {
	upHex := strings.ToUpper(shex)
	if strings.HasPrefix(upHex, "0X") {
		return shex[2:]
	}
	return shex
}

func Split(s, sep string) []string {
	if len(s) > 0 {
		return strings.Split(s, sep)
	}
	return []string{}
}

func Hex2Bytes(shex string) []byte {
	shex = HexRemovePrefix(shex)
	bhex, _ := hex.DecodeString(shex)
	return bhex
}

func IsHash(s string) bool {
	if len(Hex2Bytes(s)) == HASH_LEN {
		return true
	}
	return false
}

func CopyBytes(b []byte) (copiedBytes []byte) {
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)
	return
}

func Bytes2Hex(b []byte) string {
	return strings.ToUpper(hex.EncodeToString(b))
}

func ToUpper(s string) string {
	s = strings.ToUpper(s)
	if strings.HasPrefix(s, "0X") {
		return s[2:]
	}
	return s
}

func ParseListenAddress(s string) (string, string) {
	arry := strings.Split(s, "://")
	if len(arry) != 2 {
		return "", ""
	}
	return arry[0], arry[1]
}

func Hash(v interface{}) []byte {
	h := make([]byte, HASH_LEN)
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, v)
	hw.Sum(h[:0])
	return h
}
