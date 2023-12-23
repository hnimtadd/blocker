package api

import (
	"blocker/core"
	"blocker/crypto"
	"blocker/pool"
	"blocker/types"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
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
	Txx           []string
	Height        uint32
	Version       uint32
}

func (s *Server) GetBlockWithHeightHandler(c echo.Context) error {
	heightParam := c.QueryParam("height")
	if heightParam == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"msg": "height is empty"})
	}
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
	txx := []string{}
	for _, tx := range block.Transactions {
		txx = append(txx, tx.Hash(core.TxHasher{}).String())
	}

	// genesis block
	if block.Height == 0 {
		return JSONBlock{
			Hash:          block.Hash(core.BlockHasher{}).String(),
			DataHash:      block.DataHash.String(),
			PrevBlockHash: block.PrevBlockHash.String(),
			Validator:     "genesis",
			Signature:     "genesis",
			Txx:           txx,
			Height:        block.Height,
			Version:       block.Version,
		}, nil
	}

	return JSONBlock{
		Hash:          block.Hash(core.BlockHasher{}).String(),
		DataHash:      block.DataHash.String(),
		PrevBlockHash: block.PrevBlockHash.String(),
		Validator:     block.Validator.Address().String(),
		Signature:     block.Signature.String(),
		Txx:           txx,
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

type TransactionJSON struct {
	Data    map[string]any `json:"tx_data"`
	Hash    string         `json:"hash"`
	BlockID string         `json:"block"`
	Status  string         `json:"status"`
	Type    string         `json:"tx_type"`
}

// TransactionData returns transaction type and relevant data of that transaction type
func TransactionData(tx *core.Transaction) (string, map[string]any) {
	var txType string
	var data map[string]any
	switch ttx := tx.TxInner.(type) {
	case core.MintTx:
		switch nft := ttx.NFT.(type) {
		case core.NFTAsset:
			data = map[string]any{
				"hash":     ttx.Hash(core.TxMintHasher{}).String(),
				"metadata": ttx.Metadata,
				"nft_data": nft.Data,
				"nft_type": nft.Type,
			}
			txType = string(core.TxTypeMint)
		case core.NFTCollection:
			data = map[string]any{
				"hash":            ttx.Hash(core.TxMintHasher{}).String(),
				"metadata":        ttx.Metadata,
				"collection_type": nft.Type,
			}
			txType = string(core.TxTypeMint)
		}
	default:
		dataHash := sha256.Sum256(tx.Data)
		data = map[string]any{
			"data": types.HashFromBytes(dataHash[:]).String(),
		}
		txType = string(core.TxTypeNative)
	}
	return txType, data
}

func TransactionJSONFromPoolWithStatus(st pool.TxPoolStatus, tx *core.Transaction) (*TransactionJSON, error) {
	txType, data := TransactionData(tx)
	return &TransactionJSON{
		Status:  string(st),
		Hash:    tx.Hash(core.TxHasher{}).String(),
		BlockID: "",
		Type:    txType,
		Data:    data,
	}, nil
}

func TransactionJSONWithStatus(st core.Status, tx *core.Transaction, b *core.Block) (*TransactionJSON, error) {
	txType, data := TransactionData(tx)
	return &TransactionJSON{
		Status:  string(st),
		Hash:    tx.Hash(core.TxHasher{}).String(),
		BlockID: b.Hash(core.BlockHasher{}).String(),
		Type:    txType,
		Data:    data,
	}, nil
}

func (s *Server) GetTransactionWithHashHandler(c echo.Context) error {
	hashString := c.Param("hash")
	if hashString == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"errors": "invalid hash"})
	}
	// check from txPool if tx is denided

	hashBytes, err := hex.DecodeString(hashString)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"errors": fmt.Sprintf("cannot decode hash given hahs, (%s)", err.Error())})
	}
	poolStatus, tx, err := s.TxPool.Get(types.HashFromBytes(hashBytes))
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("cannot get transaction informatin: %s", err.Error()))
	}
	switch poolStatus {
	case pool.TxPoolReceived:
		status, b, tx, err := s.chain.GetTransaction(types.HashFromBytes(hashBytes))
		if err != nil {
			return c.String(http.StatusNotFound, fmt.Sprintf("cannot get transaction information: (%s)", err.Error()))
		}

		jsonTx, err := TransactionJSONWithStatus(status, tx, b)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("cannot get transaction information: (%s)", err.Error()))
		}
		return c.JSON(http.StatusOK, jsonTx)
	default:
		jsonTx, err := TransactionJSONFromPoolWithStatus(poolStatus, tx)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("cannot get transaction information: %s", err.Error()))
		}
		return c.JSON(http.StatusOK, jsonTx)
	}
}

func (s *Server) CheckTransactionStatus(c echo.Context) error {
	hashString := c.Param("hash")
	if hashString == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"errors": "invalid hash"})
	}

	hashBytes, err := hex.DecodeString(hashString)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"errors": fmt.Sprintf("cannot decode hash given hahs, (%s)", err.Error())})
	}
	poolStatus, tx, err := s.TxPool.Get(types.HashFromBytes(hashBytes))
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("cannot get transaction informatin: %s", err.Error()))
	}

	jsonTx, err := TransactionJSONFromPoolWithStatus(poolStatus, tx)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("cannot get transaction information: %s", err.Error()))
	}
	return c.JSON(http.StatusOK, jsonTx)
}

func (s *Server) RegisterNewAccountStateHandler(c echo.Context) error {
	pubKey := new(crypto.PublicKey)
	if err := gob.NewDecoder(c.Request().Body).Decode(pubKey); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	if err := s.chain.PutNewAccount(pubKey); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	return nil
}

func (s *Server) GetAccountStateHandler(c echo.Context) error {
	addrString := c.Param("hash")

	if addrString == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"errors": "invalid hash"})
	}
	// check from txPool if tx is denided
	addrBytes, err := hex.DecodeString(addrString)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"errors": fmt.Sprintf("cannot decode hash given hahs, (%s)", err.Error())})
	}
	addr := types.AddressFromBytes(addrBytes)

	state, fromTxx, toTxx, err := s.chain.GetAccount(addr)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("cannot get account state from blockchain: %s", err.Error()))
	}

	fromTXXString := []string{}
	for _, tx := range fromTxx {
		fromTXXString = append(fromTXXString, tx.Hash(core.TxHasher{}).String())
	}
	toTxxString := []string{}
	for _, tx := range toTxx {
		toTxxString = append(toTxxString, tx.Hash(core.TxHasher{}).String())
	}
	fmt.Println("==================")
	fmt.Println(state)
	fmt.Println("==================")
	return c.JSON(
		http.StatusOK, echo.Map{
			"state": echo.Map{
				"addr":    state.Addr.String(),
				"balance": state.Balance,
				"nonce":   state.Nonce,
			},
			"outcomeTransactions": fromTXXString,
			"incomeTransactions":  toTxxString,
		})
}
