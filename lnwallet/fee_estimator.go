package lnwallet

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	prand "math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/decred/dcrd/dcrutil/v2"
	dcrjson "github.com/decred/dcrd/rpc/jsonrpc/types/v2"
	"github.com/decred/dcrd/rpcclient/v5"
)

const (
	// FeePerKBFloor is the lowest fee rate in atom/kB that we should use
	// for determining transaction fees. Originally, this was used due to
	// the conversion from sats/kB => sats/KW causing a possible rounding
	// error, but in Decred we use this to track the widely deployed
	// minimum relay fee.
	FeePerKBFloor AtomPerKByte = 1e4

	// maxBlockTarget is the highest number of blocks confirmations that
	// a WebAPIFeeEstimator will cache fees for. This number is chosen
	// because it's the highest number of confs bitcoind will return a fee
	// estimate for.
	maxBlockTarget uint32 = 1009

	// minBlockTarget is the lowest number of blocks confirmations that
	// a WebAPIFeeEstimator will cache fees for. Requesting an estimate for
	// less than this will result in an error.
	minBlockTarget uint32 = 2

	// minFeeUpdateTimeout represents the minimum interval in which a
	// WebAPIFeeEstimator will request fresh fees from its API.
	minFeeUpdateTimeout = 5 * time.Minute

	// maxFeeUpdateTimeout represents the maximum interval in which a
	// WebAPIFeeEstimator will request fresh fees from its API.
	maxFeeUpdateTimeout = 20 * time.Minute
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

// SparseConfFeeSource is an implementation of the WebAPIFeeSource that utilizes
// a user-specified fee estimation API for Bitcoin. It expects the response
// to be in the JSON format: `fee_by_block_target: { ... }` where the value maps
// block targets to fee estimates (in sat per kilovbyte).
type SparseConfFeeSource struct {
	// URL is the fee estimation API specified by the user.
	URL string
}

// GenQueryURL generates the full query URL. The value returned by this
// method should be able to be used directly as a path for an HTTP GET
// request.
//
// NOTE: Part of the WebAPIFeeSource interface.
func (s SparseConfFeeSource) GenQueryURL() string {
	return s.URL
}

// ParseResponse attempts to parse the body of the response generated by the
// above query URL. Typically this will be JSON, but the specifics are left to
// the WebAPIFeeSource implementation.
//
// NOTE: Part of the WebAPIFeeSource interface.
func (s SparseConfFeeSource) ParseResponse(r io.Reader) (map[uint32]uint32, error) {
	type jsonResp struct {
		FeeByBlockTarget map[uint32]uint32 `json:"fee_by_block_target"`
	}

	resp := jsonResp{
		FeeByBlockTarget: make(map[uint32]uint32),
	}
	jsonReader := json.NewDecoder(r)
	if err := jsonReader.Decode(&resp); err != nil {
		return nil, err
	}

	return resp.FeeByBlockTarget, nil
}

// A compile-time assertion to ensure that SparseConfFeeSource implements the
// WebAPIFeeSource interface.
var _ WebAPIFeeSource = (*SparseConfFeeSource)(nil)

// WebAPIFeeEstimator is an implementation of the FeeEstimator interface that
// queries an HTTP-based fee estimation from an existing web API.
type WebAPIFeeEstimator struct {
	started sync.Once
	stopped sync.Once

	// apiSource is the backing web API source we'll use for our queries.
	apiSource WebAPIFeeSource

	// updateFeeTicker is the ticker responsible for updating the Estimator's
	// fee estimates every time it fires.
	updateFeeTicker *time.Ticker

	// feeByBlockTarget is our cache for fees pulled from the API. When a
	// fee estimate request comes in, we pull the estimate from this array
	// rather than re-querying the API, to prevent an inadvertent DoS attack.
	feesMtx          sync.Mutex
	feeByBlockTarget map[uint32]uint32

	// defaultFeePerKB is a fallback value that we'll use if we're unable
	// to query the API for any reason.
	defaultFeePerKB AtomPerKByte

	quit chan struct{}
	wg   sync.WaitGroup
}

// NewWebAPIFeeEstimator creates a new WebAPIFeeEstimator from a given URL and a
// fallback default fee. The fees are updated whenever a new block is mined.
func NewWebAPIFeeEstimator(
	api WebAPIFeeSource, defaultFee AtomPerKByte) *WebAPIFeeEstimator {

	return &WebAPIFeeEstimator{
		apiSource:        api,
		feeByBlockTarget: make(map[uint32]uint32),
		defaultFeePerKB:  defaultFee,
		quit:             make(chan struct{}),
	}
}

// EstimateFeePerKB takes in a target for the number of blocks until an initial
// confirmation and returns the estimated fee expressed in sat/kw.
//
// NOTE: This method is part of the FeeEstimator interface.
func (w *WebAPIFeeEstimator) EstimateFeePerKB(numBlocks uint32) (AtomPerKByte, error) {
	if numBlocks > maxBlockTarget {
		numBlocks = maxBlockTarget
	} else if numBlocks < minBlockTarget {
		return 0, fmt.Errorf("conf target of %v is too low, minimum "+
			"accepted is %v", numBlocks, minBlockTarget)
	}

	feePerKB, err := w.getCachedFee(numBlocks)
	if err != nil {
		return 0, err
	}
	atomsPerKB := AtomPerKByte(feePerKB)

	// If the result is too low, then we'll clamp it to our current fee
	// floor.
	if atomsPerKB < FeePerKBFloor {
		atomsPerKB = FeePerKBFloor
	}

	walletLog.Debugf("Web API returning %v atoms/KB for conf target of %v",
		atomsPerKB, numBlocks)

	return atomsPerKB, nil
}

// Start signals the FeeEstimator to start any processes or goroutines it needs
// to perform its duty.
//
// NOTE: This method is part of the FeeEstimator interface.
func (w *WebAPIFeeEstimator) Start() error {
	var err error
	w.started.Do(func() {
		walletLog.Infof("Starting web API fee estimator")

		w.updateFeeTicker = time.NewTicker(w.randomFeeUpdateTimeout())
		w.updateFeeEstimates()

		w.wg.Add(1)
		go w.feeUpdateManager()

	})
	return err
}

// Stop stops any spawned goroutines and cleans up the resources used by the
// fee estimator.
//
// NOTE: This method is part of the FeeEstimator interface.
func (w *WebAPIFeeEstimator) Stop() error {
	w.stopped.Do(func() {
		walletLog.Infof("Stopping web API fee estimator")

		w.updateFeeTicker.Stop()

		close(w.quit)
		w.wg.Wait()
	})
	return nil
}

// RelayFeePerKB returns the minimum fee rate required for transactions to be
// relayed.
//
// NOTE: This method is part of the FeeEstimator interface.
func (w *WebAPIFeeEstimator) RelayFeePerKB() AtomPerKByte {
	return FeePerKBFloor
}

// randomFeeUpdateTimeout returns a random timeout between minFeeUpdateTimeout
// and maxFeeUpdateTimeout that will be used to determine how often the Estimator
// should retrieve fresh fees from its API.
func (w *WebAPIFeeEstimator) randomFeeUpdateTimeout() time.Duration {
	lower := int64(minFeeUpdateTimeout)
	upper := int64(maxFeeUpdateTimeout)
	return time.Duration(prand.Int63n(upper-lower) + lower)
}

// getCachedFee takes in a target for the number of blocks until an initial
// confirmation and returns an estimated fee (if one was returned by the API). If
// the fee was not previously cached, we cache it here.
func (w *WebAPIFeeEstimator) getCachedFee(numBlocks uint32) (uint32, error) {
	w.feesMtx.Lock()
	defer w.feesMtx.Unlock()

	// Search our cached fees for the desired block target. If the target is
	// not cached, then attempt to extrapolate it from the next lowest target
	// that *is* cached. If we successfully extrapolate, then cache the
	// target's fee.
	for target := numBlocks; target >= minBlockTarget; target-- {
		fee, ok := w.feeByBlockTarget[target]
		if !ok {
			continue
		}

		_, ok = w.feeByBlockTarget[numBlocks]
		if !ok {
			w.feeByBlockTarget[numBlocks] = fee
		}
		return fee, nil
	}
	return 0, fmt.Errorf("web API does not include a fee estimation for "+
		"block target of %v", numBlocks)
}

// updateFeeEstimates re-queries the API for fresh fees and caches them.
func (w *WebAPIFeeEstimator) updateFeeEstimates() {
	// Rather than use the default http.Client, we'll make a custom one
	// which will allow us to control how long we'll wait to read the
	// response from the service. This way, if the service is down or
	// overloaded, we can exit early and use our default fee.
	netTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	netClient := &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}

	// With the client created, we'll query the API source to fetch the URL
	// that we should use to query for the fee estimation.
	targetURL := w.apiSource.GenQueryURL()
	resp, err := netClient.Get(targetURL)
	if err != nil {
		walletLog.Errorf("unable to query web api for fee response: %v",
			err)
		return
	}
	defer resp.Body.Close()

	// Once we've obtained the response, we'll instruct the WebAPIFeeSource
	// to parse out the body to obtain our final result.
	feesByBlockTarget, err := w.apiSource.ParseResponse(resp.Body)
	if err != nil {
		walletLog.Errorf("unable to query web api for fee response: %v",
			err)
		return
	}

	w.feesMtx.Lock()
	w.feeByBlockTarget = feesByBlockTarget
	w.feesMtx.Unlock()
}

// feeUpdateManager updates the fee estimates whenever a new block comes in.
func (w *WebAPIFeeEstimator) feeUpdateManager() {
	defer w.wg.Done()

	for {
		select {
		case <-w.updateFeeTicker.C:
			w.updateFeeEstimates()
		case <-w.quit:
			return
		}
	}
}

// A compile-time assertion to ensure that WebAPIFeeEstimator implements the
// FeeEstimator interface.
var _ FeeEstimator = (*WebAPIFeeEstimator)(nil)
