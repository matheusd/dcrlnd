package sweep

import (
	"github.com/decred/dcrd/wire"
)

// Wallet contains all wallet related functionality required by sweeper.
type Wallet interface {
	// PublishTransaction performs cursory validation (dust checks, etc)
	// and broadcasts the passed transaction to the Decred network.
	PublishTransaction(tx *wire.MsgTx) error
}
