package lnwallet_test

import (
	"testing"

	"github.com/decred/dcrd/chaincfg/v2"
	"github.com/decred/dcrd/dcrec"
	"github.com/decred/dcrd/dcrutil/v2"
	"github.com/decred/dcrd/txscript/v2"
	"github.com/decred/dcrd/wire"
	"github.com/decred/dcrlnd/input"
)

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
