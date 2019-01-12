package watchtower

import (
	"github.com/decred/dcrlnd/watchtower/lookout"
	"github.com/decred/dcrlnd/watchtower/wtserver"
)

// DB abstracts the persistent functionality required to run the watchtower
// daemon. It composes the database interfaces required by the lookout and
// wtserver subsystems.
type DB interface {
	lookout.DB
	wtserver.DB
}
