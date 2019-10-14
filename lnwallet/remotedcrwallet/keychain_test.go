package remotedcrwallet

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/decred/dcrlnd/channeldb"
	"github.com/decred/dcrlnd/keychain"
	"github.com/decred/dcrlnd/lnwallet"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/chaincfg/v2"
	"github.com/decred/dcrd/dcrutil/v2"
	"github.com/decred/dcrd/hdkeychain/v2"
	walletloader "github.com/decred/dcrwallet/loader"
	base "github.com/decred/dcrwallet/wallet/v3"
	"github.com/decred/dcrwallet/wallet/v3/txrules"
	"github.com/decred/dcrwallet/wallet/v3/udb"
)

var (
	testHDSeed = chainhash.Hash{
		0xb7, 0x94, 0x38, 0x5f, 0x2d, 0x1e, 0xf7, 0xab,
		0x4d, 0x92, 0x73, 0xd1, 0x90, 0x63, 0x81, 0xb4,
		0x4f, 0x2f, 0x6f, 0x25, 0x98, 0xa3, 0xef, 0xb9,
		0x69, 0x49, 0x18, 0x83, 0x31, 0x98, 0x47, 0x53,
	}
)

type mockOnchainAddrSourcer struct {
	w *base.Wallet
}

func (mas *mockOnchainAddrSourcer) NewAddress(t lnwallet.AddressType, change bool) (dcrutil.Address, error) {
	var addr dcrutil.Address
	var err error
	if change {
		addr, err = mas.w.NewInternalAddress(context.TODO(), 0)
	} else {
		addr, err = mas.w.NewExternalAddress(context.TODO(), 0)
	}

	if err != nil {
		return nil, err
	}

	// Convert to a regular p2pkh address, since the addresses returned are
	// used as paramaters to PayToScriptAddress() which doesn't understand
	// the native wallet types.
	return dcrutil.DecodeAddress(addr.Address(), chaincfg.RegNetParams())

}
func (mas *mockOnchainAddrSourcer) Bip44AddressInfo(addr dcrutil.Address) (uint32, uint32, uint32, error) {
	info, err := mas.w.AddressInfo(addr)
	if err != nil {
		return 0, 0, 0, nil
	}

	switch ma := info.(type) {
	case udb.ManagedPubKeyAddress:
		branch := uint32(0)
		if ma.Internal() {
			branch = 1
		}
		return ma.Account(), branch, ma.Index(), nil
	}

	return 0, 0, 0, fmt.Errorf("unkown address type")
}

func createTestWallet() (func(), *hdkeychain.ExtendedKey, *channeldb.DB, onchainAddrSourcer, error) {
	tempDir, err := ioutil.TempDir("", "keyring-lnwallet")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	loader := walletloader.NewLoader(chaincfg.RegNetParams(), tempDir,
		&walletloader.StakeOptions{}, 20, false,
		txrules.DefaultRelayFeePerKb.ToCoin(), 5,
		false)

	pass := []byte("test")

	baseWallet, err := loader.CreateNewWallet(
		pass, pass, testHDSeed[:],
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	if err := baseWallet.Unlock(pass, nil); err != nil {
		return nil, nil, nil, nil, err
	}

	// Create the temp chandb dir.
	cdbDir, err := ioutil.TempDir("", "channeldb")
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Next, create channeldb for the first time.
	cdb, err := channeldb.Open(cdbDir)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// The root master xpriv is the default account's one.
	acctXpriv, err := baseWallet.MasterPrivKey(0)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Create the mock onchain addresses sourcer linked to the previously
	// created wallet.
	addrSourcer := &mockOnchainAddrSourcer{w: baseWallet}

	cleanUp := func() {
		baseWallet.Lock()
		os.RemoveAll(tempDir)
		cdb.Close()
		os.RemoveAll(cdbDir)
	}

	return cleanUp, acctXpriv, cdb, addrSourcer, nil
}

// TestDcrwalletKeyRingImpl tests whether the walletKeyRing implementation
// conforms to the required interface spec.
func TestDcrwalletKeyRingImpl(t *testing.T) {
	t.Parallel()

	keychain.CheckKeyRingImpl(t,
		func() (string, func(), keychain.KeyRing, error) {
			cleanUp, rootXPriv, cdb, addrSourcer, err := createTestWallet()
			if err != nil {
				t.Fatalf("unable to create wallet: %v", err)
			}

			keyRing, err := newRemoteWalletKeyRing(
				rootXPriv, cdb, addrSourcer,
			)

			return "dcrwallet", cleanUp, keyRing, err
		},
	)

}

// TestDcrwalletSecretKeyRingImpl tests whether the walletKeyRing
// implementation conforms to the required interface spec.
func TestDcrwalletSecretKeyRingImpl(t *testing.T) {
	t.Parallel()

	keychain.CheckSecretKeyRingImpl(t,
		func() (string, func(), keychain.SecretKeyRing, error) {
			cleanUp, rootXPriv, cdb, addrSourcer, err := createTestWallet()
			if err != nil {
				t.Fatalf("unable to create wallet: %v", err)
			}

			keyRing, err := newRemoteWalletKeyRing(
				rootXPriv, cdb, addrSourcer,
			)

			return "dcrwallet", cleanUp, keyRing, err
		},
	)

}

func init() {
	// We'll clamp the max range scan to constrain the run time of the
	// private key scan test.
	keychain.MaxKeyRangeScan = 3
}
