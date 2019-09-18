package p2preactor

import (
	"fmt"

	"github.com/dappledger/AnnChain/gemmill/go-wire"
	"github.com/dappledger/AnnChain/gemmill/p2p"
	"github.com/dappledger/AnnSecChan/modules/log"
	"github.com/dappledger/AnnSecChan/p2pserver/types"
	lwr "github.com/dappledger/AnnSecChan/p2pserver/wire"
)

type P2PLegeReact struct {
	cnf       *types.P2PConfig
	react     *LegeReact
	reactName string
}

func NewP2PLegeReact(cnf *types.P2PConfig, name string) *P2PLegeReact {
	return &P2PLegeReact{
		cnf:       cnf,
		reactName: name,
		react:     NewLegeReact(cnf, name),
	}
}

func (p *P2PLegeReact) GetReactName() string {
	return p.reactName
}

func (p *P2PLegeReact) GetReact() p2p.Reactor {
	return p.react
}

func (p *P2PLegeReact) ReactNotify(chanReact chan *types.ReactorNotify) {
	p.react.ChanNotify = chanReact
}

func (p *P2PLegeReact) GetData(op byte, params []byte) ([]byte, error) {
	switch op {
	case types.Reactor_Leger_Query_Tx:
		return p.cnf.TxDB.Get(params)
	default:
		return nil, fmt.Errorf("op %v not define", op)
	}
}

type LegeReact struct {
	p2p.BaseReactor
	cnf        *types.P2PConfig
	ChanNotify chan *types.ReactorNotify
	wr         lwr.Wire
}

func NewLegeReact(cnf *types.P2PConfig, name string) *LegeReact {
	br := new(LegeReact)
	br.BaseReactor = *p2p.NewBaseReactor(name, br)
	br.cnf = cnf
	br.wr = new(lwr.WireRlp)
	return br
}

func (b *LegeReact) OnStart() error {
	log.GetLog().LogDebug("LegeReact server OnStart")
	return b.BaseReactor.OnStart()
}

func (b *LegeReact) OnStop() {
	log.GetLog().LogDebug("LegeReact server OnStop")
	b.BaseReactor.OnStop()
}

func (b *LegeReact) AddPeer(peer *p2p.Peer) {
	log.GetLog().LogDebug("LegeReact server addPeer :", peer.String())
	b.sendNotify(types.Reactor_Leger_Notify_Add_Peer, peer.PubKey.KeyString())
}

func (b *LegeReact) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		&p2p.ChannelDescriptor{
			ID:                types.Reactor_Ledger_ChanID,
			Priority:          5,
			SendQueueCapacity: 100,
		},
	}
}

func (b *LegeReact) sendNotify(ntyp byte, msg interface{}) {
	nMsg := &types.ReactorNotify{RactorName: b.String(), NotifyType: ntyp, Message: msg}
	select {
	case b.ChanNotify <- nMsg:
	default:
		log.GetLog().LogError("LegeReact sendNotify chan full")
	}
}

func (b *LegeReact) RemovePeer(peer *p2p.Peer, reason interface{}) {
	log.GetLog().LogDebug("LegeReact server removePeer :", peer.String(), "reason:", reason)
}

func (b *LegeReact) Receive(chID byte, src *p2p.Peer, msgBytes []byte) {
	byt := make([]byte, 0)
	if err := wire.ReadBinaryBytes(msgBytes, &byt); err != nil {
		log.GetLog().LogError("receive date wire error:", err.Error(), src.String())
		return
	}
	switch chID {
	case types.Reactor_Ledger_ChanID:
		msg := new(types.LegerTransMsg)
		if err := b.wr.Decode(byt, msg); err != nil {
			log.GetLog().LogError("msg decode:", err.Error())
			return
		}
		if err := msg.CheckLegerTransMsg(); err != nil {
			log.GetLog().LogError("msg check failed:", err.Error())
			return
		}
		if err := b.cnf.TxDB.Put(msg.Key, msg.Value); err != nil {
			log.GetLog().LogError("DB put error:", err.Error())
			return
		}
		break
	}
}
