package lntest

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/rpcclient/v2"
	"github.com/decred/dcrd/rpctest"
)

// logDir is the name of the temporary log directory.
const logDir = "./.backendlogs"

// DcrdBackendConfig is an implementation of the BackendConfig interface
// backed by a btcd node.
type DcrdBackendConfig struct {
	// RPCConfig houses the connection config to the backing btcd instance.
	RPCConfig rpcclient.ConnConfig

	// P2PAddress is the p2p address of the btcd instance.
	P2PAddress string
}

// GenArgs returns the arguments needed to be passed to LND at startup for
// using this node as a chain backend.
func (b DcrdBackendConfig) GenArgs() []string {
	var args []string
	encodedCert := hex.EncodeToString(b.RPCConfig.Certificates)
	args = append(args, fmt.Sprintf("--dcrd.rpchost=%v", b.RPCConfig.Host))
	args = append(args, fmt.Sprintf("--dcrd.rpcuser=%v", b.RPCConfig.User))
	args = append(args, fmt.Sprintf("--dcrd.rpcpass=%v", b.RPCConfig.Pass))
	args = append(args, fmt.Sprintf("--dcrd.rawrpccert=%v", encodedCert))

	return args
}

// P2PAddr returns the address of this node to be used when connection over the
// Bitcoin P2P network.
func (b DcrdBackendConfig) P2PAddr() string {
	return b.P2PAddress
}

// NewDcrdBackend starts a new rpctest.Harness and returns a DcrdBackendConfig
// for that node.
func NewDcrdBackend() (*DcrdBackendConfig, func(), error) {
	args := []string{
		"--rejectnonstd",
		"--txindex",
		"--trickleinterval=100ms",
		"--debuglevel=debug",
		"--logdir=" + logDir,
	}
	netParams := &chaincfg.SimNetParams
	chainBackend, err := rpctest.New(netParams, nil, args)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create dcrd node: %v", err)
	}

	if err := chainBackend.SetUp(false, 0); err != nil {
		return nil, nil, fmt.Errorf("unable to set up dcrd backend: %v", err)
	}

	bd := &DcrdBackendConfig{
		RPCConfig: chainBackend.RPCConfig(),
		// P2PAddress: chainBackend.P2PAddress(),
	}

	cleanUp := func() {
		chainBackend.TearDown()

		// After shutting down the chain backend, we'll make a copy of
		// the log file before deleting the temporary log dir.
		logFile := logDir + "/" + netParams.Name + "/dcrd.log"
		err := CopyFile("./output_dcrd_chainbackend.log", logFile)
		if err != nil {
			fmt.Printf("unable to copy file: %v\n", err)
		}
		if err = os.RemoveAll(logDir); err != nil {
			fmt.Printf("Cannot remove dir %s: %v\n", logDir, err)
		}
	}

	return bd, cleanUp, nil
}
