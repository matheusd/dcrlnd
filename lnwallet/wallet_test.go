package lnwallet

import (
	"testing"

	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrlnd/input"
)

// fundingFee is a helper method that returns the fee estimate used for a tx
// with the given number of inputs and the optional change output. This matches
// the estimate done by the wallet.
func fundingFee(feeRate AtomPerKByte, numInput int, change bool) dcrutil.Amount {
	var sizeEstimate input.TxSizeEstimator

	// All inputs.
	for i := 0; i < numInput; i++ {
		sizeEstimate.AddP2PKHInput()
	}

	// The multisig funding output.
	sizeEstimate.AddP2SHOutput()

	// Optionally count a change output.
	if change {
		sizeEstimate.AddP2PKHOutput()
	}

	totalSize := sizeEstimate.Size()
	return feeRate.FeeForSize(totalSize)
}

// TestCoinSelect tests that we pick coins adding up to the expected amount
// when creating a funding transaction, and that the calculated change is the
// expected amount.
//
// NOTE: coinSelect will always attempt to add a change output, so we must
// account for this in the tests.
func TestCoinSelect(t *testing.T) {
	t.Parallel()

	const feeRate = AtomPerKByte(100)
	const dust = dcrutil.Amount(100)

	type testCase struct {
		name        string
		outputValue dcrutil.Amount
		coins       []*Utxo

		expectedInput  []dcrutil.Amount
		expectedChange dcrutil.Amount
		expectErr      bool
	}

	testCases := []testCase{
		{
			// We have 1.0 BTC available, and wants to send 0.5.
			// This will obviously lead to a change output of
			// almost 0.5 BTC.
			name: "big change",
			coins: []*Utxo{
				{
					AddressType: PubKeyHash,
					Value:       1 * dcrutil.AtomsPerCoin,
				},
			},
			outputValue: 0.5 * dcrutil.AtomsPerCoin,

			// The one and only input will be selected.
			expectedInput: []dcrutil.Amount{
				1 * dcrutil.AtomsPerCoin,
			},
			// Change will be what's left minus the fee.
			expectedChange: 0.5*dcrutil.AtomsPerCoin - fundingFee(feeRate, 1, true),
		},
		{
			// We have 1 BTC available, and we want to send 1 BTC.
			// This should lead to an error, as we don't have
			// enough funds to pay the fee.
			name: "nothing left for fees",
			coins: []*Utxo{
				{
					AddressType: PubKeyHash,
					Value:       1 * dcrutil.AtomsPerCoin,
				},
			},
			outputValue: 1 * dcrutil.AtomsPerCoin,
			expectErr:   true,
		},
		{
			// We have a 1 BTC input, and want to create an output
			// as big as possible, such that the remaining change
			// will be dust.
			name: "dust change",
			coins: []*Utxo{
				{
					AddressType: PubKeyHash,
					Value:       1 * dcrutil.AtomsPerCoin,
				},
			},
			// We tune the output value by subtracting the expected
			// fee and a small dust amount.
			outputValue: 1*dcrutil.AtomsPerCoin - fundingFee(feeRate, 1, true) - dust,

			expectedInput: []dcrutil.Amount{
				1 * dcrutil.AtomsPerCoin,
			},

			// Change will the dust.
			expectedChange: dust,
		},
		{
			// We have a 1 BTC input, and want to create an output
			// as big as possible, such that there is nothing left
			// for change.
			name: "no change",
			coins: []*Utxo{
				{
					AddressType: PubKeyHash,
					Value:       1 * dcrutil.AtomsPerCoin,
				},
			},
			// We tune the output value to be the maximum amount
			// possible, leaving just enough for fees.
			outputValue: 1*dcrutil.AtomsPerCoin - fundingFee(feeRate, 1, true),

			expectedInput: []dcrutil.Amount{
				1 * dcrutil.AtomsPerCoin,
			},
			// We have just enough left to pay the fee, so there is
			// nothing left for change.
			// TODO(halseth): currently coinselect estimates fees
			// assuming a change output.
			expectedChange: 0,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			selected, changeAmt, err := coinSelect(
				feeRate, test.outputValue, test.coins,
			)
			if !test.expectErr && err != nil {
				t.Fatalf(err.Error())
			}

			if test.expectErr && err == nil {
				t.Fatalf("expected error")
			}

			// If we got an expected error, there is nothing more to test.
			if test.expectErr {
				return
			}

			// Check that the selected inputs match what we expect.
			if len(selected) != len(test.expectedInput) {
				t.Fatalf("expected %v inputs, got %v",
					len(test.expectedInput), len(selected))
			}

			for i, coin := range selected {
				if coin.Value != test.expectedInput[i] {
					t.Fatalf("expected input %v to have value %v, "+
						"had %v", i, test.expectedInput[i],
						coin.Value)
				}
			}

			// Assert we got the expected change amount.
			if changeAmt != test.expectedChange {
				t.Fatalf("expected %v change amt, got %v",
					test.expectedChange, changeAmt)
			}
		})
	}
}
