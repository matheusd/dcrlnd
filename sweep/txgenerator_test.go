package sweep

import (
	"testing"

	"github.com/decred/dcrd/wire"
	"github.com/decred/dcrlnd/input"
)

var (
	witnessTypes = []input.WitnessType{
		input.CommitmentTimeLock,
		input.HtlcAcceptedSuccessSecondLevel,
		input.HtlcOfferedRemoteTimeout,
		input.PublicKeyHash,
	}
	expectedSize    = int64(926)
	expectedSummary = "1 CommitmentTimeLock, 1 " +
		"HtlcAcceptedSuccessSecondLevel, 1 HtlcOfferedRemoteTimeout, " +
		"1 PublicKeyHash"
)

// TestWeightEstimate tests that the estimated weight and number of CSVs/CLTVs
// used is correct for a transaction that uses inputs with the witness types
// defined in witnessTypes.
func TestWeightEstimate(t *testing.T) {
	t.Parallel()

	var inputs []input.Input
	for _, witnessType := range witnessTypes {
		inputs = append(inputs, input.NewBaseInput(
			&wire.OutPoint{}, witnessType,
			&input.SignDescriptor{}, 0,
		))
	}

	_, size := getSizeEstimate(inputs)
	if size != expectedSize {
		t.Fatalf("unexpected size. expected %d but got %d.",
			expectedSize, size)
	}
	summary := inputTypeSummary(inputs)
	if summary != expectedSummary {
		t.Fatalf("unexpected summary. expected %s but got %s.",
			expectedSummary, summary)
	}
}
