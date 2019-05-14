package keychain

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrec/secp256k1"
	"github.com/decred/dcrd/hdkeychain"
	walletloader "github.com/decred/dcrwallet/loader"
	"github.com/decred/dcrwallet/wallet/v2"
	"github.com/decred/dcrwallet/wallet/v2/txrules"

	_ "github.com/decred/dcrwallet/wallet/v2/drivers/bdb" // Required in order to create the default database.
)

// versionZeroKeyFamilies is a slice of all the known key families for first
// version of the key derivation schema defined in this package.
var versionZeroKeyFamilies = []KeyFamily{
	KeyFamilyMultiSig,
	KeyFamilyRevocationBase,
	KeyFamilyHtlcBase,
	KeyFamilyPaymentBase,
	KeyFamilyDelayBase,
	KeyFamilyRevocationRoot,
	KeyFamilyNodeKey,
}

var (
	testHDSeed = chainhash.Hash{
		0xb7, 0x94, 0x38, 0x5f, 0x2d, 0x1e, 0xf7, 0xab,
		0x4d, 0x92, 0x73, 0xd1, 0x90, 0x63, 0x81, 0xb4,
		0x4f, 0x2f, 0x6f, 0x25, 0x98, 0xa3, 0xef, 0xb9,
		0x69, 0x49, 0x18, 0x83, 0x31, 0x98, 0x47, 0x53,
	}
)

func createTestWallet() (func(), *wallet.Wallet, error) {
	tempDir, err := ioutil.TempDir("", "keyring-lnwallet")
	if err != nil {
		return nil, nil, err
	}
	loader := walletloader.NewLoader(&chaincfg.RegNetParams, tempDir,
		&walletloader.StakeOptions{}, wallet.DefaultGapLimit, false,
		txrules.DefaultRelayFeePerKb.ToCoin(), wallet.DefaultAccountGapLimit,
		false)

	pass := []byte("test")

	baseWallet, err := loader.CreateNewWallet(
		pass, pass, testHDSeed[:],
	)
	if err != nil {
		return nil, nil, err
	}

	if err := baseWallet.Unlock(pass, nil); err != nil {
		return nil, nil, err
	}

	cleanUp := func() {
		baseWallet.Lock()
		os.RemoveAll(tempDir)
	}

	return cleanUp, baseWallet, nil
}

// createTestHDKeyRing creates a test HDKeyRing implementation that works
// purely in memory, using the default test HD seed as root master private key.
func createTestHDKeyRing() *HDKeyRing {
	// We can ignore errors here because they could only be for invalid
	// keys/seeds and the default test seed does not produce errors.
	master, _ := hdkeychain.NewMaster(testHDSeed[:], &chaincfg.RegNetParams)
	masterPubs := make(map[KeyFamily]*hdkeychain.ExtendedKey,
		len(versionZeroKeyFamilies))
	indexes := make(map[KeyFamily]uint32, len(versionZeroKeyFamilies))
	for _, keyFam := range versionZeroKeyFamilies {
		masterPubs[keyFam], _ = master.Child(uint32(keyFam))
		indexes[keyFam] = 0
	}

	fetchMasterPriv := func(keyFam KeyFamily) (*hdkeychain.ExtendedKey, error) {
		// We never neutered the keys in masterPubs, so they also
		// contain the private key.
		return masterPubs[keyFam], nil
	}

	nextIndex := func(keyFam KeyFamily) (uint32, error) {
		index := indexes[keyFam]
		indexes[keyFam] += 1
		return index, nil
	}

	return NewHDKeyRing(masterPubs, fetchMasterPriv, nextIndex)
}

func assertEqualKeyLocator(t *testing.T, a, b KeyLocator) {
	t.Helper()
	if a != b {
		t.Fatalf("mismatched key locators: expected %v, "+
			"got %v", spew.Sdump(a), spew.Sdump(b))
	}
}

// secretKeyRingConstructor is a function signature that's used as a generic
// constructor for various implementations of the KeyRing interface. A string
// naming the returned interface, a function closure that cleans up any
// resources, and the clean up interface itself are to be returned.
type keyRingConstructor func() (string, func(), KeyRing, error)

// TestKeyRingDerivation tests that each known KeyRing implementation properly
// adheres to the expected behavior of the set of interfaces.
func TestKeyRingDerivation(t *testing.T) {
	t.Parallel()

	keyRingImplementations := []keyRingConstructor{
		func() (string, func(), KeyRing, error) {
			cleanUp, wallet, err := createTestWallet()
			if err != nil {
				t.Fatalf("unable to create wallet: %v", err)
			}

			keyRing := NewWalletKeyRing(wallet)

			return "dcrwallet", cleanUp, keyRing, nil
		},
		func() (string, func(), KeyRing, error) {
			return "hdkeyring", func() {}, createTestHDKeyRing(), nil
		},
	}

	const numKeysToDerive = 10

	// For each implementation constructor registered above, we'll execute
	// an identical set of tests in order to ensure that the interface
	// adheres to our nominal specification.
	for _, keyRingConstructor := range keyRingImplementations {
		keyRingName, cleanUp, keyRing, err := keyRingConstructor()
		if err != nil {
			t.Fatalf("unable to create key ring %v: %v", keyRingName,
				err)
		}
		defer cleanUp()

		success := t.Run(fmt.Sprintf("%v", keyRingName), func(t *testing.T) {
			// First, we'll ensure that we're able to derive keys
			// from each of the known key families.
			for _, keyFam := range versionZeroKeyFamilies {
				// First, we'll ensure that we can derive the
				// *next* key in the keychain.
				keyDesc, err := keyRing.DeriveNextKey(keyFam)
				if err != nil {
					t.Fatalf("unable to derive next for "+
						"keyFam=%v: %v", keyFam, err)
				}
				assertEqualKeyLocator(t,
					KeyLocator{
						Family: keyFam,
						Index:  0,
					}, keyDesc.KeyLocator,
				)

				// We'll generate the next key and ensure it's
				// different than the first one.
				keyDescNext, err := keyRing.DeriveNextKey(keyFam)
				if err != nil {
					t.Fatalf("unable to derive next for"+
						"keyFam=%v: %v", keyFam, err)
				}
				if keyDescNext.PubKey.IsEqual(keyDesc.PubKey) {
					t.Fatal("keyring derived two " +
						"identical consecutive keys")
				}

				// We'll now re-derive that key to ensure that
				// we're able to properly access the key via
				// the random access derivation methods.
				keyLoc := KeyLocator{
					Family: keyFam,
					Index:  0,
				}
				firstKeyDesc, err := keyRing.DeriveKey(keyLoc)
				if err != nil {
					t.Fatalf("unable to derive first key for "+
						"keyFam=%v: %v", keyFam, err)
				}
				if !keyDesc.PubKey.IsEqual(firstKeyDesc.PubKey) {
					t.Fatalf("mismatched keys: expected %x, "+
						"got %x",
						keyDesc.PubKey.SerializeCompressed(),
						firstKeyDesc.PubKey.SerializeCompressed())
				}
				assertEqualKeyLocator(t,
					KeyLocator{
						Family: keyFam,
						Index:  0,
					}, firstKeyDesc.KeyLocator,
				)

				// If we now try to manually derive the next 10
				// keys (including the original key), then we
				// should get an identical public key back and
				// their KeyLocator information
				// should be set properly.
				for i := 0; i < numKeysToDerive+1; i++ {
					keyLoc := KeyLocator{
						Family: keyFam,
						Index:  uint32(i),
					}
					keyDesc, err := keyRing.DeriveKey(keyLoc)
					if err != nil {
						t.Fatalf("unable to derive first key for "+
							"keyFam=%v: %v", keyFam, err)
					}

					// Ensure that the key locator matches
					// up as well.
					assertEqualKeyLocator(
						t, keyLoc, keyDesc.KeyLocator,
					)
				}

				// If this succeeds, then we'll also try to
				// derive a random index within the range.
				randKeyIndex := uint32(rand.Int31())
				keyLoc = KeyLocator{
					Family: keyFam,
					Index:  randKeyIndex,
				}
				keyDesc, err = keyRing.DeriveKey(keyLoc)
				if err != nil {
					t.Fatalf("unable to derive key_index=%v "+
						"for keyFam=%v: %v",
						randKeyIndex, keyFam, err)
				}
				assertEqualKeyLocator(
					t, keyLoc, keyDesc.KeyLocator,
				)
			}
		})
		if !success {
			break
		}
	}
}

// secretKeyRingConstructor is a function signature that's used as a generic
// constructor for various implementations of the SecretKeyRing interface. A
// string naming the returned interface, a function closure that cleans up any
// resources, and the clean up interface itself are to be returned.
type secretKeyRingConstructor func() (string, func(), SecretKeyRing, error)

// TestSecretKeyRingDerivation tests that each known SecretKeyRing
// implementation properly adheres to the expected behavior of the set of
// interface.
func TestSecretKeyRingDerivation(t *testing.T) {
	t.Parallel()

	secretKeyRingImplementations := []secretKeyRingConstructor{
		func() (string, func(), SecretKeyRing, error) {
			cleanUp, wallet, err := createTestWallet()
			if err != nil {
				t.Fatalf("unable to create wallet: %v", err)
			}

			keyRing := NewWalletKeyRing(wallet)

			return "dcrwallet", cleanUp, keyRing, nil
		},
		func() (string, func(), SecretKeyRing, error) {
			return "hdkeyring", func() {}, createTestHDKeyRing(), nil
		},
	}

	// For each implementation constructor registered above, we'll execute
	// an identical set of tests in order to ensure that the interface
	// adheres to our nominal specification.
	for _, secretKeyRingConstructor := range secretKeyRingImplementations {
		keyRingName, cleanUp, secretKeyRing, err := secretKeyRingConstructor()
		if err != nil {
			t.Fatalf("unable to create secret key ring %v: %v",
				keyRingName, err)
		}
		defer cleanUp()

		success := t.Run(fmt.Sprintf("%v", keyRingName), func(t *testing.T) {
			// For, each key family, we'll ensure that we're able
			// to obtain the private key of a randomly select child
			// index within the key family.
			for _, keyFam := range versionZeroKeyFamilies {
				randKeyIndex := uint32(rand.Int31())
				keyLoc := KeyLocator{
					Family: keyFam,
					Index:  randKeyIndex,
				}

				// First, we'll query for the public key for
				// this target key locator.
				pubKeyDesc, err := secretKeyRing.DeriveKey(keyLoc)
				if err != nil {
					t.Fatalf("unable to derive pubkey "+
						"(fam=%v, index=%v): %v",
						keyLoc.Family,
						keyLoc.Index, err)
				}

				// With the public key derive, ensure that
				// we're able to obtain the corresponding
				// private key correctly.
				privKey, err := secretKeyRing.DerivePrivKey(KeyDescriptor{
					KeyLocator: keyLoc,
				})
				if err != nil {
					t.Fatalf("unable to derive priv "+
						"(fam=%v, index=%v): %v", keyLoc.Family,
						keyLoc.Index, err)
				}

				// Finally, ensure that the keys match up
				// properly.
				if !pubKeyDesc.PubKey.IsEqual(privKey.PubKey()) {
					t.Fatalf("pubkeys mismatched: expected %x, got %x",
						pubKeyDesc.PubKey.SerializeCompressed(),
						privKey.PubKey().SerializeCompressed())
				}

				// Next, we'll test that we're able to derive a
				// key given only the public key and key
				// family.
				//
				// Derive a new key from the key ring.
				keyDesc, err := secretKeyRing.DeriveNextKey(keyFam)
				if err != nil {
					t.Fatalf("unable to derive key: %v", err)
				}

				// We'll now construct a key descriptor that
				// requires us to scan the key range, and query
				// for the key, we should be able to find it as
				// it's valid.
				keyDesc = KeyDescriptor{
					PubKey: keyDesc.PubKey,
					KeyLocator: KeyLocator{
						Family: keyFam,
					},
				}
				privKey, err = secretKeyRing.DerivePrivKey(keyDesc)
				if err != nil {
					t.Fatalf("unable to derive priv key "+
						"via scanning: %v", err)
				}

				// Having to resort to scanning, we should be
				// able to find the target public key.
				if !keyDesc.PubKey.IsEqual(privKey.PubKey()) {
					t.Fatalf("pubkeys mismatched: expected %x, got %x",
						pubKeyDesc.PubKey.SerializeCompressed(),
						privKey.PubKey().SerializeCompressed())
				}

				// We'll try again, but this time with an
				// unknown public key.
				_, pub := secp256k1.PrivKeyFromBytes(testHDSeed[:])
				keyDesc.PubKey = pub

				// If we attempt to query for this key, then we
				// should get ErrCannotDerivePrivKey.
				_, err = secretKeyRing.DerivePrivKey(
					keyDesc,
				)
				if err != ErrCannotDerivePrivKey {
					t.Fatalf("expected %T, instead got %v",
						ErrCannotDerivePrivKey, err)
				}

				// TODO(roasbeef): scalar mult once integrated
			}
		})
		if !success {
			break
		}
	}
}

func init() {
	// We'll clamp the max range scan to constrain the run time of the
	// private key scan test.
	MaxKeyRangeScan = 3
}