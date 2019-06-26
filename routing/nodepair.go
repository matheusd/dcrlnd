package routing

import (
	"github.com/decred/dcrlnd/routing/route"
)

// DirectedNodePair stores a directed pair of nodes.
type DirectedNodePair struct {
	From, To route.Vertex
}
