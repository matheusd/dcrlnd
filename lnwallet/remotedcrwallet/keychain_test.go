package remotedcrwallet

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/decred/dcrlnd/channeldb"
	"github.com/decred/dcrlnd/keychain"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/chaincfg/v2"
	"github.com/decred/dcrd/hdkeychain/v2"
	walletloader "github.com/decred/dcrwallet/loader"
	"github.com/decred/dcrwallet/wallet/v3/txrules"
)

var (
	testHDSeed = chainhash.Hash{
		0xb7, 0x94, 0x38, 0x5f, 0x2d, 0x1e, 0xf7, 0xab,
		0x4d, 0x92, 0x73, 0xd1, 0x90, 0x63, 0x81, 0xb4,
		0x4f, 0x2f, 0x6f, 0x25, 0x98, 0xa3, 0xef, 0xb9,
		0x69, 0x49, 0x18, 0x83, 0x31, 0x98, 0x47, 0x53,
	}
)

func createTestWallet() (func(), *hdkeychain.ExtendedKey, *channeldb.DB, error) {
	tempDir, err := ioutil.TempDir("", "keyring-lnwallet")
	if err != nil {
		return nil, nil, nil, err
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
		return nil, nil, nil, err
	}

	if err := baseWallet.Unlock(pass, nil); err != nil {
		return nil, nil, nil, err
	}

	// Create the temp chandb dir.
	cdbDir, err := ioutil.TempDir("", "channeldb")
	if err != nil {
		return nil, nil, nil, err
	}

	// Next, create channeldb for the first time.
	cdb, err := channeldb.Open(cdbDir)
	if err != nil {
		return nil, nil, nil, err
	}

	// Now derive a root master xpriv for the tests from the default wallet
	// account's hardened branch key at index 1017.
	acctXpriv, err := baseWallet.MasterPrivKey(0)
	if err != nil {
		return nil, nil, nil, err
	}
	rootXPriv, err := acctXpriv.Child(hdkeychain.HardenedKeyStart + 1017)
	if err != nil {
		return nil, nil, nil, err
	}

	cleanUp := func() {
		baseWallet.Lock()
		os.RemoveAll(tempDir)
		cdb.Close()
		os.RemoveAll(cdbDir)
	}

	return cleanUp, rootXPriv, cdb, nil
}

// TestDcrwalletKeyRingImpl tests whether the walletKeyRing implementation
// conforms to the required interface spec.
func TestDcrwalletKeyRingImpl(t *testing.T) {
	t.Parallel()

	keychain.CheckKeyRingImpl(t,
		func() (string, func(), keychain.KeyRing, error) {
			cleanUp, rootXPriv, cdb, err := createTestWallet()
			if err != nil {
				t.Fatalf("unable to create wallet: %v", err)
			}

			keyRing, err := newRemoteWalletKeyRing(rootXPriv, cdb)

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
			cleanUp, rootXPriv, cdb, err := createTestWallet()
			if err != nil {
				t.Fatalf("unable to create wallet: %v", err)
			}

			keyRing, err := newRemoteWalletKeyRing(rootXPriv, cdb)

			return "dcrwallet", cleanUp, keyRing, err
		},
	)

}

func init() {
	// We'll clamp the max range scan to constrain the run time of the
	// private key scan test.
	keychain.MaxKeyRangeScan = 3
}
