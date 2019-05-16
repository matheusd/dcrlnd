package routing

import (
	"crypto/sha256"

	"github.com/decred/dcrlnd/htlcswitch"
	"github.com/decred/dcrlnd/lnwire"
)

type mockPaymentAttemptDispatcher struct {
	onPayment func(firstHop lnwire.ShortChannelID) ([32]byte, error)
}

var _ PaymentAttemptDispatcher = (*mockPaymentAttemptDispatcher)(nil)

func (m *mockPaymentAttemptDispatcher) SendHTLC(firstHop lnwire.ShortChannelID,
	_ uint64,
	_ *lnwire.UpdateAddHTLC,
	_ htlcswitch.ErrorDecrypter) ([sha256.Size]byte, error) {

	if m.onPayment != nil {
		return m.onPayment(firstHop)
	}

	return [sha256.Size]byte{}, nil
}

func (m *mockPaymentAttemptDispatcher) setPaymentResult(
	f func(firstHop lnwire.ShortChannelID) ([32]byte, error)) {

	m.onPayment = f
}
