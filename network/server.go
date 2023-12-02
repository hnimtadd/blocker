package network

import (
	"blocker/core"
	"blocker/crypto"
	"bytes"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-kit/log"
)

const (
	defaultMaxPoolLen = 10
	defaultBlockTime  = 5 * time.Second
)

type ServerOptions struct {
	Logger        log.Logger
	Transport     Transport // LocalTransport of this server.
	RPCProcessor  RPCProcessor
	PrivKey       *crypto.PrivateKey
	RPCDecodeFunc RPCDecodeFunc
	ID            string
	LocalSeed     []Transport
	MaxPoolLen    int
	blockTime     time.Duration
	Version       uint32
}

type Server struct {
	peerCh   chan Transport
	memPool  *TxPool
	chain    *core.BlockChain
	peersMap map[NetAddr]Transport
	rpcCh    chan RPC
	quitCh   chan struct{}

	ServerOptions

	blockTime time.Duration

	lock        sync.RWMutex
	isValidator bool
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
	chain, err := core.NewBlockChain(core.NewGenesisBlock(), core.NewInMemoryStorage(), opts.Logger)
	if err != nil {
		return nil, err
	}

	sv := &Server{
		ServerOptions: opts,
		blockTime:     bt,
		peerCh:        make(chan Transport, 1024),
		memPool:       NewTxPool(opts.MaxPoolLen),
		chain:         chain,
		isValidator:   opts.PrivKey != nil,
		rpcCh:         make(chan RPC, 1024),
		quitCh:        make(chan struct{}, 1),
		peersMap:      make(map[NetAddr]Transport),
	}
	if sv.RPCDecodeFunc == nil {
		sv.RPCDecodeFunc = DefaultDecodeMessageFunc
	}

	if sv.RPCProcessor == nil {
		sv.RPCProcessor = sv
	}

	sv.Transport.SetPeerCh(sv.peerCh)
	if sv.isValidator {
		go sv.validatorLoop()
	}
	return sv, nil
}

func (s *Server) Start() {
	s.Logger.Log("msg", "start Server")
	go s.initTransport()
	go s.bootstrapNetwork()
free:
	for {
		select {
		case peer := <-s.peerCh:
			s.lock.Lock()
			if _, ok := s.peersMap[peer.Addr()]; ok {
				// s.Logger.Log("msg", "Peer exists", "addr", peer.Addr())
				s.lock.Unlock()
				continue
			}
			s.peersMap[peer.Addr()] = peer
			s.lock.Unlock()
			if err := peer.Connect(s.Transport); err != nil {
				s.Logger.Log("msg", "cannot connect to perr", "err", err.Error())
			}

			if err := s.sendGetStatusMessage(peer); err != nil {
				s.Logger.Log("error", err.Error())
				continue
			}
			s.Logger.Log("msg", "peer added to the server", "addr", peer.Addr())

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
				s.Logger.Log("err", err)
			}
		default:
			continue
		}
	}
}

func (s *Server) ProcessMessage(msg *DecodedMessage) error {
	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.processTransaction(msg.From, t)
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
	s.lock.RLock()
	defer s.lock.RUnlock()
	peer, ok := s.peersMap[to]
	if !ok {
		return fmt.Errorf("peer (%s) not connected", to)
	}
	if err := peer.Send(s.Transport.Addr(), msg); err != nil {
		s.Logger.Log("error", fmt.Sprintf("peer send error => addr %s [err: %s]", peer.Addr(), err))
		return err
	}
	return nil
}

func (s *Server) broadcast(msg []byte) error {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for addr, tr := range s.peersMap {
		if err := tr.Send(s.Transport.Addr(), msg); err != nil {
			s.Logger.Log("error", fmt.Sprintf("peer send error => addr %s [err: %s]", addr, err))
			return err
		}
	}
	return nil
}

func (s *Server) processTransaction(from NetAddr, tx *core.Transaction) error {
	if err := tx.Verify(); err != nil {
		return err
	}
	hash := tx.Hash(core.TxHasher{})

	if s.memPool.Contains(hash) {
		return nil
	}
	// set first timestamp when transaction seen locally
	tx.SetTimestamp(time.Now().UnixNano())

	// s.Logger.Log(
	// 	"msg", "new tx",
	// 	"hash", hash,
	// 	"Pending length", s.memPool.PendingCount(),
	// )

	// Need to broadcast this tx to peer
	//
	go func() {
		if err := s.broadcastTx(tx); err != nil {
			s.Logger.Log("error", err.Error())
		}
	}()

	s.memPool.Add(tx)
	return nil
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
	if err := s.chain.AddBlock(b); err != nil {
		// s.Logger.Log("error", err.Error())
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

func (s *Server) initTransport() {
	// s.Logger.Log("msg", fmt.Sprintf("listening on tranport (%s)", s.Transport.Addr()))
	for {
		rpc, ok := <-s.Transport.Consume()
		if ok {
			s.rpcCh <- rpc
		}
	}
}

func (s *Server) createNewBlock() error {
	currentHeader, err := s.chain.GetHeader(s.chain.Height())
	if err != nil {
		return err
	}

	// Should get out current pending transactions in queue
	txx := s.memPool.Pending()
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
	s.memPool.ClearPending()
	return nil
}

func (s *Server) bootstrapNetwork() {
	for _, peer := range s.LocalSeed {
		s.peerCh <- peer
	}
}

func (s *Server) sendGetStatusMessage(peer Transport) error {
	requestMessage := GetStatusMessage{
		ID: s.ID,
	}
	msg := NewMesage(MessageTypeRequestStatus, requestMessage.Bytes())
	return peer.Send(s.Transport.Addr(), msg.Bytes())
}
