// +build !no_routerrpc

package main

import (
	"context"
	"encoding/hex"

	"github.com/decred/dcrlnd/lnrpc/routerrpc"

	"github.com/urfave/cli"
)

var queryMissionControlCommand = cli.Command{
	Name:     "querymc",
	Category: "Payments",
	Usage:    "Query the internal mission control state.",
	Action:   actionDecorator(queryMissionControl),
}

func queryMissionControl(ctx *cli.Context) error {
	conn := getClientConn(ctx, false)
	defer conn.Close()

	client := routerrpc.NewRouterClient(conn)

	req := &routerrpc.QueryMissionControlRequest{}
	rpcCtx := context.Background()
	snapshot, err := client.QueryMissionControl(rpcCtx, req)
	if err != nil {
		return err
	}

	type displayPairHistory struct {
		NodeFrom, NodeTo                  string
		SuccessTime, FailTime             int64
		FailAmtAtoms, FailAmtMAtoms       int64
		SuccessAmtAtoms, SuccessAmtMAtoms int64
	}

	displayResp := struct {
		Pairs []displayPairHistory
	}{}

	for _, n := range snapshot.Pairs {
		displayResp.Pairs = append(
			displayResp.Pairs,
			displayPairHistory{
				NodeFrom:         hex.EncodeToString(n.NodeFrom),
				NodeTo:           hex.EncodeToString(n.NodeTo),
				FailTime:         n.History.FailTime,
				SuccessTime:      n.History.SuccessTime,
				FailAmtAtoms:     n.History.FailAmtAtoms,
				FailAmtMAtoms:    n.History.FailAmtMAtoms,
				SuccessAmtAtoms:  n.History.SuccessAmtAtoms,
				SuccessAmtMAtoms: n.History.SuccessAmtMAtoms,
			},
		)
	}

	printJSON(displayResp)

	return nil
}
