package contractcourt

import (
	"github.com/decred/dcrlnd/channeldb"
	"github.com/decred/dcrlnd/invoices"
	"github.com/decred/dcrlnd/lntypes"
	"github.com/decred/dcrlnd/lnwire"
)

type notifyExitHopData struct {
	payHash       lntypes.Hash
	paidAmount    lnwire.MilliAtom
	hodlChan      chan<- interface{}
	expiry        uint32
	currentHeight int32
}

type mockRegistry struct {
	notifyChan  chan notifyExitHopData
	notifyErr   error
	notifyEvent *invoices.HodlEvent
}

func (r *mockRegistry) NotifyExitHopHtlc(payHash lntypes.Hash,
	paidAmount lnwire.MilliAtom, expiry uint32, currentHeight int32,
	circuitKey channeldb.CircuitKey, hodlChan chan<- interface{},
	eob []byte) (*invoices.HodlEvent, error) {

	r.notifyChan <- notifyExitHopData{
		hodlChan:      hodlChan,
		payHash:       payHash,
		paidAmount:    paidAmount,
		expiry:        expiry,
		currentHeight: currentHeight,
	}

	return r.notifyEvent, r.notifyErr
}

func (r *mockRegistry) HodlUnsubscribeAll(subscriber chan<- interface{}) {}

func (r *mockRegistry) LookupInvoice(lntypes.Hash) (channeldb.Invoice, uint32,
	error) {

	return channeldb.Invoice{}, 0, channeldb.ErrInvoiceNotFound
}
