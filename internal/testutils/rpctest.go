package testutils

import (
	"fmt"
	"time"

	"github.com/decred/dcrd/chaincfg/v2"
	"github.com/decred/dcrd/rpcclient/v5"
	"github.com/decred/dcrd/rpctest"
)

// NewSetupRPCTest attempts up to maxTries to setup an rpctest harness or
// errors. This is used to get around the fact the rpctest does not validate
// listening addresses beforehand and might try to listen on an used address in
// CI servers.
//
// The returned rpctest is already setup for use.
func NewSetupRPCTest(maxTries int, netParams *chaincfg.Params,
	handlers *rpcclient.NotificationHandlers,
	args []string, setupChain bool, numMatureOutputs uint32) (*rpctest.Harness, error) {

	var harness *rpctest.Harness
	var err error
	for i := 0; i < maxTries; i++ {
		harness, err = rpctest.New(netParams, handlers, args)
		if err == nil {
			err = harness.SetUp(setupChain, numMatureOutputs)
			if err == nil {
				return harness, nil
			} else {
				err = fmt.Errorf("unable to setup node: %v", err)
			}
		} else {
			err = fmt.Errorf("unable to create harness node: %v", err)
		}

		time.Sleep(time.Second)
	}

	return nil, err
}
