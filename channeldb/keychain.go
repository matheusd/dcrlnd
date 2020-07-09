package channeldb

import (
	"bytes"
	"errors"

	"github.com/decred/dcrlnd/channeldb/kvdb"
)

const (
	// lastUsableKeyFamily is the last key family index that can be stored
	// by the database. This value matches the last account number that can
	// be used to create an account in HD wallets, assuming accounts are
	// created as hardened branches.
	lastUsableKeyFamily = 0x7fffffff

	// lastUsableFamilyIndex is the last index that can be returned by a
	// given key family.
	lastUsableKeyFamilyIndex = 0x7fffffff
)

var (
	// errInvalidKeyFamily is returned when an invalid key family is
	// requested.
	errInvalidKeyFamily = errors.New("invalid key family")

	// errKeyFamilyExchausted is returned when a given keyfamily has
	// generated enough indexes that no more can be generated.
	errKeyFamilyExhausted = errors.New("keyfamily indexes exhausted")

	// errDifferentAccountID is returned when the account ID provided to
	// CompareAndStoreAccountID is not the same as the one stored in the
	// database.
	errDifferentAccountID = errors.New("account ID is different than stored in the database")

	// keychainBucket is the root bucket used to store keychain/keyring
	// data.
	keychainBucket = []byte("keychain")

	// keyFamilyIndexesBucket is the bucket used to store the current index
	// of each requested key famiy.
	//
	// Keys are byte-ordered uint32 slices, and values are byte-ordered
	// uint32 values that represent the last returned index for a family.
	keyFamilyIndexesBucket = []byte("kfidxs")

	// keyAccountIDBucket is the bucket used to store the identifier of the
	// account previously used with this keychain. By convention, this is
	// the first pubkey of the first keyfamily of the keychain/account.
	keyAccountIDBucket = []byte("acctid")
)

// NextFamilyIndex returns the next index for a given family of keys from the
// database-backed keyring.
//
// A _KeyFamily_ is an uint32 that maps to the key families of the keychain
// package, while the returned index can be considered the index of a
// (possibly) unused key.
//
// Repeated calls to NextKeyFamilyIndex will return different values. This
// function errors if the requested family would create an invalid HD extended
// key or if it the key family has been exhausted and no more keys can be
// generated for it.
func (d *DB) NextKeyFamilyIndex(keyFamily uint32) (uint32, error) {
	var index uint32

	// Key families higher than this limit would cause a numeric overflow
	// due to accounts using hardened HD branches.
	if keyFamily > lastUsableKeyFamily {
		return 0, errInvalidKeyFamily
	}

	err := kvdb.Update(d, func(tx kvdb.RwTx) error {
		keychain, err := tx.CreateTopLevelBucket(keychainBucket)
		if err != nil {
			return err
		}

		keyFamilies, err := keychain.CreateBucketIfNotExists(
			keyFamilyIndexesBucket,
		)
		if err != nil {
			return err
		}

		// Attempt to read the existing value for the given family.
		var k [4]byte
		var v [4]byte
		byteOrder.PutUint32(k[:], keyFamily)
		oldv := keyFamilies.Get(k[:])

		// If there is a value, decode it to get the next usable index.
		if len(oldv) == 4 {
			index = byteOrder.Uint32(oldv)

			// If we've passed the usable range for this keyfamily,
			// return an error.
			if index >= lastUsableKeyFamilyIndex {
				return errKeyFamilyExhausted
			}
		}

		// Update the database with the next usable index.
		byteOrder.PutUint32(v[:], index+1)
		keyFamilies.Put(k[:], v[:])

		return nil
	})

	return index, err
}

// CompareAndStoreAccountID attempts to compare an existing account ID to a
// given parameter if the ID exists in the database and returns an error if the
// IDs don't match. If the database is empty, then the ID is stored.
func (d *DB) CompareAndStoreAccountID(id []byte) error {
	return kvdb.Update(d, func(tx kvdb.RwTx) error {
		keychain, err := tx.CreateTopLevelBucket(keychainBucket)
		if err != nil {
			return err
		}

		acctId := keychain.Get(keyAccountIDBucket)
		if acctId == nil {
			keychain.Put(keyAccountIDBucket, id)
			return nil
		}

		if !bytes.Equal(acctId, id) {
			return errDifferentAccountID
		}

		return nil
	})
}
