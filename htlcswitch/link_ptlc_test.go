package htlcswitch

import (
	"testing"
	"time"

	"github.com/decred/dcrd/dcrutil/v3"
	"github.com/decred/dcrlnd/channeldb"
	"github.com/decred/dcrlnd/lnwire"
)

// TestPtlcChannelLinkSingleHopPayment in this test we checks the interaction
// between Alice and Bob within scope of one channel.
func TestPtlcChannelLinkSingleHopPayment(t *testing.T) {
	t.Parallel()
	const isPTLC = true

	// Setup a alice-bob network.
	alice, bob, cleanUp, err := createTwoClusterChannels(
		dcrutil.AtomsPerCoin*3,
		dcrutil.AtomsPerCoin*5)
	if err != nil {
		t.Fatalf("unable to create channel: %v", err)
	}
	defer cleanUp()

	n := newTwoHopNetwork(
		t, alice.channel, bob.channel, testStartingHeight,
	)
	if err := n.start(); err != nil {
		t.Fatal(err)
	}
	defer n.stop()

	aliceBandwidthBefore := n.aliceChannelLink.Bandwidth()
	bobBandwidthBefore := n.bobChannelLink.Bandwidth()

	debug := false
	if debug {
		// Log message that alice receives.
		n.aliceServer.intersect(createLogFunc("alice",
			n.aliceChannelLink.ChanID()))

		// Log message that bob receives.
		n.bobServer.intersect(createLogFunc("bob",
			n.bobChannelLink.ChanID()))
	}

	amount := lnwire.NewMAtomsFromAtoms(dcrutil.AtomsPerCoin)
	htlcAmt, totalTimelock, hops := generateHops(amount, testStartingHeight,
		n.bobChannelLink)

	// Wait for:
	// * HTLC add request to be sent to bob.
	// * alice<->bob commitment state to be updated.
	// * settle request to be sent back from bob to alice.
	// * alice<->bob commitment state to be updated.
	// * user notification to be sent.
	receiver := n.bobServer
	firstHop := n.bobChannelLink.ShortChanID()
	rhash, err := makePayment(
		n.aliceServer, receiver, firstHop, hops, amount, htlcAmt,
		totalTimelock, isPTLC,
	).Wait(30 * time.Second)
	if err != nil {
		t.Fatalf("unable to make the payment: %v", err)
	}

	// Wait for Alice to receive the revocation.
	//
	// TODO(roasbeef); replace with select over returned err chan
	time.Sleep(2 * time.Second)

	// Check that alice invoice was settled and bandwidth of HTLC links was
	// changed.
	invoice, err := receiver.registry.LookupInvoice(rhash)
	if err != nil {
		t.Fatalf("unable to get invoice: %v", err)
	}
	if invoice.State != channeldb.ContractSettled {
		t.Fatal("alice invoice wasn't settled")
	}

	if aliceBandwidthBefore-amount != n.aliceChannelLink.Bandwidth() {
		t.Fatal("alice bandwidth should have decrease on payment " +
			"amount")
	}

	if bobBandwidthBefore+amount != n.bobChannelLink.Bandwidth() {
		t.Fatalf("bob bandwidth isn't match: expected %v, got %v",
			bobBandwidthBefore+amount,
			n.bobChannelLink.Bandwidth())
	}
}
