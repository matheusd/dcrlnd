package sweep

import (
	"testing"

	"github.com/decred/dcrd/dcrutil/v2"
	"github.com/decred/dcrlnd/input"
)

// TestTxInputSet tests adding various sized inputs to the set.
func TestTxInputSet(t *testing.T) {
	const (
		feeRate   = 5e4
		relayFee  = 1e4
		maxInputs = 10
	)
	set := newTxInputSet(feeRate, relayFee, maxInputs)

	wantDust := dcrutil.Amount(6030)
	if set.dustLimit != 6030 {
		t.Fatalf("incorrect dust limit; want=%d got=%d", wantDust, set.dustLimit)
	}

	// Create a 10001 atom input. The fee to sweep this input to a P2PKH
	// output is 10850 atoms. That means that this input yields -849 atoms
	// and we expect it not to be added.
	if set.add(createP2PKHInput(10001)) {
		t.Fatal("expected add of negatively yielding input to fail")
	}

	// A 15000 atom input should be accepted into the set, because it
	// yields positively.
	if !set.add(createP2PKHInput(15001)) {
		t.Fatal("expected add of positively yielding input to succeed")
	}

	// The tx output should now be 15000-10850 = 4151 atoms. The dust limit
	// isn't reached yet.
	wantOutputValue := dcrutil.Amount(4151)
	if set.outputValue != wantOutputValue {
		t.Fatalf("unexpected output value. want=%d got=%d", wantOutputValue, set.outputValue)
	}
	if set.dustLimitReached() {
		t.Fatal("expected dust limit not yet to be reached")
	}

	// Add a 13703 atoms input. This increases the tx fee to 19150 atoms.
	// The tx output should now be 13703+15001 - 19150 = 9554 atoms.
	if !set.add(createP2PKHInput(13703)) {
		t.Fatal("expected add of positively yielding input to succeed")
	}
	wantOutputValue = 9554
	if set.outputValue != wantOutputValue {
		t.Fatalf("unexpected output value. want=%d got=%d", wantOutputValue, set.outputValue)
	}
	if !set.dustLimitReached() {
		t.Fatal("expected dust limit to be reached")
	}
}

// createP2PKHInput returns a P2PKH test input with the specified amount.
func createP2PKHInput(amt dcrutil.Amount) input.Input {
	input := createTestInput(int64(amt), input.PublicKeyHash)
	return &input
}
