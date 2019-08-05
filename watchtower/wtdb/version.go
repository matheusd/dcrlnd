package wtdb

import (
	"github.com/decred/dcrlnd/channeldb"
	bolt "go.etcd.io/bbolt"
)

// migration is a function which takes a prior outdated version of the database
// instances and mutates the key/bucket structure to arrive at a more
// up-to-date version of the database.
type migration func(tx *bolt.Tx) error

// version pairs a version number with the migration that would need to be
// applied from the prior version to upgrade.
type version struct {
	migration migration
}

// towerDBVersions stores all versions and migrations of the tower database.
// This list will be used when opening the database to determine if any
// migrations must be applied.
var towerDBVersions = []version{}

// clientDBVersions stores all versions and migrations of the client database.
// This list will be used when opening the database to determine if any
// migrations must be applied.
var clientDBVersions = []version{}

// getLatestDBVersion returns the last known database version.
func getLatestDBVersion(versions []version) uint32 {
	return uint32(len(versions))
}

// getMigrations returns a slice of all updates with a greater number that
// curVersion that need to be applied to sync up with the latest version.
func getMigrations(versions []version, curVersion uint32) []version {
	var updates []version
	for i, v := range versions {
		if uint32(i)+1 > curVersion {
			updates = append(updates, v)
		}
	}

	return updates
}

// getDBVersion retrieves the current database version from the metadata bucket
// using the dbVersionKey.
func getDBVersion(tx *bolt.Tx) (uint32, error) {
	metadata := tx.Bucket(metadataBkt)
	if metadata == nil {
		return 0, ErrUninitializedDB
	}

	versionBytes := metadata.Get(dbVersionKey)
	if len(versionBytes) != 4 {
		return 0, ErrNoDBVersion
	}

	return byteOrder.Uint32(versionBytes), nil
}

// initDBVersion initializes the top-level metadata bucket and writes the passed
// version number as the current version.
func initDBVersion(tx *bolt.Tx, version uint32) error {
	_, err := tx.CreateBucketIfNotExists(metadataBkt)
	if err != nil {
		return err
	}

	return putDBVersion(tx, version)
}

// putDBVersion stores the passed database version in the metadata bucket under
// the dbVersionKey.
func putDBVersion(tx *bolt.Tx, version uint32) error {
	metadata := tx.Bucket(metadataBkt)
	if metadata == nil {
		return ErrUninitializedDB
	}

	versionBytes := make([]byte, 4)
	byteOrder.PutUint32(versionBytes, version)
	return metadata.Put(dbVersionKey, versionBytes)
}

// versionedDB is a private interface implemented by both the tower and client
// databases, permitting all versioning operations to be performed generically
// on either.
type versionedDB interface {
	// bdb returns the underlying bolt database.
	bdb() *bolt.DB

	// Version returns the current version stored in the database.
	Version() (uint32, error)
}

// initOrSyncVersions ensures that the database version is properly set before
// opening the database up for regular use. When the database is being
// initialized for the first time, the caller should set init to true, which
// will simply write the latest version to the database. Otherwise, passing init
// as false will cause the database to apply any needed migrations to ensure its
// version matches the latest version in the provided versions list.
func initOrSyncVersions(db versionedDB, init bool, versions []version) error {
	// If the database has not yet been created, we'll initialize the
	// database version with the latest known version.
	if init {
		return db.bdb().Update(func(tx *bolt.Tx) error {
			return initDBVersion(tx, getLatestDBVersion(versions))
		})
	}

	// Otherwise, ensure that any migrations are applied to ensure the data
	// is in the format expected by the latest version.
	return syncVersions(db, versions)
}

// syncVersions ensures the database version is consistent with the highest
// known database version, applying any migrations that have not been made. If
// the highest known version number is lower than the database's version, this
// method will fail to prevent accidental reversions.
func syncVersions(db versionedDB, versions []version) error {
	curVersion, err := db.Version()
	if err != nil {
		return err
	}

	latestVersion := getLatestDBVersion(versions)
	switch {

	// Current version is higher than any known version, fail to prevent
	// reversion.
	case curVersion > latestVersion:
		return channeldb.ErrDBReversion

	// Current version matches highest known version, nothing to do.
	case curVersion == latestVersion:
		return nil
	}

	// Otherwise, apply any migrations in order to bring the database
	// version up to the highest known version.
	updates := getMigrations(versions, curVersion)
	return db.bdb().Update(func(tx *bolt.Tx) error {
		for i, update := range updates {
			if update.migration == nil {
				continue
			}

			version := curVersion + uint32(i) + 1
			log.Infof("Applying migration #%d", version)

			err := update.migration(tx)
			if err != nil {
				log.Errorf("Unable to apply migration #%d: %v",
					version, err)
				return err
			}
		}

		return putDBVersion(tx, latestVersion)
	})
}