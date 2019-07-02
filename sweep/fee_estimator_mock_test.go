package sweep

import (
	"sync"

	"github.com/decred/dcrlnd/lnwallet"
)

// mockFeeEstimator implements a mock fee estimator. It closely resembles
// lnwallet.StaticFeeEstimator with the addition that fees can be changed for
// testing purposes in a thread safe manner.
type mockFeeEstimator struct {
	feePerKB lnwallet.AtomPerKByte

	relayFee lnwallet.AtomPerKByte

	blocksToFee map[uint32]lnwallet.AtomPerKByte

	// A closure that when set is used instead of the
	// mockFeeEstimator.EstimateFeePerKW method.
	estimateFeePerKW func(numBlocks uint32) (lnwallet.AtomPerKByte, error)

	lock sync.Mutex
}

func newMockFeeEstimator(feePerKB,
	relayFee lnwallet.AtomPerKByte) *mockFeeEstimator {

	return &mockFeeEstimator{
		feePerKB:    feePerKB,
		relayFee:    relayFee,
		blocksToFee: make(map[uint32]lnwallet.AtomPerKByte),
	}
}

func (e *mockFeeEstimator) updateFees(feePerKB,
	relayFee lnwallet.AtomPerKByte) {

	e.lock.Lock()
	defer e.lock.Unlock()

	e.feePerKB = feePerKB
	e.relayFee = relayFee
}

func (e *mockFeeEstimator) EstimateFeePerKB(numBlocks uint32) (
	lnwallet.AtomPerKByte, error) {

	e.lock.Lock()
	defer e.lock.Unlock()

	if e.estimateFeePerKW != nil {
		return e.estimateFeePerKW(numBlocks)
	}

	if fee, ok := e.blocksToFee[numBlocks]; ok {
		return fee, nil
	}

	return e.feePerKB, nil
}

func (e *mockFeeEstimator) RelayFeePerKB() lnwallet.AtomPerKByte {
	e.lock.Lock()
	defer e.lock.Unlock()

	return e.relayFee
}

func (e *mockFeeEstimator) Start() error {
	return nil
}

func (e *mockFeeEstimator) Stop() error {
	return nil
}

var _ lnwallet.FeeEstimator = (*mockFeeEstimator)(nil)
