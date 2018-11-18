package sweep

import (
	"fmt"

	"github.com/decred/dcrlnd/lnwallet"
)

// FeePreference allows callers to express their time value for inclusion of a
// transaction into a block via either a confirmation target, or a fee rate.
type FeePreference struct {
	// ConfTarget if non-zero, signals a fee preference expressed in the
	// number of desired blocks between first broadcast, and confirmation.
	ConfTarget uint32

	// FeeRate if non-zero, signals a fee pre fence expressed in the fee
	// rate expressed in atom/KB for a particular transaction.
	FeeRate lnwallet.AtomPerKByte
}

// DetermineFeePerKw will determine the fee in atom/KB that should be paid
// given an estimator, a confirmation target, and a manual value for sat/byte.
// A value is chosen based on the two free parameters as one, or both of them
// can be zero.
func DetermineFeePerKB(feeEstimator lnwallet.FeeEstimator,
	feePref FeePreference) (lnwallet.AtomPerKByte, error) {

	switch {
	// If the target number of confirmations is set, then we'll use that to
	// consult our fee estimator for an adequate fee.
	case feePref.ConfTarget != 0:
		feePerKw, err := feeEstimator.EstimateFeePerKB(
			feePref.ConfTarget,
		)
		if err != nil {
			return 0, fmt.Errorf("unable to query fee "+
				"estimator: %v", err)
		}

		return feePerKw, nil

	// If a manual sat/byte fee rate is set, then we'll use that directly.
	// We'll need to convert it to atom/KB as this is what we use
	// internally.
	case feePref.FeeRate != 0:
		feePerKB := feePref.FeeRate
		if feePerKB < lnwallet.FeePerKBFloor {
			log.Infof("Manual fee rate input of %d atom/KB is "+
				"too low, using %d atom/KB instead", feePerKB,
				lnwallet.FeePerKBFloor)

			feePerKB = lnwallet.FeePerKBFloor
		}

		return feePerKB, nil

	// Otherwise, we'll attempt a relaxed confirmation target for the
	// transaction
	default:
		feePerKB, err := feeEstimator.EstimateFeePerKB(6)
		if err != nil {
			return 0, fmt.Errorf("unable to query fee estimator: "+
				"%v", err)
		}

		return feePerKB, nil
	}
}
