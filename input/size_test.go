package input_test

import (
	"testing"

	"github.com/decred/dcrd/chaincfg/v2"
	"github.com/decred/dcrd/dcrec"
	"github.com/decred/dcrd/dcrec/secp256k1/v3"
	"github.com/decred/dcrd/dcrutil/v3"
	"github.com/decred/dcrd/txscript/v3"
	"github.com/decred/dcrd/wire"
	"github.com/decred/dcrlnd/input"
	"github.com/decred/dcrlnd/keychain"
)

const (
	// A CSV int may be up to 5 bytes long, so test with the max uint32.
	testCSVDelay = (1 << 32) - 1

	testCLTVExpiry = 500000000

	// maxDERSignatureSize is the largest possible DER-encoded signature
	// without the trailing sighash flag.
	maxDERSignatureSize = 72
)

var (
	testPubkeyBytes = make([]byte, 33)

	testHash160  = make([]byte, 20)
	testPreimage = make([]byte, 32)

	testPrivkey = secp256k1.PrivKeyFromBytes(make([]byte, 32))
	testPubkey  = testPrivkey.PubKey()

	testTx                   = wire.NewMsgTx()
	testMaxDERSigPlusSigHash = make([]byte, maxDERSignatureSize+1)
)

func init() {
	testTx.Version = 2
}

// TestTxSizeEstimator tests that transaction size estimates are calculated
// correctly by comparing against an actual (though invalid) transaction
// matching the template.
func TestTxSizeEstimator(t *testing.T) {
	netParams := chaincfg.MainNetParams()

	// Static test data.
	var nullData [73]byte

	p2pkhAddr, err := dcrutil.NewAddressPubKeyHash(
		nullData[:20], netParams, dcrec.STEcdsaSecp256k1)
	if err != nil {
		t.Fatalf("Failed to generate address: %v", err)
	}
	p2pkhPkScript, err := txscript.PayToAddrScript(p2pkhAddr)
	if err != nil {
		t.Fatalf("Failed to generate scriptPubKey: %v", err)
	}

	signature := nullData[:73]
	compressedPubKey := nullData[:33]
	p2pkhSigScript, err := txscript.NewScriptBuilder().AddData(signature).
		AddData(compressedPubKey).Script()
	if err != nil {
		t.Fatalf("Failed to generate p2pkhSigScript: %v", err)
	}

	p2shAddr, err := dcrutil.NewAddressScriptHashFromHash(nullData[:20], netParams)
	if err != nil {
		t.Fatalf("Failed to generate address: %v", err)
	}
	p2shPkScript, err := txscript.PayToAddrScript(p2shAddr)
	if err != nil {
		t.Fatalf("Failed to generate scriptPubKey: %v", err)
	}
	p2shRedeemScript := nullData[:71] // 2-of-2 multisig
	p2shSigScript, err := txscript.NewScriptBuilder().AddData(signature).
		AddData(signature).AddData(p2shRedeemScript).Script()
	if err != nil {
		t.Fatalf("Failed to generate ps2shSigScript: %v", err)
	}

	testCases := []struct {
		numP2PKHInputs  int
		numP2SHInputs   int
		numP2PKHOutputs int
		numP2SHOutputs  int
	}{
		{
			numP2PKHInputs:  1,
			numP2PKHOutputs: 2,
		},
		{
			numP2PKHInputs: 1,
			numP2SHOutputs: 1,
		},
		{
			numP2SHInputs:   1,
			numP2PKHOutputs: 1,
		},
		{
			numP2SHInputs:  1,
			numP2SHOutputs: 1,
		},
		{
			numP2SHInputs:   1,
			numP2PKHOutputs: 1,
			numP2SHOutputs:  2,
		},
		{
			numP2SHInputs:   253,
			numP2PKHOutputs: 1,
			numP2SHOutputs:  1,
		},
		{
			numP2SHInputs:   1,
			numP2PKHOutputs: 253,
			numP2SHOutputs:  1,
		},
		{
			numP2SHInputs:   1,
			numP2PKHOutputs: 1,
			numP2SHOutputs:  253,
		},
	}

	for i, test := range testCases {
		var sizeEstimate input.TxSizeEstimator
		tx := wire.NewMsgTx()

		// Inputs.

		for j := 0; j < test.numP2PKHInputs; j++ {
			sizeEstimate.AddP2PKHInput()
			tx.AddTxIn(&wire.TxIn{SignatureScript: p2pkhSigScript})
		}
		for j := 0; j < test.numP2SHInputs; j++ {
			sizeEstimate.AddCustomInput(int64(len(p2shSigScript)))
			tx.AddTxIn(&wire.TxIn{SignatureScript: p2shSigScript})
		}

		// Outputs.

		for j := 0; j < test.numP2PKHOutputs; j++ {
			sizeEstimate.AddP2PKHOutput()
			tx.AddTxOut(&wire.TxOut{PkScript: p2pkhPkScript})
		}
		for j := 0; j < test.numP2SHOutputs; j++ {
			sizeEstimate.AddP2SHOutput()
			tx.AddTxOut(&wire.TxOut{PkScript: p2shPkScript})
		}

		expectedSize := int64(tx.SerializeSize())
		actualSize := sizeEstimate.Size()
		if actualSize != expectedSize {
			t.Errorf("Case %d: Got wrong size: expected %d, got %d",
				i, expectedSize, actualSize)
		}
	}
}

type maxDERSignature struct{}

func (s *maxDERSignature) Serialize() []byte {
	// Always return worst-case signature length, excluding the one byte
	// sighash flag.
	return make([]byte, maxDERSignatureSize)
}

func (s *maxDERSignature) Verify(_ []byte, _ *secp256k1.PublicKey) bool {
	return true
}

// dummySigner is a fake signer used for size (upper bound) calculations.
type dummySigner struct {
	input.Signer
}

// SignOutputRaw generates a signature for the passed transaction according to
// the data within the passed SignDescriptor.
func (s *dummySigner) SignOutputRaw(tx *wire.MsgTx,
	signDesc *input.SignDescriptor) (input.Signature, error) {

	return &maxDERSignature{}, nil
}

type witnessSizeTest struct {
	name       string
	expSize    int64
	genWitness func(t *testing.T) input.TxWitness
}

var witnessSizeTests = []witnessSizeTest{
	{
		name:    "funding",
		expSize: input.FundingOutputSigScriptSize,
		genWitness: func(t *testing.T) input.TxWitness {
			witnessScript, _, err := input.GenFundingPkScript(
				testPubkeyBytes, testPubkeyBytes, 1,
			)
			if err != nil {
				t.Fatal(err)
			}

			return input.SpendMultiSig(
				witnessScript,
				testPubkeyBytes, testMaxDERSigPlusSigHash,
				testPubkeyBytes, testMaxDERSigPlusSigHash,
			)
		},
	},
	{
		name:    "to local timeout",
		expSize: input.ToLocalTimeoutSigScriptSize,
		genWitness: func(t *testing.T) input.TxWitness {
			witnessScript, err := input.CommitScriptToSelf(
				testCSVDelay, testPubkey, testPubkey,
			)
			if err != nil {
				t.Fatal(err)
			}

			signDesc := &input.SignDescriptor{
				WitnessScript: witnessScript,
			}

			witness, err := input.CommitSpendTimeout(
				&dummySigner{}, signDesc, testTx,
			)
			if err != nil {
				t.Fatal(err)
			}

			return witness
		},
	},
	{
		name:    "to local revoke",
		expSize: input.ToLocalPenaltySigScriptSize,
		genWitness: func(t *testing.T) input.TxWitness {
			witnessScript, err := input.CommitScriptToSelf(
				testCSVDelay, testPubkey, testPubkey,
			)
			if err != nil {
				t.Fatal(err)
			}

			signDesc := &input.SignDescriptor{
				WitnessScript: witnessScript,
			}

			witness, err := input.CommitSpendRevoke(
				&dummySigner{}, signDesc, testTx,
			)
			if err != nil {
				t.Fatal(err)
			}

			return witness
		},
	},
	{
		name:    "to remote confirmed",
		expSize: input.ToRemoteConfirmedWitnessSize,
		genWitness: func(t *testing.T) input.TxWitness {
			witScript, err := input.CommitScriptToRemoteConfirmed(
				testPubkey,
			)
			if err != nil {
				t.Fatal(err)
			}

			signDesc := &input.SignDescriptor{
				WitnessScript: witScript,
				KeyDesc: keychain.KeyDescriptor{
					PubKey: testPubkey,
				},
			}

			witness, err := input.CommitSpendToRemoteConfirmed(
				&dummySigner{}, signDesc, testTx,
			)
			if err != nil {
				t.Fatal(err)
			}

			return witness
		},
	},
	{
		name:    "anchor",
		expSize: input.AnchorSigScriptSize,
		genWitness: func(t *testing.T) input.TxWitness {
			witScript, err := input.CommitScriptAnchor(
				testPubkey,
			)
			if err != nil {
				t.Fatal(err)
			}

			signDesc := &input.SignDescriptor{
				WitnessScript: witScript,
				KeyDesc: keychain.KeyDescriptor{
					PubKey: testPubkey,
				},
			}

			witness, err := input.CommitSpendAnchor(
				&dummySigner{}, signDesc, testTx,
			)
			if err != nil {
				t.Fatal(err)
			}

			return witness
		},
	},
	{
		name:    "anchor anyone",
		expSize: input.AnchorAnyoneSigScriptSize,
		genWitness: func(t *testing.T) input.TxWitness {
			witScript, err := input.CommitScriptAnchor(
				testPubkey,
			)
			if err != nil {
				t.Fatal(err)
			}

			witness, _ := input.CommitSpendAnchorAnyone(witScript)

			return witness
		},
	},
	{
		name:    "offered htlc revoke",
		expSize: input.OfferedHtlcPenaltySigScriptSize - 3,
		genWitness: func(t *testing.T) input.TxWitness {
			witScript, err := input.SenderHTLCScript(
				testPubkey, testPubkey, testPubkey,
				testHash160, false,
			)
			if err != nil {
				t.Fatal(err)
			}

			signDesc := &input.SignDescriptor{
				WitnessScript: witScript,
				KeyDesc: keychain.KeyDescriptor{
					PubKey: testPubkey,
				},
				DoubleTweak: testPrivkey,
			}

			witness, err := input.SenderHtlcSpendRevoke(
				&dummySigner{}, signDesc, testTx,
			)
			if err != nil {
				t.Fatal(err)
			}

			return witness
		},
	},
	{
		name:    "offered htlc revoke confirmed",
		expSize: input.OfferedHtlcPenaltySigScriptSize,
		genWitness: func(t *testing.T) input.TxWitness {
			hash := make([]byte, 20)

			witScript, err := input.SenderHTLCScript(
				testPubkey, testPubkey, testPubkey,
				hash, true,
			)
			if err != nil {
				t.Fatal(err)
			}

			signDesc := &input.SignDescriptor{
				WitnessScript: witScript,
				KeyDesc: keychain.KeyDescriptor{
					PubKey: testPubkey,
				},
				DoubleTweak: testPrivkey,
			}

			witness, err := input.SenderHtlcSpendRevoke(
				&dummySigner{}, signDesc, testTx,
			)
			if err != nil {
				t.Fatal(err)
			}

			return witness
		},
	},
	{
		name:    "offered htlc timeout",
		expSize: input.OfferedHtlcTimeoutSigScriptSize - 3,
		genWitness: func(t *testing.T) input.TxWitness {
			witScript, err := input.SenderHTLCScript(
				testPubkey, testPubkey, testPubkey,
				testHash160, false,
			)
			if err != nil {
				t.Fatal(err)
			}

			signDesc := &input.SignDescriptor{
				WitnessScript: witScript,
			}

			witness, err := input.SenderHtlcSpendTimeout(
				&maxDERSignature{}, txscript.SigHashAll,
				&dummySigner{}, signDesc, testTx,
			)
			if err != nil {
				t.Fatal(err)
			}

			return witness
		},
	},
	{
		name:    "offered htlc timeout confirmed",
		expSize: input.OfferedHtlcTimeoutSigScriptSize,
		genWitness: func(t *testing.T) input.TxWitness {
			witScript, err := input.SenderHTLCScript(
				testPubkey, testPubkey, testPubkey,
				testHash160, true,
			)
			if err != nil {
				t.Fatal(err)
			}

			signDesc := &input.SignDescriptor{
				WitnessScript: witScript,
			}

			witness, err := input.SenderHtlcSpendTimeout(
				&maxDERSignature{}, txscript.SigHashAll,
				&dummySigner{}, signDesc, testTx,
			)
			if err != nil {
				t.Fatal(err)
			}

			return witness
		},
	},
	{
		name:    "offered htlc success",
		expSize: input.OfferedHtlcSuccessSigScriptSize - 3,
		genWitness: func(t *testing.T) input.TxWitness {
			witScript, err := input.SenderHTLCScript(
				testPubkey, testPubkey, testPubkey,
				testHash160, false,
			)
			if err != nil {
				t.Fatal(err)
			}

			signDesc := &input.SignDescriptor{
				WitnessScript: witScript,
			}

			witness, err := input.SenderHtlcSpendRedeem(
				&dummySigner{}, signDesc, testTx, testPreimage,
			)
			if err != nil {
				t.Fatal(err)
			}
			for _, w := range witness {
				t.Logf("%d %x", len(w), w)
			}

			return witness
		},
	},
	{
		name:    "offered htlc success confirmed",
		expSize: input.OfferedHtlcSuccessSigScriptSize,
		genWitness: func(t *testing.T) input.TxWitness {
			witScript, err := input.SenderHTLCScript(
				testPubkey, testPubkey, testPubkey,
				testHash160, true,
			)
			if err != nil {
				t.Fatal(err)
			}

			signDesc := &input.SignDescriptor{
				WitnessScript: witScript,
			}

			witness, err := input.SenderHtlcSpendRedeem(
				&dummySigner{}, signDesc, testTx, testPreimage,
			)
			if err != nil {
				t.Fatal(err)
			}

			return witness
		},
	},
	{
		name:    "accepted htlc revoke",
		expSize: input.AcceptedHtlcPenaltySigScriptSize - 3,
		genWitness: func(t *testing.T) input.TxWitness {
			witScript, err := input.ReceiverHTLCScript(
				testCLTVExpiry, testPubkey, testPubkey,
				testPubkey, testHash160, false,
			)
			if err != nil {
				t.Fatal(err)
			}

			signDesc := &input.SignDescriptor{
				WitnessScript: witScript,
				KeyDesc: keychain.KeyDescriptor{
					PubKey: testPubkey,
				},
				DoubleTweak: testPrivkey,
			}

			witness, err := input.ReceiverHtlcSpendRevoke(
				&dummySigner{}, signDesc, testTx,
			)
			if err != nil {
				t.Fatal(err)
			}

			return witness
		},
	},
	{
		name:    "accepted htlc revoke confirmed",
		expSize: input.AcceptedHtlcPenaltySigScriptSize,
		genWitness: func(t *testing.T) input.TxWitness {
			witScript, err := input.ReceiverHTLCScript(
				testCLTVExpiry, testPubkey, testPubkey,
				testPubkey, testHash160, true,
			)
			if err != nil {
				t.Fatal(err)
			}

			signDesc := &input.SignDescriptor{
				WitnessScript: witScript,
				KeyDesc: keychain.KeyDescriptor{
					PubKey: testPubkey,
				},
				DoubleTweak: testPrivkey,
			}

			witness, err := input.ReceiverHtlcSpendRevoke(
				&dummySigner{}, signDesc, testTx,
			)
			if err != nil {
				t.Fatal(err)
			}

			return witness
		},
	},
	{
		name:    "accepted htlc timeout",
		expSize: input.AcceptedHtlcTimeoutSigScriptSize - 3,
		genWitness: func(t *testing.T) input.TxWitness {

			witScript, err := input.ReceiverHTLCScript(
				testCLTVExpiry, testPubkey, testPubkey,
				testPubkey, testHash160, false,
			)
			if err != nil {
				t.Fatal(err)
			}

			signDesc := &input.SignDescriptor{
				WitnessScript: witScript,
			}

			witness, err := input.ReceiverHtlcSpendTimeout(
				&dummySigner{}, signDesc, testTx,
				testCLTVExpiry,
			)
			if err != nil {
				t.Fatal(err)
			}

			return witness
		},
	},
	{
		name:    "accepted htlc timeout confirmed",
		expSize: input.AcceptedHtlcTimeoutSigScriptSize,
		genWitness: func(t *testing.T) input.TxWitness {
			witScript, err := input.ReceiverHTLCScript(
				testCLTVExpiry, testPubkey, testPubkey,
				testPubkey, testHash160, true,
			)
			if err != nil {
				t.Fatal(err)
			}

			signDesc := &input.SignDescriptor{
				WitnessScript: witScript,
			}

			witness, err := input.ReceiverHtlcSpendTimeout(
				&dummySigner{}, signDesc, testTx,
				testCLTVExpiry,
			)
			if err != nil {
				t.Fatal(err)
			}

			return witness
		},
	},
	{
		name:    "accepted htlc success",
		expSize: input.AcceptedHtlcSuccessSigScriptSize - 3,
		genWitness: func(t *testing.T) input.TxWitness {
			witScript, err := input.ReceiverHTLCScript(
				testCLTVExpiry, testPubkey, testPubkey,
				testPubkey, testHash160, false,
			)
			if err != nil {
				t.Fatal(err)
			}

			signDesc := &input.SignDescriptor{
				WitnessScript: witScript,
				KeyDesc: keychain.KeyDescriptor{
					PubKey: testPubkey,
				},
			}

			witness, err := input.ReceiverHtlcSpendRedeem(
				&maxDERSignature{}, txscript.SigHashAll,
				testPreimage, &dummySigner{}, signDesc, testTx,
			)
			if err != nil {
				t.Fatal(err)
			}

			return witness
		},
	},
	{
		name:    "accepted htlc success confirmed",
		expSize: input.AcceptedHtlcSuccessSigScriptSize,
		genWitness: func(t *testing.T) input.TxWitness {
			witScript, err := input.ReceiverHTLCScript(
				testCLTVExpiry, testPubkey, testPubkey,
				testPubkey, testHash160, true,
			)
			if err != nil {
				t.Fatal(err)
			}

			signDesc := &input.SignDescriptor{
				WitnessScript: witScript,
				KeyDesc: keychain.KeyDescriptor{
					PubKey: testPubkey,
				},
			}

			witness, err := input.ReceiverHtlcSpendRedeem(
				&maxDERSignature{}, txscript.SigHashAll,
				testPreimage, &dummySigner{}, signDesc, testTx,
			)
			if err != nil {
				t.Fatal(err)
			}

			return witness
		},
	},
}

// TestWitnessSizes asserts the correctness of our magic witness constants.
// Witnesses involving signatures will have maxDERSignatures injected so that we
// can determine upper bounds for the witness sizes. These constants are
// predominately used for fee estimation, so we want to be certain that we
// aren't under estimating or our transactions could get stuck.
func TestWitnessSizes(t *testing.T) {
	for _, test := range witnessSizeTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			witness := test.genWitness(t)
			sigScript, err := input.WitnessStackToSigScript(witness)
			if err != nil {
				t.Fatalf("unable to convert witness to sigScript: %v", err)
			}
			size := int64(len(sigScript))
			if size != test.expSize {
				t.Fatalf("size mismatch, want: %v, got: %v",
					test.expSize, size)
			}
		})
	}
}
