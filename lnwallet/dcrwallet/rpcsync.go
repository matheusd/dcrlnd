package dcrwallet

import (
	"context"
	"sync"

	"github.com/decred/dcrd/chaincfg/v2"
	"github.com/decred/dcrd/rpcclient/v5"

	"github.com/decred/dcrwallet/chain/v3"
	"github.com/decred/dcrwallet/errors/v2"
)

// RPCSyncer implements the required methods for synchronizing a DcrWallet
// instance using a full node dcrd backend.
type RPCSyncer struct {
	cancel    func()
	rpcConfig rpcclient.ConnConfig
	net       *chaincfg.Params
	wg        sync.WaitGroup
}

// NewRPCSyncer initializes a new syncer backed by a full dcrd node. It
// requires the config for reaching the dcrd instance and the corresponding
// network this instance should be in.
func NewRPCSyncer(rpcConfig rpcclient.ConnConfig, net *chaincfg.Params) (*RPCSyncer, error) {
	return &RPCSyncer{
		rpcConfig: rpcConfig,
		net:       net,
	}, nil
}

// start the syncer backend and begin synchronizing the given wallet.
func (s *RPCSyncer) start(w *DcrWallet) error {

	dcrwLog.Debugf("Starting rpc syncer")

	// This context will be canceled by `w` once its Stop() method is
	// called.
	var ctx context.Context
	ctx, s.cancel = context.WithCancel(context.Background())

	chainRpcOpts := chain.RPCOptions{
		Address: s.rpcConfig.Host,
		User:    s.rpcConfig.User,
		Pass:    s.rpcConfig.Pass,
		CA:      s.rpcConfig.Certificates,
	}
	syncer := chain.NewSyncer(w.wallet, &chainRpcOpts)
	syncer.SetCallbacks(&chain.Callbacks{
		Synced: w.onRPCSyncerSynced,
	})

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		err := syncer.Run(ctx)
		dcrwLog.Debugf("RPCsyncer shutting down")

		// TODO: convert to errors.Is
		if werr, is := err.(*errors.Error); is && werr.Err == context.Canceled {
			// This was a graceful shutdown, so ignore the error.
			dcrwLog.Debugf("RPCsyncer shutting down")
			return
		}

		dcrwLog.Errorf("RPCSyncer error: %v", err)
	}()

	return nil
}

func (s *RPCSyncer) stop() {
	dcrwLog.Debugf("RPCSyncer requested shutdown")
	s.cancel()
}

func (s *RPCSyncer) waitForShutdown() {
	s.wg.Wait()
}
