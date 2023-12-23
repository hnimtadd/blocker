package api

import (
	"blocker/core"
	"blocker/pool"

	"github.com/labstack/echo/v4"
)

type ServerOpts struct {
	TxPool *pool.TxPool
	Addr   string
}

type Server struct {
	// ECHO server, serve JSON api
	chain  *core.BlockChain
	txChan chan<- *core.Transaction
	ServerOpts
}

func NewServer(bc *core.BlockChain, txChan chan<- *core.Transaction, opts ServerOpts) *Server {
	sv := &Server{
		chain:      bc,
		ServerOpts: opts,
		txChan:     txChan,
	}
	return sv
}

func (s *Server) initRoute() *echo.Echo {
	app := echo.New()
	app.GET("/health", s.HealthHandler)
	app.GET("/api/height", s.GetHeightHandler)
	app.GET("/api/block", s.GetBlockWithHeightHandler)
	app.POST("/api/tx", s.SendTransactionHandler)
	app.GET("/api/tx/:hash", s.GetTransactionWithHashHandler)
	app.POST("/api/account/register", s.RegisterNewAccountStateHandler)
	app.GET("/api/account/summary/:hash", s.GetAccountStateSummaryHandler)
	app.GET("/api/account/state/:hash", s.GetAccountStateHandler)
	return app
}

func (s *Server) Start() error {
	app := s.initRoute()
	return app.Start(s.Addr)
}
