package remotedcrwallet

import (
	"errors"

	"github.com/decred/dcrd/hdkeychain/v2"
	"github.com/decred/dcrlnd/channeldb"
	"github.com/decred/dcrlnd/keychain"
)

// remoteWalletKeyRing is an implementation of both the KeyRing and
// SecretKeyRing interfaces backed by a root master HD key.
type remoteWalletKeyRing struct {
	*keychain.HDKeyRing

	rootXPriv       *hdkeychain.ExtendedKey
	multiSigKFXpriv *hdkeychain.ExtendedKey

	// db is a pointer to a channeldb.DB instance that stores the indices
	// of used keys of the keyring. This is used to track the next
	// available key to prevent key reuse when establishing channels.
	db *channeldb.DB
}

// Compile time type assertions to ensure remoteWalletKeyRing fulfills the desired
// interfaces.
var _ keychain.KeyRing = (*remoteWalletKeyRing)(nil)
var _ keychain.SecretKeyRing = (*remoteWalletKeyRing)(nil)

// newRemoteWalletKeyRing creates a new implementation of the
// keychain.SecretKeyRing interface backed by the given root extended private
// key.
func newRemoteWalletKeyRing(rootAccountXPriv *hdkeychain.ExtendedKey, db *channeldb.DB) (*remoteWalletKeyRing, error) {

	if !rootAccountXPriv.IsPrivate() {
		return nil, errors.New("Provided key is not an extended private key")
	}

	// By convention, the root LN key is a hardened key derived from index
	// 1017 of a wallet account.
	idx := uint32(hdkeychain.HardenedKeyStart + keychain.BIP0043Purpose)
	rootXPriv, err := rootAccountXPriv.Child(idx)
	if err != nil {
		return nil, err
	}

	// This assumes that there are no discontinuities within the KeyFamily
	// constants.
	lastKeyFam := uint32(keychain.KeyFamilyLastKF)
	masterPubs := make(map[keychain.KeyFamily]*hdkeychain.ExtendedKey,
		lastKeyFam+1)

	// The masterpub for the multisig key family is still the external
	// branch for the root account xpriv provided. This allows the wallet
	// itself to watch for on-chain spends from it and correctly account
	// for its funds.
	multiSigXPriv, err := rootAccountXPriv.Child(0)
	if err != nil {
		return nil, err
	}
	multiSigXPub, err := multiSigXPriv.Neuter()
	if err != nil {
		return nil, err
	}
	masterPubs[keychain.KeyFamilyMultiSig] = multiSigXPub

	// Derive the master pubs for the other key families.
	for i := uint32(0); i <= lastKeyFam; i++ {
		if i == uint32(keychain.KeyFamilyMultiSig) {
			continue
		}

		// Errors here cause fatal failures due to the wallet not
		// attempting to generate the next account.
		famKey, err := rootXPriv.Child(hdkeychain.HardenedKeyStart + i)
		if err != nil {
			return nil, err
		}

		famPub, err := famKey.Neuter()
		if err != nil {
			return nil, err
		}

		masterPubs[keychain.KeyFamily(i)] = famPub
	}

	wkr := &remoteWalletKeyRing{
		rootXPriv:       rootXPriv,
		multiSigKFXpriv: multiSigXPriv,
		db:              db,
	}
	wkr.HDKeyRing = keychain.NewHDKeyRing(masterPubs, wkr.fetchMasterPriv,
		wkr.nextIndex)
	return wkr, nil
}

func (kr *remoteWalletKeyRing) nextIndex(keyFam keychain.KeyFamily) (uint32, error) {
	return kr.db.NextKeyFamilyIndex(uint32(keyFam))
}

func (kr *remoteWalletKeyRing) fetchMasterPriv(keyFam keychain.KeyFamily) (*hdkeychain.ExtendedKey,
	error) {

	if keyFam == keychain.KeyFamilyMultiSig {
		return kr.multiSigKFXpriv, nil
	}

	idx := uint32(hdkeychain.HardenedKeyStart + keyFam)
	famKey, err := kr.rootXPriv.Child(idx)
	if err != nil {
		return nil, err
	}

	return famKey, nil
}
