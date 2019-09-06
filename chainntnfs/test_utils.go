// +build dev

package chainntnfs

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/chaincfg/v2"
	"github.com/decred/dcrd/dcrec/secp256k1"
	"github.com/decred/dcrd/dcrjson/v2"
	"github.com/decred/dcrd/dcrutil/v2"
	"github.com/decred/dcrd/rpctest"
	"github.com/decred/dcrd/txscript/v2"
	"github.com/decred/dcrd/wire"
)

var (
	// trickleInterval is the interval at which the miner should trickle
	// transactions to its peers. We'll set it small to ensure the miner
	// propagates transactions quickly in the tests.
	trickleInterval = 10 * time.Millisecond
)

var (
	NetParams = chaincfg.RegNetParams()
)

// randPubKeyHashScript generates a P2PKH script that pays to the public key of
// a randomly-generated private key.
func randPubKeyHashScript() ([]byte, *secp256k1.PrivateKey, error) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, nil, err
	}
	addrPk, _ := dcrutil.NewAddressSecpPubKey(
		privKey.PubKey().SerializeCompressed(), NetParams,
	)
	testAddr := addrPk.AddressPubKeyHash()

	pkScript, err := txscript.PayToAddrScript(testAddr)
	if err != nil {
		return nil, nil, err
	}
	return pkScript, privKey, nil
}

// GetTestTxidAndScript generate a new test transaction and returns its txid and
// the script of the output being generated.
func GetTestTxidAndScript(h *rpctest.Harness) (*chainhash.Hash, []byte, error) {
	pkScript, _, err := randPubKeyHashScript()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate pkScript: %v", err)
	}
	output := &wire.TxOut{Value: 2e8, PkScript: pkScript}
	txid, err := h.SendOutputs([]*wire.TxOut{output}, 10)
	if err != nil {
		return nil, nil, err
	}

	return txid, pkScript, nil
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
func CreateSpendableOutput(t *testing.T,
	miner *rpctest.Harness) (*wire.OutPoint, *wire.TxOut, *secp256k1.PrivateKey) {

	t.Helper()

	// Create a transaction that only has one output, the one destined for
	// the recipient.
	pkScript, privKey, err := randPubKeyHashScript()
	if err != nil {
		t.Fatalf("unable to generate pkScript: %v", err)
	}
	output := &wire.TxOut{Value: 2e8, PkScript: pkScript}
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

	return wire.NewOutPoint(txid, 0, wire.TxTreeRegular), output, privKey
}

// CreateSpendTx creates a transaction spending the specified output.
func CreateSpendTx(t *testing.T, prevOutPoint *wire.OutPoint,
	prevOutput *wire.TxOut, privKey *secp256k1.PrivateKey) *wire.MsgTx {

	t.Helper()

	spendingTx := wire.NewMsgTx()
	spendingTx.Version = 1
	spendingTx.AddTxIn(&wire.TxIn{PreviousOutPoint: *prevOutPoint})
	spendingTx.AddTxOut(&wire.TxOut{Value: 1e8, PkScript: prevOutput.PkScript})

	sigScript, err := txscript.SignatureScript(
		spendingTx, 0, prevOutput.PkScript, txscript.SigHashAll,
		privKey, true,
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
