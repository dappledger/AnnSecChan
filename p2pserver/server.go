package p2pserver

import (
	"fmt"
	"path"

	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnSecChan/modules/common"
	"github.com/dappledger/AnnSecChan/modules/config"
	"github.com/dappledger/AnnSecChan/modules/kvdb"
	"github.com/dappledger/AnnSecChan/p2pserver/p2preactor"
	"github.com/dappledger/AnnSecChan/p2pserver/p2pswitch"
	"github.com/dappledger/AnnSecChan/p2pserver/types"
)

type P2PServer struct {
	cnf *types.P2PConfig
	Sw  *p2pswitch.P2PSwitch
}

func (s *P2PServer) initCnf(cfg *config.Config) (err error) {

	s.cnf = new(types.P2PConfig)

	bytPrivkey := common.Hex2Bytes(cfg.GetString("p2p_privkey"))

	if len(bytPrivkey) == crypto.PrivKeyLenEd25519 {
		var byt25519 [crypto.PrivKeyLenEd25519]byte
		copy(byt25519[:], bytPrivkey)
		s.cnf.PrivKey = crypto.PrivKeyEd25519(byt25519)
		s.cnf.CrypType = types.CrypType_Ed25519
	} else if len(bytPrivkey) == crypto.PrivKeyLenSecp256k1 {
		var byt256k1 [crypto.PrivKeyLenSecp256k1]byte
		copy(byt256k1[:], bytPrivkey)
		s.cnf.PrivKey = crypto.PrivKeySecp256k1(byt256k1)
		s.cnf.CrypType = types.CrypType_Secp256k1
	}

	s.cnf.LocalAddr = cfg.GetString("p2p_listen_addr")
	if s.cnf.LocalAddr == "" {
		return fmt.Errorf("you should config listen addr , such as : tcp://127.0.0.1:26657")
	}

	s.cnf.Moniker = cfg.GetString("p2p_moniker")

	dBPath := cfg.GetString("db_path")
	if dBPath == "" {
		dBPath = "./"
	}
	if s.cnf.TxDB, err = kvdb.NewLDBDatabase(path.Join(dBPath, "tx_data"), 1024, 256); err != nil {
		return err
	}

	s.cnf.Peers = common.Split(cfg.GetString("p2p_peers"), ",")

	refusePubkeys := common.Split(cfg.GetString("p2p_blacklist_pubkey"), ",")
	for _, refusePubkey := range refusePubkeys {
		if s.cnf.CrypType == types.CrypType_Ed25519 {
			var bytP25519 [crypto.PubKeyLenEd25519]byte
			copy(bytP25519[:], common.Hex2Bytes(refusePubkey))
			s.cnf.BlackListPubkeys = append(s.cnf.BlackListPubkeys, crypto.PubKeyEd25519(bytP25519))
		} else {
			var bytP256k1 [crypto.PubKeyLenSecp256k1]byte
			copy(bytP256k1[:], common.Hex2Bytes(refusePubkey))
			s.cnf.BlackListPubkeys = append(s.cnf.BlackListPubkeys, crypto.PubKeySecp256k1(bytP256k1))
		}
	}

	accessPubkeys := common.Split(cfg.GetString("p2p_whitelist_pubkey"), ",")
	for _, accessPubkey := range accessPubkeys {
		if s.cnf.CrypType == types.CrypType_Ed25519 {
			var bytP25519 [crypto.PubKeyLenEd25519]byte
			copy(bytP25519[:], common.Hex2Bytes(accessPubkey))
			s.cnf.WhiteListPubkeys = append(s.cnf.WhiteListPubkeys, crypto.PubKeyEd25519(bytP25519))
		} else {
			var bytP256k1 [crypto.PubKeyLenSecp256k1]byte
			copy(bytP256k1[:], common.Hex2Bytes(accessPubkey))
			s.cnf.WhiteListPubkeys = append(s.cnf.WhiteListPubkeys, crypto.PubKeySecp256k1(bytP256k1))
		}
	}

	return nil
}

func (s *P2PServer) OnStart(cfg *config.Config) (err error) {

	if err = s.initCnf(cfg); err != nil {
		return err
	}

	s.Sw = p2pswitch.NewP2PSwitch(s.cnf, []p2preactor.P2PReactor{p2preactor.NewP2PLegeReact(s.cnf, types.Reactor_Ledger_Name)})

	if err = s.Sw.OnStart(); err != nil {
		return err
	}

	return nil
}

func (s *P2PServer) Shutdown() {
	s.cnf.TxDB.Close()
	s.Sw.Shutdown()
}
