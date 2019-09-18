package p2preactor

import (
	"github.com/dappledger/AnnChain/gemmill/p2p"
	"github.com/dappledger/AnnSecChan/p2pserver/types"
)

type P2PReactor interface {
	GetReactName() string
	GetReact() p2p.Reactor
	GetData(op byte, params []byte) ([]byte, error)
	ReactNotify(chan *types.ReactorNotify)
}
