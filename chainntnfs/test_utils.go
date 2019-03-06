// +build dev

package chainntnfs

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrec/secp256k1"
	"github.com/decred/dcrd/dcrjson/v2"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/rpctest"
	"github.com/decred/dcrd/txscript"
	"github.com/decred/dcrd/wire"
)

var (
	// trickleInterval is the interval at which the miner should trickle
	// transactions to its peers. We'll set it small to ensure the miner
	// propagates transactions quickly in the tests.
	trickleInterval = 10 * time.Millisecond
)

var (
	NetParams = &chaincfg.RegNetParams

	testPrivKey = []byte{
		0x81, 0xb6, 0x37, 0xd8, 0xfc, 0xd2, 0xc6, 0xda,
		0x63, 0x59, 0xe6, 0x96, 0x31, 0x13, 0xa1, 0x17,
		0xd, 0xe7, 0x95, 0xe4, 0xb7, 0x25, 0xb8, 0x4d,
		0x1e, 0xb, 0x4c, 0xfd, 0x9e, 0xc5, 0x8c, 0xe9,
	}
	privKey, pubKey = secp256k1.PrivKeyFromBytes(testPrivKey)
	addrPk, _       = dcrutil.NewAddressSecpPubKey(
		pubKey.SerializeCompressed(), NetParams,
	)
	testAddr = addrPk.AddressPubKeyHash()
)

// GetTestTxidAndScript generate a new test transaction and returns its txid and
// the script of the output being generated.
func GetTestTxidAndScript(h *rpctest.Harness) (*chainhash.Hash, []byte, error) {
	script, err := txscript.PayToAddrScript(testAddr)
	if err != nil {
		return nil, nil, err
	}

	output := &wire.TxOut{Value: 2e8, PkScript: script}
	txid, err := h.SendOutputs([]*wire.TxOut{output}, 10)
	if err != nil {
		return nil, nil, err
	}

	return txid, script, nil
}

// WaitForMempoolTx waits for the txid to be seen in the miner's mempool.
func WaitForMempoolTx(miner *rpctest.Harness, txid *chainhash.Hash) error {
	timeout := time.After(10 * time.Second)
	trickle := time.After(2 * trickleInterval)
	for {
		// Check for the harness' knowledge of the txid.
		tx, err := miner.Node.GetRawTransaction(txid)
		if err != nil {
			jsonErr, ok := err.(*dcrjson.RPCError)
			if ok && jsonErr.Code == dcrjson.ErrRPCNoTxInfo {
				continue
			}
			return err
		}

		if tx != nil && tx.Hash().IsEqual(txid) {
			break
		}

		select {
		case <-time.After(100 * time.Millisecond):
		case <-timeout:
			return errors.New("timed out waiting for tx")
		}
	}

	// To ensure any transactions propagate from the miner to the peers
	// before returning, ensure we have waited for at least
	// 2*trickleInterval before returning.
	select {
	case <-trickle:
	case <-timeout:
		return errors.New("timeout waiting for trickle interval. " +
			"Trickle interval to large?")
	}

	return nil
}

// CreateSpendableOutput creates and returns an output that can be spent later
// on.
func CreateSpendableOutput(t *testing.T, miner *rpctest.Harness) (*wire.OutPoint, []byte) {
	t.Helper()

	// Create a transaction that only has one output, the one destined for
	// the recipient.
	script, err := txscript.PayToAddrScript(testAddr)
	if err != nil {
		t.Fatalf("unable to create p2pkh script: %v", err)
	}
	output := &wire.TxOut{Value: 2e8, PkScript: script}
	// TODO(decred): SendOutputsWithoutChange
	txid, err := miner.SendOutputs([]*wire.TxOut{output}, 10)
	if err != nil {
		t.Fatalf("unable to create tx: %v", err)
	}

	// Mine the transaction to mark the output as spendable.
	if err := WaitForMempoolTx(miner, txid); err != nil {
		t.Fatalf("tx not relayed to miner: %v", err)
	}
	if _, err := miner.Node.Generate(1); err != nil {
		t.Fatalf("unable to generate single block: %v", err)
	}

	return wire.NewOutPoint(txid, 0, wire.TxTreeRegular), script
}

// CreateSpendTx creates a transaction spending the specified output.
func CreateSpendTx(t *testing.T, outpoint *wire.OutPoint, pkScript []byte) *wire.MsgTx {
	t.Helper()

	spendingTx := wire.NewMsgTx()
	spendingTx.Version = 1
	spendingTx.AddTxIn(&wire.TxIn{PreviousOutPoint: *outpoint})
	spendingTx.AddTxOut(&wire.TxOut{Value: 1e8, PkScript: pkScript})

	sigScript, err := txscript.SignatureScript(
		spendingTx, 0, pkScript, txscript.SigHashAll, privKey, true,
	)
	if err != nil {
		t.Fatalf("unable to sign tx: %v", err)
	}
	spendingTx.TxIn[0].SignatureScript = sigScript

	return spendingTx
}

// NewMiner spawns a testing harness backed by a dcrd node that can serve as a
// miner.
func NewMiner(t *testing.T, extraArgs []string, createChain bool,
	spendableOutputs uint32) (*rpctest.Harness, func()) {

	t.Helper()

	// TODO(decred): Test and either remove or add as needed.
	//
	// Add the trickle interval argument to the extra args.
	trickle := fmt.Sprintf("--trickleinterval=%v", trickleInterval)
	//extraArgs = append(extraArgs, trickle)
	_ = trickle

	node, err := rpctest.New(NetParams, nil, extraArgs)
	if err != nil {
		t.Fatalf("unable to create backend node: %v", err)
	}
	if err := node.SetUp(createChain, spendableOutputs); err != nil {
		node.TearDown()
		t.Fatalf("unable to set up backend node: %v", err)
	}

	return node, func() { node.TearDown() }
}
