package lnwallet

import (
	"context"

	"github.com/decred/dcrd/dcrjson/v2"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/rpcclient/v2"
)

const (
	// FeePerKBFloor is the lowest fee rate in atom/kB that we should use
	// for determining transaction fees. Originally, this was used due to
	// the conversion from sats/kB => sats/KW causing a possible rounding
	// error, but in Decred we use this to track the widely deployed
	// minimum relay fee.
	FeePerKBFloor AtomPerKByte = 1e4
)

// AtomPerKByte represents a fee rate in atom/kB.
type AtomPerKByte dcrutil.Amount

// FeeForSize calculates the fee resulting from this fee rate and the given
// size in bytes.
func (s AtomPerKByte) FeeForSize(bytes int64) dcrutil.Amount {
	return dcrutil.Amount(s) * dcrutil.Amount(bytes) / 1000
}

// String returns a pretty string representation for the rate in DCR/kB.
func (s AtomPerKByte) String() string {
	return dcrutil.Amount(s).String() + "/kB"
}

// FeeEstimator provides the ability to estimate on-chain transaction fees for
// various combinations of transaction sizes and desired confirmation time
// (measured by number of blocks).
type FeeEstimator interface {
	// EstimateFeePerKB takes in a target for the number of blocks until
	// an initial confirmation and returns the estimated fee expressed in
	// atoms/byte.
	EstimateFeePerKB(numBlocks uint32) (AtomPerKByte, error)

	// Start signals the FeeEstimator to start any processes or goroutines
	// it needs to perform its duty.
	Start() error

	// Stop stops any spawned goroutines and cleans up the resources used
	// by the fee estimator.
	Stop() error

	// RelayFeePerKB returns the minimum fee rate required for transactions
	// to be relayed. This is also the basis for calculation of the dust
	// limit.
	RelayFeePerKB() AtomPerKByte
}

// StaticFeeEstimator will return a static value for all fee calculation
// requests. It is designed to be replaced by a proper fee calculation
// implementation. The fees are not accessible directly, because changing them
// would not be thread safe.
type StaticFeeEstimator struct {
	// feePerKB is the static fee rate in atoms-per-kB that will be
	// returned by this fee estimator.
	feePerKB AtomPerKByte

	// relayFee is the minimum fee rate required for transactions to be
	// relayed.
	relayFee AtomPerKByte
}

// NewStaticFeeEstimator returns a new static fee estimator instance.
func NewStaticFeeEstimator(feePerKB,
	relayFee AtomPerKByte) *StaticFeeEstimator {

	return &StaticFeeEstimator{
		feePerKB: feePerKB,
		relayFee: relayFee,
	}
}

// EstimateFeePerKB will return the static value for fee calculations.
//
// NOTE: This method is part of the FeeEstimator interface.
func (e StaticFeeEstimator) EstimateFeePerKB(numBlocks uint32) (AtomPerKByte, error) {
	return e.feePerKB, nil
}

// RelayFeePerKB returns the minimum fee rate required for transactions to be
// relayed.
//
// NOTE: This method is part of the FeeEstimator interface.
func (e StaticFeeEstimator) RelayFeePerKB() AtomPerKByte {
	return e.relayFee
}

// Start signals the FeeEstimator to start any processes or goroutines
// it needs to perform its duty.
//
// NOTE: This method is part of the FeeEstimator interface.
func (e StaticFeeEstimator) Start() error {
	return nil
}

// Stop stops any spawned goroutines and cleans up the resources used
// by the fee estimator.
//
// NOTE: This method is part of the FeeEstimator interface.
func (e StaticFeeEstimator) Stop() error {
	return nil
}

// A compile-time assertion to ensure that StaticFeeEstimator implements the
// FeeEstimator interface.
var _ FeeEstimator = (*StaticFeeEstimator)(nil)

// DcrdFeeEstimator is an implementation of the FeeEstimator interface backed
// by the RPC interface of an active dcrd node. This implementation will proxy
// any fee estimation requests to dcrd's RPC interface.
type DcrdFeeEstimator struct {
	// fallbackFeePerKB is the fall back fee rate in atoms/kB that is returned
	// if the fee estimator does not yet have enough data to actually
	// produce fee estimates.
	fallbackFeePerKB AtomPerKByte

	// minFeePerKB is the minimum fee, in atoms/kB, that we should enforce.
	// This will be used as the default fee rate for a transaction when the
	// estimated fee rate is too low to allow the transaction to propagate
	// through the network.
	minFeePerKB AtomPerKByte

	dcrdConn *rpcclient.Client
}

// NewDcrdFeeEstimator creates a new DcrdFeeEstimator given a fully populated
// rpc config that is able to successfully connect and authenticate with the
// dcrd node, and also a fall back fee rate. The fallback fee rate is used in
// the occasion that the estimator has insufficient data, or returns zero for a
// fee estimate.
func NewDcrdFeeEstimator(rpcConfig rpcclient.ConnConfig,
	fallBackFeeRate AtomPerKByte) (*DcrdFeeEstimator, error) {

	rpcConfig.DisableConnectOnNew = true
	rpcConfig.DisableAutoReconnect = false
	chainConn, err := rpcclient.New(&rpcConfig, nil)
	if err != nil {
		return nil, err
	}

	return &DcrdFeeEstimator{
		fallbackFeePerKB: fallBackFeeRate,
		dcrdConn:         chainConn,
	}, nil
}

// Start signals the FeeEstimator to start any processes or goroutines
// it needs to perform its duty.
//
// NOTE: This method is part of the FeeEstimator interface.
func (b *DcrdFeeEstimator) Start() error {
	ctx := context.Background()
	if err := b.dcrdConn.Connect(ctx, true); err != nil {
		return err
	}

	// Once the connection to the backend node has been established, we'll
	// query it for its minimum relay fee.
	info, err := b.dcrdConn.GetInfo()
	if err != nil {
		return err
	}

	relayFee, err := dcrutil.NewAmount(info.RelayFee)
	if err != nil {
		return err
	}

	// By default, we'll use the backend node's minimum relay fee as the
	// minimum fee rate we'll propose for transactions. However, if this
	// happens to be lower than our fee floor, we'll enforce that instead.
	b.minFeePerKB = AtomPerKByte(relayFee)
	if b.minFeePerKB < FeePerKBFloor {
		walletLog.Warnf("Dcrd returned fee rate of %s which is "+
			"lower than the floor rate", b.minFeePerKB)
		b.minFeePerKB = FeePerKBFloor
	}

	walletLog.Debugf("Using minimum fee rate of %s",
		b.minFeePerKB)

	return nil
}

// Stop stops any spawned goroutines and cleans up the resources used
// by the fee estimator.
//
// NOTE: This method is part of the FeeEstimator interface.
func (b *DcrdFeeEstimator) Stop() error {
	b.dcrdConn.Shutdown()

	return nil
}

// RelayFeePerKB returns the minimum fee rate required for transactions to be
// relayed.
//
// NOTE: This method is part of the FeeEstimator interface.
func (b *DcrdFeeEstimator) RelayFeePerKB() AtomPerKByte {
	return b.minFeePerKB
}

// EstimateFeePerKB queries the connected chain client for a fee estimation for
// the given block range.
//
// NOTE: This method is part of the FeeEstimator interface.
func (b *DcrdFeeEstimator) EstimateFeePerKB(numBlocks uint32) (AtomPerKByte, error) {
	feeEstimate, err := b.fetchEstimate(numBlocks)
	switch {
	// If the estimator doesn't have enough data, or returns an error, then
	// to return a proper value, then we'll return the default fall back
	// fee rate.
	case err != nil:
		walletLog.Errorf("unable to query estimator: %v", err)
		fallthrough

	case feeEstimate == 0:
		return b.fallbackFeePerKB, nil
	}

	return feeEstimate, nil
}

// fetchEstimate returns a fee estimate for a transaction to be confirmed in
// confTarget blocks. The estimate is returned in atom/kB.
func (b *DcrdFeeEstimator) fetchEstimate(confTarget uint32) (AtomPerKByte, error) {
	// First, we'll fetch the estimate for our confirmation target.
	dcrPerKB, err := b.dcrdConn.EstimateSmartFee(int64(confTarget),
		dcrjson.EstimateSmartFeeConservative)
	if err != nil {
		return 0, err
	}

	// Next, we'll convert the returned value to atoms, as it's
	// currently returned in DCR.
	atoms, err := dcrutil.NewAmount(dcrPerKB)
	if err != nil {
		return 0, err
	}
	atomsPerKB := AtomPerKByte(atoms)

	// Finally, we'll enforce our fee floor.
	if atomsPerKB < b.minFeePerKB {
		walletLog.Debugf("Estimated fee rate of %s is too low, "+
			"using fee floor of %s", atomsPerKB,
			b.minFeePerKB)
		atomsPerKB = b.minFeePerKB
	}

	walletLog.Debugf("Returning %s for conf target of %d",
		atomsPerKB, confTarget)

	return atomsPerKB, nil
}

// A compile-time assertion to ensure that DcrdFeeEstimator implements the
// FeeEstimator interface.
var _ FeeEstimator = (*DcrdFeeEstimator)(nil)

// WebAPIFeeSource is an interface allows the WebAPIFeeEstimator to query an
// arbitrary HTTP-based fee estimator. Each new set/network will gain an
// implementation of this interface in order to allow the WebAPIFeeEstimator to
// be fully generic in its logic.
type WebAPIFeeSource interface {
	// GenQueryURL generates the full query URL. The value returned by this
	// method should be able to be used directly as a path for an HTTP GET
	// request.
	GenQueryURL() string

	// ParseResponse attempts to parse the body of the response generated
	// by the above query URL. Typically this will be JSON, but the
	// specifics are left to the WebAPIFeeSource implementation.
	ParseResponse(r io.Reader) (map[uint32]uint32, error)
}
