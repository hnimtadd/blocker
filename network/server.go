package network

import (
	"blocker/api"
	"blocker/core"
	"blocker/crypto"
	"blocker/pool"
	"blocker/types"
	"bytes"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/go-kit/log"
)

const (
	defaultMaxPoolLen = 50
	defaultBlockTime  = 5 * time.Second
)

type ServerOptions struct {
	// RootAccount   core.Account
	Transport     Transport
	Logger        log.Logger
	Addr          string // if addr not empty, mean that this node could be API server
	RPCProcessor  RPCProcessor
	Listener      net.Listener
	PrivKey       *crypto.PrivateKey
	RPCDecodeFunc RPCDecodeFunc
	ID            string
	TCPSeed       []string
	LocalSeed     []Peer
	MaxPoolLen    int
	blockTime     time.Duration
	Version       uint32
}

type Server struct {
	peerCh        <-chan Peer // Read-only peer chan from transport
	rpcCh         <-chan RPC  // Read-only rpc chan from Transport
	chain         *core.BlockChain
	memPool       *pool.TxPool
	quitCh        chan struct{}
	Storage       core.Storage // cached storage, different than chain storage
	txChan        chan *core.Transaction
	ServerOptions               // Embed ServerOptions
	blockTime     time.Duration // duration of generating new blokc
	isValidator   bool
}

func NewServer(opts ServerOptions) (*Server, error) {
	bt := opts.blockTime
	if bt == 0 {
		bt = defaultBlockTime
	}

	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stderr)
		opts.Logger = log.With(opts.Logger, "ID", opts.ID)
	}

	if opts.MaxPoolLen == 0 {
		opts.MaxPoolLen = defaultMaxPoolLen
	}
	chain, err := core.NewBlockChain(Genesis(), core.NewInMemoryStorage(), opts.Logger)
	if err != nil {
		return nil, err
	}
	sv := &Server{
		ServerOptions: opts,
		blockTime:     bt,
		Storage:       core.NewInMemoryStorage(),
		memPool:       pool.NewTxPool(opts.MaxPoolLen),
		chain:         chain,
		isValidator:   opts.PrivKey != nil,
		quitCh:        make(chan struct{}, 1),
		txChan:        make(chan *core.Transaction, 1024),
	}
	if sv.RPCDecodeFunc == nil {
		sv.RPCDecodeFunc = DefaultDecodeMessageFunc
	}

	if sv.RPCProcessor == nil {
		sv.RPCProcessor = sv
	}

	if sv.isValidator {
		go sv.validatorLoop()
	}

	if sv.Transport != nil {
		sv.rpcCh = sv.Transport.ConsumeRPC()
		sv.peerCh = sv.Transport.ConsumePeer()
	}

	if sv.Addr != "" {
		// Init API Server with this server
		opts := api.ServerOpts{
			Addr: sv.Addr,
		}
		apiServer := api.NewServer(sv.chain, sv.txChan, opts)
		go func() {
			panic(fmt.Sprintf("error while serving JSON: %v", apiServer.Start()))
		}()
	}
	return sv, nil
}

func (s *Server) Start() {
	s.Logger.Log("msg", "start Server")
	// go s.bootstrapNetwork()
	go s.bootstrapTCPNetwork()
free:
	for {
		select {
		case peer := <-s.peerCh:
			if err := s.sendGetStatusMessage(peer); err != nil {
				s.Logger.Log("error", err.Error())
				continue
			}
			s.Logger.Log("msg", "peer added to the server", "addr", peer.Addr())

		case tx := <-s.txChan:
			if err := s.processTransaction(tx); err != nil {
				s.Logger.Log("process tx error", err)
			}

		case rpc := <-s.rpcCh:
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("error", err, "from", rpc.From)
				continue
			}
			if err := s.RPCProcessor.ProcessMessage(msg); err != nil {
				// s.Logger.Log("error", err, "from", rpc.From)
				continue
			}
		case <-s.quitCh:
			break free
		}
	}

	fmt.Println("Server down")
}

func (s *Server) validatorLoop() {
	ticker := time.NewTicker(s.blockTime)

	s.Logger.Log("msg", "start validator loop", "blockTime", s.blockTime)
	for {
		select {
		case <-ticker.C:
			if err := s.createNewBlock(); err != nil {
				s.Logger.Log("error", err)
				// panic(err)
			}
		default:
			continue
		}
	}
}

func (s *Server) ProcessMessage(msg *DecodedMessage) error {
	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.processTransaction(t)
	case *core.Block:
		return s.processBlock(t)
	case *RequestBlocksMessage:
		return s.processRequestBlocksMessage(msg.From, t)
	case *ResponseBlocksMessage:
		return s.processResponseBlocksMessage(msg.From, t)
	case *GetStatusMessage:
		return s.processGetStatusMessage(msg.From)
	case *StatusMessage:
		return s.processStatusMessage(msg.From, t)
	}
	return nil
}

func (s *Server) send(to NetAddr, msg []byte) error {
	return s.Transport.Send(to, msg)
}

func (s *Server) broadcast(msg []byte) error {
	return s.Transport.Broadcast(msg)
}

func (s *Server) processTransaction(tx *core.Transaction) error {
	if err := tx.Verify(); err != nil {
		return err
	}
	hash := tx.Hash(core.TxHasher{})

	if s.memPool.Contains(hash) {
		return nil
	}
	// set first timestamp when transaction seen locally
	tx.SetTimestamp(time.Now().UnixNano())

	go func() {
		if err := s.broadcastTx(tx); err != nil {
			s.Logger.Log("error", err.Error())
		}
	}()

	return s.memPool.Add(tx)
}

func (s *Server) broadcastTx(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}
	msg := NewMesage(MessageTypeTx, buf.Bytes())
	return s.broadcast(msg.Bytes())
}

func (s *Server) processBlock(b *core.Block) error {
	// TODO: if the block is higher than the current blockchain, sync block, add currently txx in the orphan block to into the memPool again
	if err := s.chain.AddBlock(b); err != nil {
		return err
	}

	go func() {
		if err := s.broadcastBlock(b); err != nil {
			s.Logger.Log("error", err.Error())
		}
	}()

	return nil
}

func (s *Server) broadcastBlock(b *core.Block) error {
	buf := &bytes.Buffer{}
	if err := b.Encode(core.NewGobBlockEncoder(buf)); err != nil {
		return err
	}
	msg := NewMesage(MessageTypeBlock, buf.Bytes())
	return s.broadcast(msg.Bytes())
}

func (s *Server) processRequestBlocksMessage(from NetAddr, data *RequestBlocksMessage) error {
	if data.To == 0 {
		data.To = s.chain.Height()
	}
	blocks := []*core.Block{}
	for i := data.From; i <= data.To; i++ {
		// check if block exists
		if s.chain.HasBlock(i) {
			b, err := s.chain.GetBlock(i)
			if err != nil {
				return err
			}
			blocks = append(blocks, b)
		}
	}
	rsp := &ResponseBlocksMessage{
		Blocks: blocks,
	}
	msg := NewMesage(MessageTypeResponseBlocks, rsp.Bytes())

	return s.broadcast(msg.Bytes())
}

func (s *Server) processResponseBlocksMessage(from NetAddr, msg *ResponseBlocksMessage) error {
	s.Logger.Log("msg", "Received respone blocks message", "from", from)
	for _, b := range msg.Blocks {
		if err := s.chain.AddBlock(b); err != nil {
			// s.Logger.Log("error", err)
			return nil
		}
	}
	return nil
}

func (s *Server) processGetStatusMessage(from NetAddr) error {
	status := StatusMessage{
		ID:            s.ID,
		Version:       s.Version,
		CurrentHeight: s.chain.Height(),
	}
	msg := NewMesage(MessageTypeResponseStatus, status.Bytes())
	return s.send(from, msg.Bytes())
}

func (s *Server) processStatusMessage(from NetAddr, data *StatusMessage) error {
	if s.Version != data.Version {
		s.Version = data.Version
	}

	if s.chain.Height() != data.CurrentHeight {
		// current chain have lower height with other peer, should fetch
		req := RequestBlocksMessage{
			From: s.chain.Height() + 1,
			To:   data.CurrentHeight,
		}
		msg := NewMesage(MessageTypeRequestBlocks, req.Bytes())
		if err := s.send(from, msg.Bytes()); err != nil {
			s.Logger.Log("msg", fmt.Sprintf("cannot send msg to (%s), err: (%s)", from, err.Error()))
		}
	}
	return nil
}

func (s *Server) createNewBlock() error {
	// minter work
	currentHeader, err := s.chain.GetHeader(s.chain.Height())
	if err != nil {
		return err
	}

	// Should get out current pending transactions in queue
	// pending and remove pending maybe remove unprocessed transaction
	txx := s.memPool.Pending()
	s.memPool.LockPending()
	fmt.Println("==========PENDING-TRANSACTION=============")
	for _, tx := range txx {
		s.Logger.Log("pending", tx.String())
	}
	fmt.Println("==========END-PENDING-TRANSACTION=============")
	idxx := s.chain.SoftcheckTransactions(txx)
	if len(idxx) > 0 {
		denidedTXX := s.memPool.Denide(idxx)

		s.memPool.UnlockPending()
		txx = s.memPool.Pending()
		s.memPool.LockPending()

		fmt.Println("==========DENIDED-TRANSACTION=============")
		for _, tx := range denidedTXX {
			s.Logger.Log("denided", tx.String())
		}

		fmt.Println("==========END-DENIDED-TRANSACTION=============")

		fmt.Println("==========PROCESSED-TRANSACTION=============")
		for _, tx := range txx {
			s.Logger.Log("processed", tx.String())
		}
		fmt.Println("==========END-PROCESSED-TRANSACTION=============")

		fmt.Println("==========ACCOUNT-STATE==========")
		fmt.Println(s.chain.AccountState())
		fmt.Println("==========END-ACCOUNT-STATE==========")
	}

	block, err := core.NewBlockFromPrevHeader(currentHeader, txx)
	if err != nil {
		return err
	}

	if err := block.Sign(s.PrivKey); err != nil {
		return err
	}

	if err := s.chain.AddBlock(block); err != nil {
		return err
	}
	go func() {
		if err := s.broadcastBlock(block); err != nil {
			s.Logger.Log("error", err.Error())
		}
	}()

	s.memPool.Processed(txx)
	return nil
}

func (s *Server) bootstrapTCPNetwork() {
	for _, addr := range s.TCPSeed {
		s.Transport.Dial(NetAddr(addr))
	}
}

func (s *Server) sendGetStatusMessage(toPeer Peer) error {
	requestMessage := GetStatusMessage{
		ID: s.ID,
	}
	msg := NewMesage(MessageTypeRequestStatus, requestMessage.Bytes())
	s.Logger.Log("action", "send get status message", "to", toPeer.Addr())
	return s.Transport.Send(toPeer.Addr(), msg.Bytes())
}

func Genesis() *core.Block {
	// coinbase := core.Account{}
	transferTx := core.TransferTx{
		From:  types.Address{},
		To:    types.Address{},
		Value: 1000000,
	}

	tx := &core.Transaction{
		TxInner: transferTx,
		Nonce:   0,
	}
	block := &core.Block{
		Header: &core.Header{
			Version:       1,
			PrevBlockHash: types.Hash{},
			Height:        0,
			Timestamp:     00000000,
		},
		Transactions: []*core.Transaction{
			tx,
		},
	}
	return block
}
