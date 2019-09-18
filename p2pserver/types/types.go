package types

import (
	"bytes"
	"fmt"

	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnSecChan/modules/common"
	"github.com/dappledger/AnnSecChan/modules/kvdb"
)

var (
	Reactor_Ledger_Name string = "ledgerReactor"

	Reactor_Ledger_ChanID byte = 0x01

	Reactor_Leger_Query_Tx byte = 0x01

	Reactor_Leger_Notify_Add_Peer byte = 0x01
)

var (
	CrypType_Ed25519   byte = 0x01
	CrypType_Secp256k1 byte = 0x02
)

type P2PConfig struct {
	PrivKey          crypto.PrivKey
	Moniker          string
	LocalAddr        string
	Peers            []string
	MaxPeers         uint32
	DialTimeOut      uint32
	HandShakeTimeOut uint32
	IsEncryption     bool
	BlackListPubkeys []crypto.PubKey
	WhiteListPubkeys []crypto.PubKey
	TxDB             kvdb.Database
	BkDB             kvdb.Database
	CrypType         byte
}

type Peer struct {
	Moniker string `json:"moniker"`
	Address string `json:"address"`
	PubKey  string `json:"pubkey"`
}

type BackUpMsg struct {
	ChanId byte
	Msg    []byte
}

type ReactorNotify struct {
	RactorName string
	NotifyType byte
	Message    interface{}
}

type LegerTransMsg struct {
	Key   []byte
	Value []byte
}

func (m *LegerTransMsg) CheckLegerTransMsg() error {
	if len(m.Key) != common.HASH_LEN {
		return fmt.Errorf("invalid key")
	}
	bytH := common.Hash(m.Value)
	if 0 != bytes.Compare(m.Key, bytH) {
		return fmt.Errorf("key and hash inconformity")
	}
	return nil
}

func (m *LegerTransMsg) String() string {
	return fmt.Sprintf("key:%s,value:%s", common.Bytes2Hex(m.Key), common.Bytes2Hex(m.Value))
}
