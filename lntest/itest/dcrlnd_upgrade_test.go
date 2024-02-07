package itest

import (
	"os"
	"time"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/decred/dcrlnd/lnrpc"
	"github.com/decred/dcrlnd/lnrpc/routerrpc"
	"github.com/decred/dcrlnd/lntest"
	"github.com/decred/dcrlnd/sweep"
	"github.com/stretchr/testify/require"
	"matheusd.com/testctx"
)

func lndOldItestVersionBin() string {
	return os.Getenv("DCRLND_MIGITEST_OLDBIN")
}

var oldVersionInteractiontests = []*testCase{
	{
		name: "old version interactions",
		test: testOldVersionInteractions,
	},
}

// testOldVersionInteractions tests interactions between an old version and the
// new version.
func testOldVersionInteractions(net *lntest.NetworkHarness, t *harnessTest) {
	oldVersionBin := lndOldItestVersionBin()
	if oldVersionBin == "" {
		t.t.Skip("DCRLND_MIGITEST_OLDBIN is not set")
	}

	// Alice will always run the old node. Bob will always run the new node.
	// Charlie will start using the old node, then it will upgrade to the
	// new node halfway through the test.
	alice := net.NewNode(t.t, "alice", nil, lntest.WithLndBinary(oldVersionBin))
	bob := net.NewNode(t.t, "bob", nil)
	charlie := net.NewNode(t.t, "charlie", nil, lntest.WithLndBinary(oldVersionBin))
	defer shutdownAndAssert(net, t, alice)
	defer shutdownAndAssert(net, t, bob)
	defer shutdownAndAssert(net, t, charlie)

	net.ConnectNodes(t.t, charlie, alice)
	net.ConnectNodes(t.t, charlie, bob)

	// charlieAmt will be a running amount of how much charlie should have
	// at the end of the test.
	charlieAmt := dcrutil.Amount(1e8)

	// Various amounts.
	chanAmt := dcrutil.Amount(1e6)
	pushAmt := dcrutil.Amount(1e5)
	payAmt := dcrutil.Amount(25)
	net.SendCoins(t.t, charlieAmt, charlie)
	net.SendCoins(t.t, charlieAmt, alice)
	net.SendCoins(t.t, charlieAmt, bob)

	// These are channels that will be force-closed by Charlie after he
	// upgrades, at the end of the test.
	toForceCloseAtEnd := map[string]struct{}{}

	// A single sweep tx may be used to sweep multiple closed channels,
	// so keep track of them on a map to dedupe and avoid double counting.
	closeSweeps := make(map[chainhash.Hash]dcrutil.Amount)
	addSweepsFromClose := func(closeTx *chainhash.Hash) {
		t.t.Helper()
		sweepTxs := findSweepTxsInNode(t, charlie, closeTx)
		if len(sweepTxs) == 0 {
			t.Fatalf("Did not find sweep for closed tx %s", closeTx)
		}
		for txh := range sweepTxs {
			if _, ok := closeSweeps[txh]; ok {
				continue
			}

			closeSweeps[txh] = getTxFeeFromId(t, &txh)
		}

	}

	// We'll run this loop twice: on the first iteration, Charlie is at the
	// old version, and on the second iteration he's using the new version.
	//
	// We test opening, sending payments, coop-closing and force-closing
	// with both Charlie and the other nodes initiating (to test all
	// possible combinations).
	for i := 0; i < 2; i++ {
		// These are channels that will be force-closed before
		// suspending Charlie. This verifies persistence of the
		// close state during migration.
		toForceCloseBeforeSuspend := map[string]bool{}

		// These are channels that will be force-closed by the other
		// node, while Charlie is offline. This verifies detection of
		// force-closed channels during migration.
		toForceCloseWhileOffline := map[string]map[string]bool{
			"alice": {},
			"bob":   {},
		}

		// Open one channel from and one channel to each of the other
		// nodes.
		src := charlie
		for _, other := range []*lntest.HarnessNode{alice, bob} {
			otherName := other.Name()
			for i := 0; i < 5*2; i++ {
				// Bump the values so that everything is not
				// symmetrical.
				// pushAmt += 1
				// payAmt += 1

				chanPoint := openChannelAndAssert(t, net, src, other,
					lntest.OpenChannelParams{Amt: chanAmt, PushAmt: pushAmt})
				// t.Logf("Opened %s -> %s %s", src.Name(), other.Name(), chanPointToStr(t, chanPoint))
				fundingTxFee := getTxFeeFromId(t, chanPointTxHash(t, chanPoint))

				// Sanity check the funding tx fee.
				require.LessOrEqual(t.t, fundingTxFee, dcrutil.Amount(10000))

				// Send at least one payment through this
				// channel so that it advances its state.
				payReqs, _, _, err := createPayReqs(other, payAmt, 1)
				require.NoError(t.t, err)
				sendPayReq := &routerrpc.SendPaymentRequest{
					PaymentRequest: payReqs[0],
					OutgoingChanId: getShortChannelID(t, src, chanPoint).ToUint64(),
					TimeoutSeconds: 60,
				}
				payStream, err := src.RouterClient.SendPaymentV2(testctx.New(t), sendPayReq)
				require.NoError(t.t, err)
				for {
					res, err := payStream.Recv()
					require.NoError(t.t, err)
					if res.Status == lnrpc.Payment_SUCCEEDED {
						break
					} else if res.Status == lnrpc.Payment_FAILED {
						t.t.Fatalf("Payment from %s to %s failed",
							src.Name(), other.Name())
					}
				}

				// When Charlie is opening, he pays the funding
				// tx fee, push amount and pay amount. When
				// he's the target, he receives the pushed
				// amount and the payment amount.
				charlieInitiator := i%2 == 0
				if charlieInitiator {
					charlieAmt -= fundingTxFee
					charlieAmt -= pushAmt
					charlieAmt -= payAmt
				} else {
					charlieAmt += pushAmt
					charlieAmt += payAmt
				}

				// Depending on the set of channels being
				// opened, we'll track them to do some
				// force-closes later on.
				strChanPoint := chanPointToStr(t, chanPoint)
				switch {
				case i/2 == 1:
					toForceCloseAtEnd[strChanPoint] = struct{}{}
				case i/2 == 2:
					toForceCloseBeforeSuspend[strChanPoint] = charlieInitiator
				case i/2 == 3:
					toForceCloseWhileOffline[otherName][strChanPoint] = charlieInitiator
				case i/2 == 4:
					// Coop close this channel.
					closeTx := closeChannelAndAssert(t, net, charlie, chanPoint, false)
					if charlieInitiator {
						closeTxFee := getTxFeeFromId(t, closeTx)
						charlieAmt -= closeTxFee
					}
				}

				// Switch around who is opening the channel.
				src, other = other, src
			}
		}

		// Suspend the other nodes to do some force closes.
		aliceRestart, err := net.SuspendNode(alice)
		require.NoError(t.t, err)
		bobRestart, err := net.SuspendNode(bob)
		require.NoError(t.t, err)

		// Before restarting, force close (but do not cleanup) a few
		// channels. This verifies the upgrade did not break
		// persistence of closed channel tracking.
		var forceCloseTxs []*chainhash.Hash
		for strChanPoint, charlieInitiator := range toForceCloseBeforeSuspend {
			// t.Logf("Trying to close before charlie suspend %s", strChanPoint)

			chanPoint := strToChainPoint(t, strChanPoint)
			_, closeTx, err := net.CloseChannel(charlie, chanPoint, true)
			require.NoError(t.t, err)

			forceCloseTxs = append(forceCloseTxs, closeTx)
			if charlieInitiator {
				closeTxFee := getTxFeeFromId(t, closeTx)
				charlieAmt -= closeTxFee
				// t.Logf("Closed with tx %s fee %d", closeTx, closeTxFee)
			} else {
				// t.Logf("Closed with tx %s", closeTx)
			}
		}

		// Suspend Charlie and restart the other nodes.
		charlieRestart, err := net.SuspendNode(charlie)
		require.NoError(t.t, err)
		aliceRestart()
		bobRestart()

		// Before restarting, force-close channels from other nodes and
		// mine blocks. This verifies the upgrade did not break
		// detection of closed channels.
		for otherName, toForceClose := range toForceCloseWhileOffline {
			other := alice
			if otherName == "bob" {
				other = bob
			}

			for strChanPoint, charlieInitiator := range toForceClose {
				// t.Logf("Trying to close %s %s", other.Name(), strChanPoint)
				chanPoint := strToChainPoint(t, strChanPoint)
				_, closeTx, err := net.CloseChannel(other, chanPoint, true)
				require.NoError(t.t, err)
				forceCloseTxs = append(forceCloseTxs, closeTx)
				if charlieInitiator {
					closeTxFee := getTxFeeFromId(t, closeTx)
					charlieAmt -= closeTxFee
				}
			}
		}

		// Ensure other nodes sweep.
		mineBlocks(t, net, defaultCSV-1, 0)
		time.Sleep(sweep.DefaultBatchWindowDuration + time.Second)
		mineBlocks(t, net, 3, 0)

		// Set Charlie to use the new version and restart it.
		charlie.Cfg.LndBinary = t.getLndBinary()
		charlie.Cfg.NeedsFilteredArgs = false
		err = charlieRestart()
		require.NoError(t.t, err)

		// After the sweep duration, Charlie should be sweeping from
		// the force closed channels (because enough blocks have
		// already been mined, even for the ones he force-closed before
		// shutting down).
		time.Sleep(sweep.DefaultBatchWindowDuration + time.Second)
		mineBlocks(t, net, 1, 0)
		for _, closeTx := range forceCloseTxs {
			addSweepsFromClose(closeTx)
		}
	}

	// Close all of Charlie's outstanding channels.
	chans, err := charlie.ListChannels(testctx.New(t), &lnrpc.ListChannelsRequest{})
	require.NoError(t.t, err)
	for _, openChan := range chans.Channels {
		// Some channels are force-closed, others are coop closed.
		chanPoint := strToChainPoint(t, openChan.ChannelPoint)
		_, force := toForceCloseAtEnd[openChan.ChannelPoint]
		// t.Logf("Closing channel %s (force=%v)", openChan.ChannelPoint, force)
		closeTx := closeChannelAndAssert(t, net, charlie, chanPoint, force)

		if !force && openChan.Initiator {
			closeTxFee := getTxFeeFromId(t, closeTx)
			charlieAmt -= closeTxFee
		} else if force && openChan.Initiator {
			charlieAmt -= dcrutil.Amount(openChan.CommitFee)
		}

		// Cooperative closes already send the funds to Charlie without
		// the need for an additional sweep.
		if !force {
			continue
		}
		cleanupForceClose(t, net, charlie, chanPoint)

		// Determine the fee paid by charlie to sweep the force close
		// transaction.
		addSweepsFromClose(closeTx)
	}

	// Finally, deduct the cost to sweep the force-closed channels.
	mineBlocks(t, net, 1, 0)
	assertNumPendingChannels(t, charlie, 0, 0, 0, 0)
	assertNodeNumChannels(t, charlie, 0)
	for _, sweepFee := range closeSweeps {
		charlieAmt -= sweepFee
	}

	// We should get the correct final on-chain balance on Charlie.
	finalBal, err := charlie.WalletBalance(testctx.New(t), &lnrpc.WalletBalanceRequest{})
	require.NoError(t.t, err)
	require.Equal(t.t, int64(charlieAmt), finalBal.ConfirmedBalance)
}
