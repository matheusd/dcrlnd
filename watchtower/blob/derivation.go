package blob

import (
	"encoding/hex"

	"github.com/decred/dcrd/chaincfg/chainhash"
)

// BreachHintSize is the length of the identifier used to detect remote
// commitment broadcasts.
const BreachHintSize = 16

// BreachHint is the first 16-bytes of chainhash(txid), which is used to
// identify the breach transaction.
type BreachHint [BreachHintSize]byte

// NewBreachHintFromHash creates a breach hint from a transaction ID.
func NewBreachHintFromHash(hash *chainhash.Hash) BreachHint {
	h := chainhash.HashB(hash[:])
	var hint BreachHint
	copy(hint[:], h)
	return hint
}

// String returns a hex encoding of the breach hint.
func (h BreachHint) String() string {
	return hex.EncodeToString(h[:])
}
