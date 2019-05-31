package sweep

import (
	"fmt"
	"sort"

	"github.com/decred/dcrd/blockchain"
	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/wire"
	"github.com/decred/dcrlnd/lnwallet"
)

var (
	// DefaultMaxInputsPerTx specifies the default maximum number of inputs
	// allowed in a single sweep tx. If more need to be swept, multiple txes
	// are created and published.
	DefaultMaxInputsPerTx = 100
)

// inputSet is a set of inputs that can be used as the basis to generate a tx
// on.
type inputSet []Input

// generateInputPartitionings goes through all given inputs and constructs sets
// of inputs that can be used to generate a sensible transaction. Each set
// contains up to the configured maximum number of inputs. Negative yield
// inputs are skipped. No input sets with a total value after fees below the
// dust limit are returned.
func generateInputPartitionings(sweepableInputs []Input,
	relayFeePerKB, feePerKB lnwallet.AtomPerKByte,
	maxInputsPerTx int) ([]inputSet, error) {

	// Calculate dust limit based on the P2PKH output script of the sweep
	// txes.
	dustLimit := lnwallet.DustThresholdForRelayFee(relayFeePerKB)

	// Sort input by yield. We will start constructing input sets starting
	// with the highest yield inputs. This is to prevent the construction
	// of a set with an output below the dust limit, causing the sweep
	// process to stop, while there are still higher value inputs
	// available. It also allows us to stop evaluating more inputs when the
	// first input in this ordering is encountered with a negative yield.
	//
	// Yield is calculated as the difference between value and added fee
	// for this input. The fee calculation excludes fee components that are
	// common to all inputs, as those wouldn't influence the order. The
	// single component that is differentiating is witness size.
	//
	// For witness size, the upper limit is taken. The actual size depends
	// on the signature length, which is not known yet at this point.
	yields := make(map[wire.OutPoint]int64)
	for _, input := range sweepableInputs {
		size, err := getInputSigScriptSizeUpperBound(input)
		if err != nil {
			return nil, fmt.Errorf(
				"failed adding input size: %v", err)
		}

		yields[*input.OutPoint()] = input.SignDesc().Output.Value -
			int64(feePerKB.FeeForSize(size))
	}

	sort.Slice(sweepableInputs, func(i, j int) bool {
		return yields[*sweepableInputs[i].OutPoint()] >
			yields[*sweepableInputs[j].OutPoint()]
	})

	// Select blocks of inputs up to the configured maximum number.
	var sets []inputSet
	for len(sweepableInputs) > 0 {
		// Get the maximum number of inputs from sweepableInputs that
		// we can use to create a positive yielding set from.
		count, outputValue := getPositiveYieldInputs(
			sweepableInputs, maxInputsPerTx, feePerKB,
		)

		// If there are no positive yield inputs left, we can stop
		// here.
		if count == 0 {
			return sets, nil
		}

		// If the output value of this block of inputs does not reach
		// the dust limit, stop sweeping. Because of the sorting,
		// continuing with the remaining inputs will only lead to sets
		// with a even lower output value.
		if outputValue < dustLimit {
			log.Debugf("Set value %v below dust limit of %v",
				outputValue, dustLimit)
			return sets, nil
		}

		log.Infof("Candidate sweep set of size=%v, has yield=%v",
			count, outputValue)

		sets = append(sets, sweepableInputs[:count])
		sweepableInputs = sweepableInputs[count:]
	}

	return sets, nil
}

// getPositiveYieldInputs returns the maximum of a number n for which holds
// that the inputs [0,n) of sweepableInputs have a positive yield.
// Additionally, the total values of these inputs minus the fee is returned.
//
// TODO(roasbeef): Consider including some negative yield inputs too to clean
// up the utxo set even if it costs us some fees up front.  In the spirit of
// minimizing any negative externalities we cause for the Bitcoin system as a
// whole.
func getPositiveYieldInputs(sweepableInputs []Input, maxInputs int,
	feePerKB lnwallet.AtomPerKByte) (int, dcrutil.Amount) {

	var sizeEstimate lnwallet.TxSizeEstimator

	// Add the sweep tx output to the size estimate.
	sizeEstimate.AddP2PKHOutput()

	var total, outputValue dcrutil.Amount
	for idx, input := range sweepableInputs {
		// Can ignore error, because it has already been checked when
		// calculating the yields.
		sigScriptSize, _ := getInputSigScriptSizeUpperBound(input)

		// Keep a running size estimate of the input set.
		sizeEstimate.AddCustomInput(sigScriptSize)

		newTotal := total + dcrutil.Amount(input.SignDesc().Output.Value)

		size := sizeEstimate.Size()
		fee := feePerKB.FeeForSize(size)

		// Calculate the output value if the current input would be
		// added to the set.
		newOutputValue := newTotal - fee

		// If adding this input makes the total output value of the set
		// decrease, this is a negative yield input. It shouldn't be
		// added to the set. We return the current index as the number
		// of inputs, so the current input is being excluded.
		if newOutputValue <= outputValue {
			return idx, outputValue
		}

		// Update running values.
		total = newTotal
		outputValue = newOutputValue

		// Stop if max inputs is reached.
		if idx == maxInputs-1 {
			return maxInputs, outputValue
		}
	}

	// We could add all inputs to the set, so return them all.
	return len(sweepableInputs), outputValue
}

// createSweepTx builds a signed tx spending the inputs to a the output script.
func createSweepTx(inputs []Input, outputPkScript []byte,
	currentBlockHeight uint32, feePerKB lnwallet.AtomPerKByte,
	signer lnwallet.Signer, netParams *chaincfg.Params) (*wire.MsgTx, error) {

	inputs, txSize, csvCount, cltvCount := getSizeEstimate(inputs)

	log.Infof("Creating sweep transaction for %v inputs (%v CSV, %v CLTV) "+
		"using %v atom/kB", len(inputs), csvCount, cltvCount,
		int64(feePerKB))

	txFee := feePerKB.FeeForSize(txSize)

	// Sum up the total value contained in the inputs.
	var totalSum dcrutil.Amount
	for _, o := range inputs {
		totalSum += dcrutil.Amount(o.SignDesc().Output.Value)
	}

	// Sweep as much possible, after subtracting txn fees.
	sweepAmt := int64(totalSum - txFee)

	// Create the sweep transaction that we will be building. We use
	// version 2 as it is required for CSV. The txn will sweep the amount
	// after fees to the pkscript generated above.
	sweepTx := wire.NewMsgTx()
	sweepTx.Version = 2
	sweepTx.AddTxOut(&wire.TxOut{
		PkScript: outputPkScript,
		Value:    sweepAmt,
	})

	sweepTx.LockTime = currentBlockHeight

	// Add all inputs to the sweep transaction. Ensure that for each
	// csvInput, we set the sequence number properly.
	for _, input := range inputs {
		sweepTx.AddTxIn(&wire.TxIn{
			ValueIn:          input.SignDesc().Output.Value,
			PreviousOutPoint: *input.OutPoint(),
			Sequence:         input.BlocksToMaturity(),
		})
	}

	// Before signing the transaction, check to ensure that it meets some
	// basic validity requirements.
	//
	// TODO(conner): add more control to sanity checks, allowing us to
	// delay spending "problem" outputs, e.g. possibly batching with other
	// classes if fees are too low.
	btx := dcrutil.NewTx(sweepTx)
	if err := blockchain.CheckTransactionSanity(btx.MsgTx(), netParams); err != nil {
		if ruleErr, is := err.(blockchain.RuleError); is {
			return nil, fmt.Errorf("rule error checking sweepTx sanity: %s %v",
				ruleErr.ErrorCode, err)
		}

		return nil, fmt.Errorf("error checking sweepTx sanity: %v", err)
	}

	// With all the inputs in place, use each output's unique input script
	// function to generate the final witness required for spending.
	addInputScript := func(idx int, tso Input) error {
		inputScript, err := tso.CraftInputScript(
			signer, sweepTx, idx,
		)
		if err != nil {
			return fmt.Errorf("error building witness for input %d of "+
				"type %s: %v", idx, tso.WitnessType().String(), err)
		}

		sigScript, err := lnwallet.WitnessStackToSigScript(inputScript.Witness)
		if err != nil {
			return err
		}
		sweepTx.TxIn[idx].SignatureScript = sigScript

		return nil
	}

	// Finally we'll attach a valid input script to each csv and cltv input
	// within the sweeping transaction.
	for i, input := range inputs {
		if err := addInputScript(i, input); err != nil {
			return nil, err
		}
	}

	return sweepTx, nil
}

// getInputSigScriptSizeUpperBound returns the maximum length of the sig script
// for the given input if it would be included in a tx.
func getInputSigScriptSizeUpperBound(input Input) (int64, error) {
	switch input.WitnessType() {

	// Outputs on a remote commitment transaction that pay directly
	// to us.
	case lnwallet.CommitmentNoDelay:
		return lnwallet.P2PKHSigScriptSize, nil

	// Outputs on a past commitment transaction that pay directly
	// to us.
	case lnwallet.CommitmentTimeLock:
		return lnwallet.ToLocalTimeoutSigScriptSize, nil

	// Outgoing second layer HTLC's that have confirmed within the
	// chain, and the output they produced is now mature enough to
	// sweep.
	case lnwallet.HtlcOfferedTimeoutSecondLevel:
		return lnwallet.ToLocalTimeoutSigScriptSize, nil

	// Incoming second layer HTLC's that have confirmed within the
	// chain, and the output they produced is now mature enough to
	// sweep.
	case lnwallet.HtlcAcceptedSuccessSecondLevel:
		return lnwallet.ToLocalTimeoutSigScriptSize, nil

	// An HTLC on the commitment transaction of the remote party,
	// that has had its absolute timelock expire.
	case lnwallet.HtlcOfferedRemoteTimeout:
		return lnwallet.AcceptedHtlcTimeoutSigScriptSize, nil

	// An HTLC on the commitment transaction of the remote party,
	// that can be swept with the preimage.
	case lnwallet.HtlcAcceptedRemoteSuccess:
		return lnwallet.OfferedHtlcSuccessSigScriptSize, nil

	// A standard p2pkh signature script.
	case lnwallet.PublicKeyHash:
		return lnwallet.P2PKHSigScriptSize, nil

	}

	return 0, fmt.Errorf("unexpected witness type: %v", input.WitnessType())
}

// getSizeEstimate returns a size estimate for the given inputs.
// Additionally, it returns counts for the number of csv and cltv inputs.
func getSizeEstimate(inputs []Input) ([]Input, int64, int, int) {
	// We initialize a size estimator so we can accurately asses the
	// amount of fees we need to pay for this sweep transaction.
	//
	// TODO(roasbeef): can be more intelligent about buffering outputs to
	// be more efficient on-chain.
	var sizeEstimate lnwallet.TxSizeEstimator

	// Our sweep transaction will pay to a single p2pkh address,
	// ensure it contributes to our size estimate.
	sizeEstimate.AddP2PKHOutput()

	// For each output, use its witness type to determine the estimate
	// size of its witness, and add it to the proper set of spendable
	// outputs.
	var (
		csvCount, cltvCount int
	)
	sweepInputs := make([]Input, 0, len(inputs))
	for i := range inputs {
		input := inputs[i]

		size, err := getInputSigScriptSizeUpperBound(input)
		if err != nil {
			log.Warn(err)

			// Skip inputs for which no size estimate can be
			// given.
			continue
		}
		sizeEstimate.AddCustomInput(size)

		switch input.WitnessType() {
		case lnwallet.CommitmentTimeLock,
			lnwallet.HtlcOfferedTimeoutSecondLevel,
			lnwallet.HtlcAcceptedSuccessSecondLevel:
			csvCount++
		case lnwallet.HtlcOfferedRemoteTimeout:
			cltvCount++
		}
		sweepInputs = append(sweepInputs, input)
	}

	txSize := sizeEstimate.Size()

	return sweepInputs, txSize, csvCount, cltvCount
}
