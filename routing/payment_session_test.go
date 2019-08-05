package routing

import (
	"testing"

	"github.com/decred/dcrlnd/channeldb"
	"github.com/decred/dcrlnd/lnwire"
	"github.com/decred/dcrlnd/routing/route"
)

func TestRequestRoute(t *testing.T) {
	const (
		height = 10
	)

	findPath := func(g *graphParams, r *RestrictParams,
		cfg *PathFindingConfig, source, target route.Vertex,
		amt lnwire.MilliAtom) ([]*channeldb.ChannelEdgePolicy,
		error) {

		// We expect find path to receive a cltv limit excluding the
		// final cltv delta (including the block padding).
		if *r.CltvLimit != 22-uint32(BlockPadding) {
			t.Fatal("wrong cltv limit")
		}

		path := []*channeldb.ChannelEdgePolicy{
			{
				Node: &channeldb.LightningNode{},
			},
		}

		return path, nil
	}

	sessionSource := &SessionSource{
		SelfNode: &channeldb.LightningNode{},
		MissionControl: &MissionControl{
			cfg: &MissionControlConfig{},
		},
	}

	session := &paymentSession{
		sessionSource: sessionSource,
		pathFinder:    findPath,
	}

	cltvLimit := uint32(30)
	finalCltvDelta := uint16(8)

	payment := &LightningPayment{
		CltvLimit:      &cltvLimit,
		FinalCLTVDelta: finalCltvDelta,
	}

	route, err := session.RequestRoute(payment, height, finalCltvDelta)
	if err != nil {
		t.Fatal(err)
	}

	// We expect an absolute route lock value of height + finalCltvDelta
	// + BlockPadding.
	if route.TotalTimeLock != 18+uint32(BlockPadding) {
		t.Fatalf("unexpected total time lock of %v",
			route.TotalTimeLock)
	}
}