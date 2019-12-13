package kvdb

import (
	_ "github.com/decred/dcrlnd/channeldb/internal/walletdb/bdb" // Import to register backend.
)

// BoltBackendName is the name of the backend that should be passed into
// kvdb.Create to initialize a new instance of kvdb.Backend backed by a live
// instance of bolt.
const BoltBackendName = "bdb"
