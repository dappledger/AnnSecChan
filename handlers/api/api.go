package api

import (
	"errors"
	"fmt"

	"github.com/dappledger/AnnSecChan/modules/common"
	"github.com/dappledger/AnnSecChan/modules/log"
	"github.com/dappledger/AnnSecChan/p2pserver"
	"github.com/dappledger/AnnSecChan/p2pserver/types"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	p2pServer *p2pserver.P2PServer
}

func NewHandlers(s *p2pserver.P2PServer) *Handlers {
	return &Handlers{
		p2pServer: s,
	}
}

func (h *Handlers) HandlerStartRecovery(c *gin.Context) {
	var (
		err        error
		reqRecover ReqStartRecoverTask
	)
	if err = c.BindJSON(&reqRecover); err != nil {
		goto errDeal
	}
	h.p2pServer.Sw.AddRecoveryTask(common.ToUpper(reqRecover.PubKey), reqRecover.IsResume)
	HandleSuccessMsg(c, "HandlerRecoveryPayload", "Success", "")
	return
errDeal:
	HandleErrorMsg(c, "HandlerRecoveryPayload", err.Error())
	return
}

func (h *Handlers) HandlerStopRecover(c *gin.Context) {
	var (
		err error
	)
	pubKey := c.Param("pubkey")
	if len(pubKey) <= 0 {
		err = errors.New("params error")
	}
	if err = h.p2pServer.Sw.StopRecoveryTask(common.ToUpper(pubKey)); err != nil {
		goto errDeal
	}
	HandleSuccessMsg(c, "HandlerStopRecover", "Success", "")
	return
errDeal:
	HandleErrorMsg(c, "HandlerStopRecover", err.Error())
	return
}

func (h *Handlers) HandlerGetRecover(c *gin.Context) {
	tasks := h.p2pServer.Sw.GetRecoveryTasks()
	HandleSuccessMsg(c, "HandlerGetRecover", "", tasks)
}

func (h *Handlers) HandlerNodePeers(c *gin.Context) {
	var err error
	peers := h.p2pServer.Sw.Peers()
	if len(peers) <= 0 {
		err = errors.New("no peers")
		goto errDeal
	}
	HandleSuccessMsg(c, "HandlerNodePeers", "", peers)
	return
errDeal:
	HandleErrorMsg(c, "HandlerNodePeers", err.Error())
	return
}

func (h *Handlers) HandlerTxPut(c *gin.Context) {
	var (
		err    error
		reqPut ReqPut
		msg    *types.LegerTransMsg
	)
	if err = c.BindJSON(&reqPut); err != nil {
		goto errDeal
	}

	msg = &types.LegerTransMsg{Key: common.Hash(reqPut.Value), Value: reqPut.Value}

	if err = h.multiSendMsg(reqPut.PubKeys, types.Reactor_Ledger_ChanID, msg); err != nil {
		goto errDeal
	}

	HandleSuccessMsg(c, "HandlerPut", "Success", common.Bytes2Hex(msg.Key))
	return
errDeal:
	HandleErrorMsg(c, "HandlerPut", err.Error())
	return
}

func (h *Handlers) HandlerTxGet(c *gin.Context) {
	var (
		err   error
		value []byte
	)
	if !common.IsHash(c.Param("key")) {
		err = fmt.Errorf("invalid key:%s", c.Param("key"))
		goto errDeal
	}
	if value, err = h.p2pServer.Sw.GetData(types.Reactor_Ledger_Name, types.Reactor_Leger_Query_Tx, common.Hex2Bytes(c.Param("key"))); err != nil {
		goto errDeal
	}
	HandleSuccessMsg(c, "HandlerGet", "", value)
	return
errDeal:
	HandleErrorMsg(c, "HandlerGet", err.Error())
	return
}

func (h *Handlers) OnClose() {}

//parallel sendmsg to p2p nodes
func (h *Handlers) multiSendMsg(pubKeys []string, chanId byte, msg interface{}) error {
	if len(pubKeys) <= 0 {
		return h.p2pServer.Sw.SendMsg("", chanId, msg)
	}
	chanResult := make(chan error, len(pubKeys))
	defer close(chanResult)

	for _, pubKey := range pubKeys {

		go func(pubKey string, chanId byte, msg interface{}, chanResult chan error) {
			chanResult <- h.p2pServer.Sw.SendMsg(pubKey, chanId, msg)
		}(common.ToUpper(pubKey), chanId, msg, chanResult)
	}
	strError := ""
	for i := 0; i < len(pubKeys); i++ {
		select {
		case err := <-chanResult:
			if err != nil {
				strError += err.Error()
			}
			break
		}
	}
	if len(strError) > 0 {
		return errors.New(strError)
	}
	return nil
}

func HandleSuccessMsg(c *gin.Context, requestType, msg string, data interface{}) {
	responseWrite(c, true, msg, data)
	logMsg := fmt.Sprintf("type[%s] From [%s] Params [%s]", requestType, c.Request.RemoteAddr, msg)
	log.GetLog().LogInfo(logMsg)
}

func HandleDebugMsg(c *gin.Context, requestType string, info string) {
	logMsg := fmt.Sprintf("type[%s] From [%s] Params [%s]", requestType, c.Request.RemoteAddr, info)
	log.GetLog().LogDebug(logMsg)
}
func HandleErrorMsg(c *gin.Context, requestType string, result string) {
	msg := fmt.Sprintf("type[%s] From [%s] Error [%s] ", requestType, c.Request.RemoteAddr, result)
	responseWrite(c, false, msg, "")
	log.GetLog().LogError(msg)
}
func responseWrite(ctx *gin.Context, isSuccess bool, result string, data interface{}) {
	ctx.JSON(200, gin.H{
		"success": isSuccess,
		"message": result,
		"data":    data,
	})
}
