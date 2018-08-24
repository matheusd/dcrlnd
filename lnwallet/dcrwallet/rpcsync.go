package dcrwallet

import (
	"context"
	"encoding/hex"
	"sync"
	"time"

	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/rpcclient/v2"
	"github.com/decred/dcrd/wire"

	"github.com/decred/dcrlnd/lnwallet"

	"github.com/decred/dcrwallet/chain/v2"
	"github.com/decred/dcrwallet/errors"
)

var (
	// ErrOutputSpent is returned by the GetUtxo method if the target output
	// for lookup has already been spent.
	ErrOutputSpent = errors.New("target output has been spent")

	// ErrOutputNotFound signals that the desired output could not be
	// located.
	ErrOutputNotFound = errors.New("target output was not found")

	// ErrUnconnected is returned when an IO operation was requested by the
	// backend is not connected to the network.
	//
	// TODO(decred) this should probably be exported by lnwallet and
	// expected by the BlockChainIO interface.
	ErrUnconnected = errors.New("unconnected to the network")
)

// RPCSyncer implements the required methods for synchronizing a DcrWallet
// instance using a full node dcrd backend.
type RPCSyncer struct {
	cancel    func()
	rpcConfig rpcclient.ConnConfig
	net       *chaincfg.Params

	// mu is a mutex that protects the chain field.
	mu    sync.Mutex
	chain *chain.RPCClient
}

// Compile time check to ensure RPCSyncer fulfills lnwallet.BlockChainIO.
var _ lnwallet.BlockChainIO = (*RPCSyncer)(nil)

// NewRPCSyncer initializes a new syncer backed by a full dcrd node. It
// requires the config for reaching the dcrd instance and the corresponding
// network this instance should be in.
func NewRPCSyncer(rpcConfig rpcclient.ConnConfig, net *chaincfg.Params) (*RPCSyncer, error) {
	return &RPCSyncer{
		rpcConfig: rpcConfig,
		net:       net,
	}, nil
}

// GetBestBlock returns the current height and hash of the best known block
// within the main chain.
//
// This method is a part of the lnwallet.BlockChainIO interface.
func (s *RPCSyncer) GetBestBlock() (*chainhash.Hash, int32, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.chain == nil {
		return nil, 0, ErrUnconnected
	}
	hash, height, err := s.chain.GetBestBlock()
	return hash, int32(height), err
}

// GetUtxo returns the original output referenced by the passed outpoint that
// create the target pkScript.
//
// This method is a part of the lnwallet.BlockChainIO interface.
func (s *RPCSyncer) GetUtxo(op *wire.OutPoint, pkScript []byte,
	heightHint uint32, cancel <-chan struct{}) (*wire.TxOut, error) {

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.chain == nil {
		return nil, ErrUnconnected
	}

	txout, err := s.chain.GetTxOut(&op.Hash, op.Index, false)
	if err != nil {
		return nil, err
	} else if txout == nil {
		return nil, ErrOutputSpent
	}

	pkScript, err = hex.DecodeString(txout.ScriptPubKey.Hex)
	if err != nil {
		return nil, err
	}

	// Sadly, gettxout returns the output value in DCR instead of atoms.
	amt, err := dcrutil.NewAmount(txout.Value)
	if err != nil {
		return nil, err
	}

	return &wire.TxOut{
		Value:    int64(amt),
		PkScript: pkScript,
	}, nil
}

// GetBlock returns a raw block from the server given its hash.
//
// This method is a part of the lnwallet.BlockChainIO interface.
func (s *RPCSyncer) GetBlock(blockHash *chainhash.Hash) (*wire.MsgBlock, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.chain == nil {
		return nil, ErrUnconnected
	}
	return s.chain.GetBlock(blockHash)
}

// GetBlockHash returns the hash of the block in the best blockchain at the
// given height.
//
// This method is a part of the lnwallet.BlockChainIO interface.
func (s *RPCSyncer) GetBlockHash(blockHeight int64) (*chainhash.Hash, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.chain == nil {
		return nil, ErrUnconnected
	}
	return s.chain.GetBlockHash(blockHeight)
}

// start the syncer backend and begin synchronizing the given wallet.
func (s *RPCSyncer) start(w *DcrWallet) error {

	dcrwLog.Debugf("Starting rpc syncer")

	// A simple backoff function for the moment.
	backoff := func() {
		time.Sleep(time.Second * 5)
	}

	firstConnectionChan := make(chan struct{})

	var ctx context.Context
	ctx, s.cancel = context.WithCancel(context.Background())
	go func() {
		firstConnection := true

		// Maintain a running syncer while we haven't been commanded to
		// stop.
		for {
			select {
			case <-ctx.Done():
				// If the context has been canceled, stop
				// trying to connect.
				break
			default:
			}

			chainRPC, err := chain.NewRPCClient(s.net,
				s.rpcConfig.Host, s.rpcConfig.User,
				s.rpcConfig.Pass, s.rpcConfig.Certificates,
				false)
			if err != nil {
				dcrwLog.Errorf("Error creating dcrd client: %v ",
					err)
				backoff()
				continue
			}

			// Establish an RPC connection in addition to starting
			// the goroutines in the underlying wallet.
			err = chainRPC.Start(ctx, true)
			if err != nil {
				dcrwLog.Errorf("Error starting chain client: %v", err)
				backoff()
				continue
			}

			dcrwLog.Debugf("Reconnected to chain rpc backend")

			// Pass the rpc client into the wallet so it can sync
			// up to the current main chain.
			walletNetBackend := chain.BackendFromRPCClient(chainRPC.Client)
			w.wallet.SetNetworkBackend(walletNetBackend)
			w.loader.SetNetworkBackend(walletNetBackend)

			s.mu.Lock()
			s.chain = chainRPC
			s.mu.Unlock()

			// Close the initialConnect chan to signal the first
			// connection has been made.
			if firstConnection {
				close(firstConnectionChan)
				firstConnection = false
			}

			syncer := chain.NewRPCSyncer(w.wallet, chainRPC)
			syncer.SetNotifications(&chain.Notifications{
				Synced: w.onRPCSyncerSynced,
			})
			err = syncer.Run(ctx, true)

			if werr, is := err.(*errors.Error); is && werr.Err == context.Canceled {
				// This was a graceful shutdown, so ignore the error.
				dcrwLog.Debugf("RPCsyncer shutting down")
				return
			}

			// Set the wallet backend as nil, to signal we're
			// disconnected from the network.
			w.wallet.SetNetworkBackend(nil)
			w.loader.SetNetworkBackend(nil)
			s.mu.Lock()
			s.chain = nil
			s.mu.Unlock()

			dcrwLog.Errorf("Error during rpc sync: %v", err)
		}
	}()

	// Wait until we've been signalled the initial connection has
	// completed.
	<-firstConnectionChan

	return nil
}

func (s *RPCSyncer) stop() {
	dcrwLog.Debugf("RPCSyncer requested shutdown")
	s.cancel()
}
