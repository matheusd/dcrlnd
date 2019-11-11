package lnrpc

import (
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrlnd/lnwire"
)

// CalculateFeeLimit returns the fee limit in milliatoms. If a percentage based
// fee limit has been requested, we'll factor in the ratio provided with the
// amount of the payment.
func CalculateFeeLimit(feeLimit *FeeLimit,
	amount lnwire.MilliAtom) lnwire.MilliAtom {

	switch feeLimit.GetLimit().(type) {

	case *FeeLimit_Fixed:
		return lnwire.NewMAtomsFromAtoms(
			dcrutil.Amount(feeLimit.GetFixed()),
		)

	case *FeeLimit_Percent:
		return amount * lnwire.MilliAtom(feeLimit.GetPercent()) / 100

	default:
		// If a fee limit was not specified, we'll use the payment's
		// amount as an upper bound in order to avoid payment attempts
		// from incurring fees higher than the payment amount itself.
		return amount
	}
}
