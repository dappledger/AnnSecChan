package p2pswitch

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/p2p"
	"github.com/dappledger/AnnSecChan/modules/common"
	"github.com/dappledger/AnnSecChan/modules/log"
	"github.com/dappledger/AnnSecChan/p2pserver/p2preactor"
	"github.com/dappledger/AnnSecChan/p2pserver/types"
	"github.com/dappledger/AnnSecChan/p2pserver/wire"
	"github.com/spf13/viper"
)

type P2PSwitch struct {
	sw         *p2p.Switch
	reacts     []p2preactor.P2PReactor
	cnf        *types.P2PConfig
	wr         wire.Wire
	chanNotify chan *types.ReactorNotify
	chanClose  chan struct{}
}

func NewP2PSwitch(cnf *types.P2PConfig, reacts []p2preactor.P2PReactor) *P2PSwitch {
	return &P2PSwitch{
		sw:         p2p.NewSwitch(viper.New()),
		reacts:     reacts,
		cnf:        cnf,
		wr:         new(wire.WireRlp),
		chanNotify: make(chan *types.ReactorNotify, 1024),
		chanClose:  make(chan struct{}),
	}
}

func (p *P2PSwitch) OnStart() error {

	if err := p.prepareP2PSwitch(); err != nil {
		return err
	}
	if _, err := p.sw.Start(); err != nil {
		return err
	}
	return nil
}

func (p *P2PSwitch) prepareP2PSwitch() error {

	protol, address := common.ParseListenAddress(p.cnf.LocalAddr)

	if len(protol) == 0 || len(address) == 0 {
		return errors.New("p2p_listen_addr is wrong format")
	}

	for _, ract := range p.reacts {
		p.sw.AddReactor(ract.GetReactName(), ract.GetReact())
		ract.ReactNotify(p.chanNotify)
	}

	listen, err := p2p.NewDefaultListener(protol, address, true)
	if err != nil {
		return err
	}

	p.sw.AddListener(listen)

	nodeInfo := &p2p.NodeInfo{
		PubKey:     p.cnf.PrivKey.PubKey(),
		Moniker:    p.cnf.Moniker,
		ListenAddr: listen.ExternalAddress().String(),
		Version:    "1.0.1",
	}
	p.sw.SetNodeInfo(nodeInfo)
	p.sw.SetNodePrivKey(p.cnf.PrivKey)
	p.sw.SetPubKeyFilter(p.funcPubKeyFilter)
	p.sw.DialSeeds(p.cnf.Peers)
	go p.routineHandlerNoticeFromReactor()
	return nil
}

func (p *P2PSwitch) funcPubKeyFilter(pubkey crypto.PubKey) error {
	if len(p.cnf.WhiteListPubkeys) > 0 {
		for _, accessPubkey := range p.cnf.WhiteListPubkeys {
			if !pubkey.Equals(accessPubkey) {
				log.GetLog().LogWarn("refuse connect pubkey:", pubkey.KeyString())
				return fmt.Errorf("refuse connect pubkey")
			}
		}
	} else {
		for _, refusePubkey := range p.cnf.BlackListPubkeys {
			if pubkey.Equals(refusePubkey) {
				log.GetLog().LogWarn("refuse connect pubkey:", pubkey.KeyString())
				return fmt.Errorf("refuse connect pubkey")
			}
		}
	}
	return nil
}

func (p *P2PSwitch) SendMsg(pubKey string, chanId byte, msg interface{}) (err error) {
	if len(pubKey) <= 0 {
		if err = p.localHandler(chanId, msg); err != nil {
			return err
		}
		return nil
	}
	bMsg, err := p.wr.Encode(msg)
	if err != nil {
		return err
	}
	sendP := p.sw.Peers().Get(pubKey)
	if sendP == nil {
		err = fmt.Errorf("peer %s is not connect", pubKey)
		goto errDeal
	}
	if err = p.sendMsg(sendP, chanId, bMsg); err != nil {
		goto errDeal
	}
	if err = p.localHandler(chanId, msg); err != nil {
		return err
	}
	return nil
errDeal:
	p.backupSendMsg(pubKey, chanId, bMsg)
	return err
}

func (p *P2PSwitch) localHandler(chanId byte, msg interface{}) error {
	switch chanId {
	case types.Reactor_Ledger_ChanID:
		if reflect.TypeOf(msg) != reflect.TypeOf(&types.LegerTransMsg{}) {
			return fmt.Errorf("localhandler wrong msg type :%s", reflect.TypeOf(msg).String())
		}
		if err := msg.(*types.LegerTransMsg).CheckLegerTransMsg(); err != nil {
			log.GetLog().LogError("localhandler msg check failed:", err.Error())
			return err
		}
		if err := p.cnf.TxDB.Put(msg.(*types.LegerTransMsg).Key, msg.(*types.LegerTransMsg).Value); err != nil {
			log.GetLog().LogError("localhandler DB put error:", err.Error())
			return err
		}
		return nil
	default:
		return fmt.Errorf("localhandler wrong chanid %v", chanId)
	}
}

func (p *P2PSwitch) Peers() (peers []*types.Peer) {
	for _, peer := range p.sw.Peers().List() {
		pr := &types.Peer{
			Moniker: peer.Moniker,
			Address: peer.ListenAddr,
			PubKey:  peer.Key,
		}
		peers = append(peers, pr)
	}
	return
}

func (p *P2PSwitch) sendMsg(sendPeer *p2p.Peer, chanId byte, bMsg []byte) error {
	if sendPeer.Send(chanId, bMsg) {
		return nil
	}
	return fmt.Errorf("msg send to %s error;", sendPeer.String())
}

func (p *P2PSwitch) backupSendMsg(pubKey string, chanId byte, bMsg []byte) error {

	var (
		bacKey  []byte
		bkValue []byte
	)
	bkValue = append(bkValue, chanId)
	bkValue = append(bkValue, bMsg...)

	bacKey = append(common.Hex2Bytes(pubKey), common.Hash(bkValue)...)
	if err := p.cnf.BkDB.Put(bacKey, bkValue); err != nil {
		log.GetLog().LogError("backupsendmsg put error:", err.Error(), pubKey, chanId)
	}
	log.GetLog().LogError("SendMsg Error We Backup", pubKey, chanId)
	return nil
}

func (p *P2PSwitch) restoreSendMsg(pubKey string) error {
	sendP := p.sw.Peers().Get(pubKey)
	if sendP == nil {
		log.GetLog().LogError("restoreSendMsg failed,peer ", pubKey, "not connected")
		return fmt.Errorf("peer %s not connected", pubKey)
	}
	handler := func(k []byte, v []byte) error {
		if err := p.sendMsg(sendP, v[0], v[1:]); err != nil {
			log.GetLog().LogError("restoreSendMsg sendMsg error:", err.Error(), common.Bytes2Hex(k))
			return err
		}
		if err := p.cnf.BkDB.Delete(k); err != nil {
			log.GetLog().LogError("restoreSendMsg delete key:", common.Bytes2Hex(k), "error:", err.Error())
		}
		log.GetLog().LogDebug("restoreSendMsg success key:", common.Bytes2Hex(k))
		return nil
	}
	return p.cnf.BkDB.GetWithPrefixHandler(common.Hex2Bytes(pubKey), handler)
}

func (p *P2PSwitch) routineHandlerNoticeFromReactor() {
	for {
		select {
		case msg := <-p.chanNotify:
			p.handlerNotify(msg)
		case <-p.chanClose:
			log.GetLog().LogDebug("routine handler notify stop")
			return
		}
	}
}

func (p *P2PSwitch) handlerNotify(msg interface{}) {
	if !(reflect.TypeOf(&types.ReactorNotify{}) == reflect.TypeOf(msg)) {
		log.GetLog().LogError("handlerNotify not NotifyMsg type")
		return
	}
	nMsg := msg.(*types.ReactorNotify)
	switch nMsg.NotifyType {
	case types.Reactor_Leger_Notify_Add_Peer:
		switch nMsg.Message.(type) {
		case string:
			go p.restoreSendMsg(nMsg.Message.(string))
		default:
			log.GetLog().LogError("handlerNotify wrong params", nMsg.Message)
		}
	default:
		log.GetLog().LogError("handlerNotify wrong NotifyType", nMsg.NotifyType)
	}
}

func (p *P2PSwitch) GetData(reactName string, op byte, params []byte) ([]byte, error) {
	for _, react := range p.reacts {
		if react.GetReactName() == reactName {
			return react.GetData(op, params)
		}
	}
	return nil, fmt.Errorf("reactor not exist:%s", reactName)
}

func (p *P2PSwitch) BroadCase(chanId byte, msg interface{}) error {
	var (
		count      int
		falseCount int
	)
	chanResults := p.sw.Broadcast(chanId, msg)
	l := len(p.sw.Peers().List())
	for {
		if count >= l {
			break
		}
		result := <-chanResults
		count++
		if result == false {
			falseCount++
		}
	}
	if falseCount != 0 {
		return fmt.Errorf("broadCase send error peer count:", falseCount)
	}
	return nil
}

func (p *P2PSwitch) Shutdown() {
	close(p.chanClose)
	p.sw.Stop()
}
