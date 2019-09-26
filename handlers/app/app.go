package app

import (
	"fmt"
	"net/http"

	"github.com/dappledger/AnnSecChan/handlers/api"
	"github.com/dappledger/AnnSecChan/modules/config"
	"github.com/dappledger/AnnSecChan/modules/log"
	"github.com/dappledger/AnnSecChan/p2pserver"
	gin "github.com/gin-gonic/gin"
)

type ConfigData struct {
	Port     string
	LogDir   string
	LogLevel int
}

type App struct {
	handlers  *api.Handlers
	p2pServer *p2pserver.P2PServer
	cnf       *ConfigData
}

func (app *App) initFlag(c *config.Config) (err error) {
	app.cnf = new(ConfigData)
	app.cnf.Port = c.GetString("port")
	app.cnf.LogDir = c.GetString("log_dir")
	app.cnf.LogLevel = int(c.GetFloat("log_level"))
	return
}

func (app *App) OnStart(c *config.Config) error {
	if err := app.initFlag(c); err != nil {
		return err
	}
	if _, err := log.NewLog(app.cnf.LogDir, "node", app.cnf.LogLevel); err != nil {
		return err
	}
	app.p2pServer = new(p2pserver.P2PServer)
	if err := app.p2pServer.OnStart(c); err != nil {
		return err
	}
	app.handlers = api.NewHandlers(app.p2pServer)
	router := gin.Default()
	v1 := router.Group("/v1")
	{
		v1.GET("/transaction/:key", app.handlers.HandlerTxGet)
		v1.PUT("/transaction", app.handlers.HandlerTxPut)
		v1.PUT("/transaction/withsignature", app.handlers.HandlerTxPutWithSign)

		v1.GET("/node/peers", app.handlers.HandlerNodePeers)

		v1.POST("/recovery/start", app.handlers.HandlerStartRecovery)
		v1.GET("/recovery/stop/:pubkey", app.handlers.HandlerStopRecover)
		v1.GET("/recovery/show", app.handlers.HandlerGetRecover)
	}
	fmt.Println("Listen:", app.cnf.Port)
	http.ListenAndServe(":"+app.cnf.Port, router)
	return nil
}

func (app *App) Shutdown() {
	app.p2pServer.Shutdown()
	app.handlers.OnClose()
	fmt.Println("server shutdown")
}
