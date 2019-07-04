package main

import "github.com/decred/dcrlnd/lnrpc/walletrpc"

// PendingSweep is a CLI-friendly type of the walletrpc.PendingSweep proto. We
// use this to show more useful string versions of byte slices and enums.
type PendingSweep struct {
	OutPoint            OutPoint `json:"outpoint"`
	WitnessType         string   `json:"witness_type"`
	AmountAtoms         uint32   `json:"amount_atoms"`
	AtomsPerByte        uint32   `json:"atoms_per_byte"`
	BroadcastAttempts   uint32   `json:"broadcast_attempts"`
	NextBroadcastHeight uint32   `json:"next_broadcast_height"`
}

// NewPendingSweepFromProto converts the walletrpc.PendingSweep proto type into
// its corresponding CLI-friendly type.
func NewPendingSweepFromProto(pendingSweep *walletrpc.PendingSweep) *PendingSweep {
	return &PendingSweep{
		OutPoint:            NewOutPointFromProto(pendingSweep.Outpoint),
		WitnessType:         pendingSweep.WitnessType.String(),
		AmountAtoms:         pendingSweep.AmountAtoms,
		AtomsPerByte:        pendingSweep.AtomsPerByte,
		BroadcastAttempts:   pendingSweep.BroadcastAttempts,
		NextBroadcastHeight: pendingSweep.NextBroadcastHeight,
	}
}
