package api

import (
	"blocker/core"
	"encoding/gob"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (s *Server) HealthHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Healthy")
}

func (s *Server) GetHeightHandler(c echo.Context) error {
	height := s.chain.Height()
	return c.JSON(http.StatusOK, echo.Map{"height": int(height)})
}

type JSONBlock struct {
	Hash          string
	DataHash      string
	PrevBlockHash string
	Validator     string
	Signature     string
	Height        uint32
	Version       uint32
}

func (s *Server) GetBlockWithHeightHandler(c echo.Context) error {
	heightParam := c.QueryParam("height")
	height, err := strconv.Atoi(heightParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	block, err := s.chain.GetBlock(uint32(height))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	jsonBlock, err := toJSONBlock(block)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, jsonBlock)
}

func toJSONBlock(block *core.Block) (JSONBlock, error) {
	return JSONBlock{
		Hash:          block.Hash(core.BlockHasher{}).String(),
		DataHash:      block.DataHash.String(),
		PrevBlockHash: block.PrevBlockHash.String(),
		Validator:     block.Validator.Address().String(),
		Signature:     block.Signature.String(),
		Height:        block.Height,
		Version:       block.Version,
	}, nil
}

func (s *Server) SendTransactionHandler(c echo.Context) error {
	transaction := new(core.Transaction)
	if err := gob.NewDecoder(c.Request().Body).Decode(transaction); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	s.txChan <- transaction
	return nil
}
