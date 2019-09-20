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
	sw               *p2p.Switch
	reacts           []p2preactor.P2PReactor
	cnf              *types.P2PConfig
	wr               wire.Wire
	mapRecoveryTasks map[string]*types.RecoveryTask
	chanNotify       chan *types.ReactorNotify
	chanClose        chan struct{}
}

func NewP2PSwitch(cnf *types.P2PConfig, reacts []p2preactor.P2PReactor) *P2PSwitch {
	return &P2PSwitch{
		sw:               p2p.NewSwitch(viper.New()),
		reacts:           reacts,
		cnf:              cnf,
		wr:               new(wire.WireRlp),
		mapRecoveryTasks: make(map[string]*types.RecoveryTask, 0),
		chanNotify:       make(chan *types.ReactorNotify, 1024),
		chanClose:        make(chan struct{}),
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

//filter connect p2p node with blacklist or whitelist
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
	//p.backupSendMsg(pubKey, chanId, bMsg)
	return err
}

//handler local message with chanId,such as store payload in the local.
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

//get p2p peer nodes information
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

//send message to peer
func (p *P2PSwitch) sendMsg(sendPeer *p2p.Peer, chanId byte, bMsg []byte) error {
	if sendPeer.Send(chanId, bMsg) {
		return nil
	}
	return fmt.Errorf("msg send to %s error;", sendPeer.String())
}

func (p *P2PSwitch) genRecoveryKey(pubKey string) []byte {
	return append(types.Backup_LastKey_Name, common.Hex2Bytes(pubKey)...)
}

func (p *P2PSwitch) getRecoveryLastKey(pubKey string) []byte {
	key, err := p.cnf.TxDB.Get(p.genRecoveryKey(pubKey))
	if err != nil {
		return nil
	}
	return key
}

func (p *P2PSwitch) putRecoveryLastKey(pubKey string, key []byte) error {
	return p.cnf.TxDB.Put(p.genRecoveryKey(pubKey), key)
}

//add recoverytask , isRsume=true, recovery payload offset last failed key; false,recovery payload from begin.
func (p *P2PSwitch) AddRecoveryTask(pubKey string, isResume bool) {
	go func() {
		t := types.NewRecoveryTask(pubKey, isResume)
		p.mapRecoveryTasks[pubKey] = t
		if err := p.recoverySendMsg(t); err != nil {
			log.GetLog().LogError("recoverySendMsg failed , peer ", pubKey, "not connected")
			return
		}
		t.SetSuccess()
	}()
}

//stop one pubkey peer node recovery task
func (p *P2PSwitch) StopRecoveryTask(pubKey string) error {
	t, ok := p.mapRecoveryTasks[pubKey]
	if ok {
		t.Close()
		delete(p.mapRecoveryTasks, pubKey)
		return nil
	}
	return errors.New("task not exist")
}

//get recovery tasks
func (p *P2PSwitch) GetRecoveryTasks() []*types.RecoveryTask {
	var tasks []*types.RecoveryTask
	for _, t := range p.mapRecoveryTasks {
		tasks = append(tasks, t)
	}
	return tasks
}

//recovery payload send to peer node who lost data.
func (p *P2PSwitch) recoverySendMsg(t *types.RecoveryTask) error {
	sendP := p.sw.Peers().Get(t.PubKey)
	if sendP == nil {
		return fmt.Errorf("peer %s not connected", t.PubKey)
	}
	handler := func(k []byte, v []byte) error {
		select {
		case <-t.Wait():
			p.putRecoveryLastKey(t.PubKey, k)
			return errors.New("recoverySendMsg closed")
		case <-p.chanClose:
			t.SetFailed(k)
			p.putRecoveryLastKey(t.PubKey, k)
			return errors.New("p2p server closed")
		default:
			tryCount := 3
			for {
				msg := &types.LegerTransMsg{Key: common.Hash(v), Value: v}
				bMsg, err := p.wr.Encode(msg)
				if err != nil {
					return err
				}
				if err := p.sendMsg(sendP, types.Reactor_Ledger_ChanID, bMsg); err != nil {
					tryCount--
					if tryCount <= 0 {
						t.SetFailed(k)
						p.putRecoveryLastKey(t.PubKey, k)
						return err
					}
				}
				break
			}
			return nil
		}

	}
	if t.IsResume() {
		return p.cnf.TxDB.GetWithPrefixHandler(nil, p.getRecoveryLastKey(t.PubKey), handler)
	}
	return p.cnf.TxDB.GetWithPrefixHandler(nil, nil, handler)
}

//handler notify from reactor with chan.
func (p *P2PSwitch) routineHandlerNoticeFromReactor() {
	for {
		select {
		case <-p.chanNotify:
		case <-p.chanClose:
			log.GetLog().LogDebug("routine handler notify stop")
			return
		}
	}
}

//be active in getting message from reactor with reactName and optype.
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
