package lnwallet

import (
	"fmt"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrutil/v3"
	"github.com/decred/dcrd/txscript/v3"
	"github.com/decred/dcrlnd/chainntnfs"
	"github.com/decred/dcrlnd/channeldb"
	"github.com/decred/dcrlnd/input"
	"github.com/decred/dcrlnd/lnwire"
	"github.com/decred/slog"
	"github.com/matheusd/dcr_adaptor_sigs"
)

func pch(ch chainhash.Hash) *chainhash.Hash {
	return &ch
}

// \/\/\/\/\/\/\/\/\/ channel_test.go \/\/\/\/\/\/\/\/\/\/\/

// createPTLC is a utility function for generating an PTLC with a given pubkey
// and a given amount.
func createPTLC(id int, amount lnwire.MilliAtom) (*lnwire.UpdateAddHTLC, *dcr_adaptor_sigs.SecretScalar) {
	var secret dcr_adaptor_sigs.SecretScalar
	secret[0] = 0x0f // So we avoid 0 keys.
	secret[31] = byte(id)
	secret[30] = byte(id >> 8)

	return &lnwire.UpdateAddHTLC{
		ID:           uint64(id),
		PaymentPoint: secret.PublicPoint(),
		Amount:       amount,
		Expiry:       uint32(5),
	}, &secret
}

// TestBasicPTLCFlows tests that adding, settling and failing PTLCs work on a
// channel and that PTLC outputs of the generated commitment are redeemable.
func TestBasicPTLCFlows(t *testing.T) {
	t.Parallel()

	// TODO: remove
	spew.Sdump(nil)
	fmt.Println("")

	// We'll kick off the test by creating our channels which both are
	// loaded with 5 DCR each.
	chanType := channeldb.SingleFunderTweaklessBit
	aliceChannel, bobChannel, cleanUp, err := CreateTestChannels(
		chanType,
	)
	if err != nil {
		t.Fatalf("unable to create test channels: %v", err)
	}
	defer cleanUp()

	fmt.Println("LLLLL Created test channels")

	// First, we will add and confirm a PTLC of 0.1 DCR.
	ptlcAmt := lnwire.NewMAtomsFromAtoms(0.1 * dcrutil.AtomsPerCoin)
	ptlc, secret := createPTLC(0, ptlcAmt)
	if _, err := aliceChannel.AddHTLC(ptlc, nil); err != nil {
		t.Fatalf("unable to add htlc: %v", err)
	}
	if _, err := bobChannel.ReceiveHTLC(ptlc); err != nil {
		t.Fatalf("unable to recv htlc: %v", err)
	}
	if err := ForceStateTransition(aliceChannel, bobChannel); err != nil {
		t.Fatalf("unable to complete state update: %v", err)
	}

	// Now we generate a unilateral force-close (i.e. the signed commitment
	// tx of the last channel state) to check some scenarios. aliceClose is
	// the force-close from the PoV of Alice, bobClose is the force-close
	// from the PoV of Bob.
	fmt.Println("force closing alice")
	aliceClose, err := aliceChannel.ForceClose()
	if err != nil {
		t.Fatalf("unable to generate bob's force close: %v", err)
	}
	fmt.Println("alice force closed")

	bobClose, err := bobChannel.ForceClose()
	if err != nil {
		t.Fatalf("unable to generate bob's force close: %v", err)
	}

	// If Bob force-closes, then it should be possible to complete Bob's
	// success tx once he has learned of the secret scalar needed to
	// complete the adaptor sig.  We do just so below, complete the success
	// tx input script and verify it's reedemable on-chain. We can pass in
	// `nil` for the adaptorSig argument because the channel should have
	// filled in the serialized adaptor sig in the existing sig script.
	successTx := bobClose.HtlcResolutions.IncomingHTLCs[0].SignedSuccessTx
	newSigScript, err := input.ReplaceReceiverPtlcSpendRedeemPreimage(
		successTx.TxIn[0].SignatureScript, nil, secret,
	)
	if err != nil {
		t.Fatalf("Failed creating newSigScript: %v", err)
	}

	bobClosePTLCOut := bobClose.CloseTx.TxOut[successTx.TxIn[0].PreviousOutPoint.Index]
	successTx.TxIn[0].SignatureScript = newSigScript

	// Test that the script is in fact correct.
	vm, err := txscript.NewEngine(
		bobClosePTLCOut.PkScript, successTx, 0,
		input.ScriptVerifyFlags, bobClosePTLCOut.Version, nil,
	)
	if err != nil {
		t.Fatalf("cannot create script engine: %s", err)
	}
	if err = vm.Execute(); err != nil {
		t.Fatalf("cannot validate success transaction: %s", err)
	}

	fmt.Printf("GGGGGGGG remoteCommitment for unilateralCS: current %x next %x\n",
		bobChannel.channelState.RemoteCurrentRevocation.X(),
		bobChannel.channelState.RemoteNextRevocation.X())

	// If Alice force-closes, it should be possible for Bob to sweep the
	// PTLC output via the success-no-delay tx by filling in the adaptor
	// sig with the secret scalar.
	//
	// First, generate the successNoDelay tx.
	bobUniClose, err := NewUnilateralCloseSummary(
		bobChannel.channelState, bobChannel.Signer,
		&chainntnfs.SpendDetail{
			SpendingTx:    aliceClose.CloseTx,
			SpenderTxHash: pch(aliceClose.CloseTx.TxHash()),
		},
		bobChannel.channelState.RemoteCommitment,
		bobChannel.channelState.RemoteCurrentRevocation,
	)
	if err != nil {
		t.Fatalf("unable to fetch bob unilateral close summary: %v", err)
	}
	successNoDelayTx := bobUniClose.HtlcResolutions.IncomingHTLCs[0].SignedSuccessNoDelayTx

	// Now, fill in the secret scalar in the adaptor sig, generating a
	// fully valid Schnorr sig.
	newSigScript, err = input.ReplaceSenderPtlcSpendRedeemScalar(
		successNoDelayTx.TxIn[0].SignatureScript, nil, secret,
	)
	if err != nil {
		t.Fatalf("Failed creating newSigScript: %v", err)
	}
	successNoDelayTx.TxIn[0].SignatureScript = newSigScript

	bknd := slog.NewBackend(os.Stdout)
	logg := bknd.Logger("XXXX")
	//logg.SetLevel(slog.LevelTrace)
	txscript.UseLogger(logg)

	// Fetch the original pkscript that was broadcast on Alice's force
	// close and ensure the successNoDelayTx input actually has the correct
	// signatureScript.
	aliceClosePTLCOut := aliceClose.CloseTx.TxOut[successNoDelayTx.TxIn[0].PreviousOutPoint.Index]
	vm, err = txscript.NewEngine(
		aliceClosePTLCOut.PkScript, successNoDelayTx, 0,
		input.ScriptVerifyFlags, aliceClosePTLCOut.Version, nil,
	)
	if err != nil {
		t.Fatalf("cannot create script engine: %s", err)
	}
	if err = vm.Execute(); err != nil {
		t.Fatalf("cannot validate success-no-delay transaction: %s", err)
	}

	// TODO: ensure Alice can find out the preimage from either her or
	// Bob's unlateral close.
}
