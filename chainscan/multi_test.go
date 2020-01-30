package chainscan

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/decred/dcrd/wire"
)

// scannerTestCase is a test case that must be fulfilled by both the historical
// and the tip watcher individually.
type scannerTestCase struct {
	name      string
	target    func(*testBlock) Target
	manglers  []blockMangler
	wantFound bool
	wantMF    MatchField
}

var scannerTestCases = []scannerTestCase{

	// Basic tests where a block should match the target.

	{
		name: "ConfirmedScript",
		target: func(b *testBlock) Target {
			return ConfirmedScript(0, testPkScript)
		},
		manglers: []blockMangler{
			confirmScript(testPkScript),
			cfilterData(testPkScript),
		},
		wantFound: true,
		wantMF:    MatchTxOut,
	},

	{
		name: "ConfirmedOutPoint",
		target: func(b *testBlock) Target {
			outp := wire.OutPoint{
				Hash:  b.block.Transactions[0].TxHash(),
				Index: 1,
			}
			return ConfirmedOutPoint(outp, 0, testPkScript)
		},
		manglers: []blockMangler{
			confirmScript(testPkScript),
			cfilterData(testPkScript),
		},
		wantFound: true,
		wantMF:    MatchTxOut,
	},

	{
		name: "SpentScript",
		target: func(b *testBlock) Target {
			return SpentScript(0, testPkScript)
		},
		manglers: []blockMangler{
			spendScript(testSigScript),
			cfilterData(testPkScript),
		},
		wantFound: true,
		wantMF:    MatchTxIn,
	},

	{
		name: "SpentOutPoint",
		target: func(b *testBlock) Target {
			outp := b.block.Transactions[0].TxIn[1].PreviousOutPoint
			return SpentOutPoint(outp, 0, testPkScript)
		},
		manglers: []blockMangler{
			spendOutPoint(testOutPoint),
			cfilterData(testPkScript),
		},
		wantFound: true,
		wantMF:    MatchTxIn,
	},

	// This test forces something that would ordinarily match in the block
	// to not be included in the cfilter. This is not a realistic scenario
	// in practice (due to consensus rules enforcing the correct behavior)
	// but shows that if the cfilter doesn't match the block itself isn't
	// tested.
	{
		name: "ConfirmedScript with cfilter miss",
		target: func(b *testBlock) Target {
			return ConfirmedScript(0, testPkScript)
		},
		manglers: []blockMangler{
			confirmScript(testPkScript),
		},
	},

	// The rest of the tests all force trigger a block check since a
	// cfilter miss is trivial.

	{
		name: "ConfirmedScript without match",
		target: func(b *testBlock) Target {
			return ConfirmedScript(0, testPkScript)
		},
		manglers: []blockMangler{
			// Note testPkScript is _not_ confirmed
			cfilterData(testPkScript),
		},
	},

	{
		name: "ConfirmedOutPoint without match",
		target: func(b *testBlock) Target {
			outp := wire.OutPoint{
				Hash:  b.block.Transactions[0].TxHash(),
				Index: 1,
			}
			return ConfirmedOutPoint(outp, 0, testPkScript)
		},
		manglers: []blockMangler{
			cfilterData(testPkScript),
		},
	},

	// This tests that trying to watch for a specific outpoint fails when
	// the script is confirmed in a _different_ outpoint.
	{
		name: "ConfirmedOutPoint with different outpoint",
		target: func(b *testBlock) Target {
			outp := wire.OutPoint{
				Hash:  b.block.Transactions[0].TxHash(),
				Index: 0,
			}
			return ConfirmedOutPoint(outp, 0, testPkScript)
		},
		manglers: []blockMangler{
			confirmScript(testPkScript),
			cfilterData(testPkScript),
		},
	},

	{
		name: "SpentScript without match",
		target: func(b *testBlock) Target {
			return SpentScript(0, testPkScript)
		},
		manglers: []blockMangler{
			cfilterData(testPkScript),
		},
	},

	{
		name: "SpentOutPoint without match",
		target: func(b *testBlock) Target {
			outp := wire.OutPoint{
				Hash:  b.block.Transactions[0].TxHash(),
				Index: 1,
			}
			return SpentOutPoint(outp, 0, testPkScript)
		},
		manglers: []blockMangler{
			cfilterData(testPkScript),
		},
	},

	// This tests that a match is triggered even when the signatureScript
	// of a watched outpoint does not correspond to the requested pkscript
	// to watch for.
	//
	// Note that this is technically a client error given that watching for
	// an outpoint when its correspondong pkscript is not the one specified
	// in SpentOutpoint might cause the cfilter to never trigger a block
	// download.
	//
	// Nevertheless we provide this test to fixate the TipWatcher's
	// behavior in this situation.
	{
		name: "SpentOutPoint with different script",
		target: func(b *testBlock) Target {
			outp := b.block.Transactions[0].TxIn[1].PreviousOutPoint
			return SpentOutPoint(outp, 0, testPkScript)
		},
		manglers: []blockMangler{
			spendScript([]byte{0x00}),
			cfilterData(testPkScript),
		},
		wantFound: true,
		wantMF:    MatchTxIn,
	},

	// This asserts that a match is not triggered for a confirmation when
	// the script is spent in a random input.
	{
		name: "ConfirmedScript with spent script",
		target: func(b *testBlock) Target {
			return ConfirmedScript(0, testPkScript)
		},
		manglers: []blockMangler{
			spendScript(testSigScript),
			cfilterData(testPkScript),
		},
	},

	// The next tests all deal with transactions in the stake tree.

	// This test asserts that spending an output which is in the stake tree
	// gets detected.
	{
		name: "Spent Stake OutPoint",
		target: func(b *testBlock) Target {
			outp := b.block.Transactions[0].TxIn[1].PreviousOutPoint
			return SpentOutPoint(outp, 0, testPkScript)
		},
		manglers: []blockMangler{
			spendOutPoint(wire.OutPoint{
				Hash: testOutPoint.Hash,
				Tree: wire.TxTreeStake,
			}),
			cfilterData(testPkScript),
		},
		wantFound: true,
		wantMF:    MatchTxIn,
	},

	{
		name: "ConfirmedOutPoint in stake tx",
		target: func(b *testBlock) Target {
			outp := wire.OutPoint{
				Hash:  b.block.STransactions[0].TxHash(),
				Index: 1,
				Tree:  wire.TxTreeStake,
			}
			return ConfirmedOutPoint(outp, 0, testPkScript)
		},
		manglers: []blockMangler{
			confirmScript(testPkScript),
			cfilterData(testPkScript),
			moveRegularToStakeTree(),
		},
		wantFound: true,
		wantMF:    MatchTxOut,
	},

	{
		name: "ConfirmedScript in stake tx",
		target: func(b *testBlock) Target {
			return ConfirmedScript(0, testPkScript)
		},
		manglers: []blockMangler{
			confirmScript(testPkScript),
			cfilterData(testPkScript),
			moveRegularToStakeTree(),
		},
		wantFound: true,
		wantMF:    MatchTxOut,
	},

	// This test ensures that trying to watch for a script which should be
	// confirmed in a specific output in the regular transaction tree fails
	// to trigger a found event when that same script is actually confirmed
	// in the stake tree.
	{
		name: "ConfirmedOutPoint in wrong tx tree",
		target: func(b *testBlock) Target {
			outp := wire.OutPoint{
				Hash:  b.block.STransactions[0].TxHash(),
				Index: 1,
				Tree:  wire.TxTreeRegular,
			}
			return ConfirmedOutPoint(outp, 0, testPkScript)
		},
		manglers: []blockMangler{
			confirmScript(testPkScript),
			cfilterData(testPkScript),
			moveRegularToStakeTree(),
		},
	},

	{
		name: "ConfirmedScript with large script",
		target: func(b *testBlock) Target {
			return ConfirmedScript(0, bytes.Repeat([]byte{0x55}, 128))
		},
		manglers: []blockMangler{
			confirmScript(bytes.Repeat([]byte{0x55}, 128)),
			cfilterData(bytes.Repeat([]byte{0x55}, 128)),
		},
		wantFound: true,
		wantMF:    MatchTxOut,
	},

	{
		name: "ConfirmedScript with large script without match",
		target: func(b *testBlock) Target {
			return ConfirmedScript(0, bytes.Repeat([]byte{0x55}, 128))
		},
		manglers: []blockMangler{
			confirmScript(bytes.Repeat([]byte{0x55}, 127)),
			cfilterData(bytes.Repeat([]byte{0x55}, 128)),
		},
	},
}

// TestSimultaneousScanners tests that when running both a TipWatcher and a
// Historical rescan against the same chain, events are consistent with the
// expected behavior of both scanners.
func TestSimultaneousScanners(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	chain := newMockChain()
	chain.extend(chain.newFromTip()) // Genesis block

	// The manglers that generate a block with a match.
	confirmManglers := []blockMangler{
		confirmScript(testPkScript),
		cfilterData(testPkScript),
	}

	// The generated test chain is:
	// - 5 blocks that miss the cfilter match
	// - 5 blocks with a cfilter match
	// - block which confirms the test pkscript
	// - 5 blocks with a cfilter match
	chain.genBlocks(5)
	chain.genBlocks(5, cfilterData(testPkScript))
	b := chain.newFromTip(confirmManglers...)
	chain.extend(b)
	chain.genBlocks(5, cfilterData(testPkScript))

	// Create and run the scanners.
	tw := NewTipWatcher(chain)
	hist := NewHistorical(chain)
	go func() {
		tw.Run(ctx)
	}()
	go func() {
		hist.Run(ctx)
	}()

	// Give it enough time for the TipWatcher to process the chain.
	time.Sleep(10 * time.Millisecond)

	// Attempt a search in both scanners in a consistent way. We first
	// watch for the desired target in the tip watcher and use the returned
	// starting watch height as the end of the historical search.
	histFoundChan := make(chan Event)
	tipFoundChan := make(chan Event)
	swhChan := make(chan int32)

	tw.Find(
		ConfirmedScript(0, testPkScript),
		WithFoundChan(tipFoundChan),
		WithStartWatchHeightChan(swhChan),
	)

	endHeight := assertStartWatchHeightSignalled(t, swhChan)

	// Generate a new block with the target script to simulate the tip
	// changing between the call to TipWatcher.Find() and
	// Historical.Find(). This could lead to multiple matches if the usage
	// of the two scanners is inconsistent.
	tip := chain.newFromTip(confirmManglers...)
	chain.extend(tip)
	chain.signalNewTip()

	// The TipWatcher should trigger a match.
	assertFoundChanRcvHeight(t, tipFoundChan, int32(tip.block.Header.Height))

	// Run the historical scanner.
	hist.Find(
		ConfirmedScript(0, testPkScript),
		WithFoundChan(histFoundChan),
		WithEndHeight(endHeight),
	)

	// Generate a new tip confirming the test script.
	tip = chain.newFromTip(confirmManglers...)
	chain.extend(tip)
	chain.signalNewTip()

	// We expect to find one (and only one) signal in both chans, each
	// pointing to their respective triggered event.
	assertFoundChanRcvHeight(t, histFoundChan, int32(b.block.Header.Height))
	assertFoundChanRcvHeight(t, tipFoundChan, int32(tip.block.Header.Height))
	assertFoundChanEmpty(t, histFoundChan)
	assertFoundChanEmpty(t, tipFoundChan)
}
