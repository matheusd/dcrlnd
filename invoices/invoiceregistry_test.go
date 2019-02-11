package invoices

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrlnd/channeldb"
	"github.com/decred/dcrlnd/lntypes"
	"github.com/decred/dcrlnd/lnwire"
	"github.com/decred/dcrlnd/zpay32"
)

var (
	testTimeout = 5 * time.Second

	preimage = lntypes.Preimage{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	}

	hash = preimage.Hash()

	// testPayReq is a dummy payment request that does parse properly. It
	// has no relation with the real invoice parameters and isn't asserted
	// on in this test. LookupInvoice requires this to have a valid value.
	testPayReq = "lndcr1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdq5xysxxatsyp3k7enxv4jsmvp0ygkvzd3zh9wkfj59cuze0se5fzuh4f7rysdukv68n6fafa45sudrzg8d33paaw50zczd5mzmppqaalvzneu0yd3zfrvzhnfzpkgppyrza2"
)

func decodeExpiry(payReq string) (uint32, error) {
	invoice, err := zpay32.Decode(payReq, &chaincfg.MainNetParams)
	if err != nil {
		return 0, err
	}
	return uint32(invoice.MinFinalCLTVExpiry()), nil
}

var (
	testInvoice = &channeldb.Invoice{
		Terms: channeldb.ContractTerm{
			PaymentPreimage: preimage,
			Value:           lnwire.MilliAtom(100000),
		},
		PaymentRequest: []byte(testPayReq),
	}
)

func newTestContext(t *testing.T) (*InvoiceRegistry, func()) {
	cdb, cleanup, err := newDB()
	if err != nil {
		t.Fatal(err)
	}

	// Instantiate and start the invoice registry.
	registry := NewRegistry(cdb, decodeExpiry)

	err = registry.Start()
	if err != nil {
		cleanup()
		t.Fatal(err)
	}

	return registry, func() {
		registry.Stop()
		cleanup()
	}
}

// TestSettleInvoice tests settling of an invoice and related notifications.
func TestSettleInvoice(t *testing.T) {
	registry, cleanup := newTestContext(t)
	defer cleanup()

	allSubscriptions := registry.SubscribeNotifications(0, 0)
	defer allSubscriptions.Cancel()

	// Subscribe to the not yet existing invoice.
	subscription := registry.SubscribeSingleInvoice(hash)
	defer subscription.Cancel()

	if subscription.hash != hash {
		t.Fatalf("expected subscription for provided hash")
	}

	// Give enough time for the invoice ntnf goroutine to complete
	time.Sleep(time.Millisecond * 5)

	// Add the invoice.
	addIdx, err := registry.AddInvoice(testInvoice, hash)
	if err != nil {
		t.Fatal(err)
	}

	if addIdx != 1 {
		t.Fatalf("expected addIndex to start with 1, but got %v",
			addIdx)
	}

	// We expect the open state to be sent to the single invoice subscriber.
	select {
	case update := <-subscription.Updates:
		if update.Terms.State != channeldb.ContractOpen {
			t.Fatalf("expected state ContractOpen, but got %v",
				update.Terms.State)
		}
	case <-time.After(testTimeout):
		t.Fatal("no update received")
	}

	// We expect a new invoice notification to be sent out.
	select {
	case newInvoice := <-allSubscriptions.NewInvoices:
		if newInvoice.Terms.State != channeldb.ContractOpen {
			t.Fatalf("expected state ContractOpen, but got %v",
				newInvoice.Terms.State)
		}
	case <-time.After(testTimeout):
		t.Fatal("no update received")
	}

	hodlChan := make(chan interface{}, 1)

	// Settle invoice with a slightly higher amount.
	amtPaid := lnwire.MilliAtom(100500)
	_, err = registry.NotifyExitHopHtlc(hash, amtPaid, hodlChan)
	if err != nil {
		t.Fatal(err)
	}

	// We expect the settled state to be sent to the single invoice
	// subscriber.
	select {
	case update := <-subscription.Updates:
		if update.Terms.State != channeldb.ContractSettled {
			t.Fatalf("expected state ContractOpen, but got %v",
				update.Terms.State)
		}
		if update.AmtPaid != amtPaid {
			t.Fatal("invoice AmtPaid incorrect")
		}
	case <-time.After(testTimeout):
		t.Fatal("no update received")
	}

	// We expect a settled notification to be sent out.
	select {
	case settledInvoice := <-allSubscriptions.SettledInvoices:
		if settledInvoice.Terms.State != channeldb.ContractSettled {
			t.Fatalf("expected state ContractOpen, but got %v",
				settledInvoice.Terms.State)
		}
	case <-time.After(testTimeout):
		t.Fatal("no update received")
	}

	// Try to settle again.
	_, err = registry.NotifyExitHopHtlc(hash, amtPaid, hodlChan)
	if err != nil {
		t.Fatal("expected duplicate settle to succeed")
	}

	// Try to settle again with a different amount.
	_, err = registry.NotifyExitHopHtlc(hash, amtPaid+600, hodlChan)
	if err != nil {
		t.Fatal("expected duplicate settle to succeed")
	}

	// Check that settled amount remains unchanged.
	inv, _, err := registry.LookupInvoice(hash)
	if err != nil {
		t.Fatal(err)
	}
	if inv.AmtPaid != amtPaid {
		t.Fatal("expected amount to be unchanged")
	}

	// Try to cancel.
	err = registry.CancelInvoice(hash)
	if err != channeldb.ErrInvoiceAlreadySettled {
		t.Fatal("expected cancelation of a settled invoice to fail")
	}

	// As this is a direct sette, we expect nothing on the hodl chan.
	select {
	case <-hodlChan:
		t.Fatal("unexpected event")
	default:
	}
}

// TestCancelInvoice tests cancelation of an invoice and related notifications.
func TestCancelInvoice(t *testing.T) {
	registry, cleanup := newTestContext(t)
	defer cleanup()

	allSubscriptions := registry.SubscribeNotifications(0, 0)
	defer allSubscriptions.Cancel()

	// Try to cancel the not yet existing invoice. This should fail.
	err := registry.CancelInvoice(hash)
	if err != channeldb.ErrInvoiceNotFound {
		t.Fatalf("expected ErrInvoiceNotFound, but got %v", err)
	}

	// Subscribe to the not yet existing invoice.
	subscription := registry.SubscribeSingleInvoice(hash)
	defer subscription.Cancel()

	if subscription.hash != hash {
		t.Fatalf("expected subscription for provided hash")
	}

	// Give enough time for the invoice ntnf goroutine to complete
	time.Sleep(time.Millisecond * 5)

	// Add the invoice.
	amt := lnwire.MilliAtom(100000)
	_, err = registry.AddInvoice(testInvoice, hash)
	if err != nil {
		t.Fatal(err)
	}

	// We expect the open state to be sent to the single invoice subscriber.
	select {
	case update := <-subscription.Updates:
		if update.Terms.State != channeldb.ContractOpen {
			t.Fatalf(
				"expected state ContractOpen, but got %v",
				update.Terms.State,
			)
		}
	case <-time.After(testTimeout):
		t.Fatal("no update received")
	}

	// We expect a new invoice notification to be sent out.
	select {
	case newInvoice := <-allSubscriptions.NewInvoices:
		if newInvoice.Terms.State != channeldb.ContractOpen {
			t.Fatalf(
				"expected state ContractOpen, but got %v",
				newInvoice.Terms.State,
			)
		}
	case <-time.After(testTimeout):
		t.Fatal("no update received")
	}

	// Cancel invoice.
	err = registry.CancelInvoice(hash)
	if err != nil {
		t.Fatal(err)
	}

	// We expect the canceled state to be sent to the single invoice
	// subscriber.
	select {
	case update := <-subscription.Updates:
		if update.Terms.State != channeldb.ContractCanceled {
			t.Fatalf(
				"expected state ContractCanceled, but got %v",
				update.Terms.State,
			)
		}
	case <-time.After(testTimeout):
		t.Fatal("no update received")
	}

	// We expect no cancel notification to be sent to all invoice
	// subscribers (backwards compatibility).

	// Try to cancel again.
	err = registry.CancelInvoice(hash)
	if err != nil {
		t.Fatal("expected cancelation of a canceled invoice to succeed")
	}

	// Notify arrival of a new htlc paying to this invoice. This should
	// succeed.
	hodlChan := make(chan interface{})
	event, err := registry.NotifyExitHopHtlc(hash, amt, hodlChan)
	if err != nil {
		t.Fatal("expected settlement of a canceled invoice to succeed")
	}

	if event.Preimage != nil {
		t.Fatal("expected cancel hodl event")
	}
}

// TestHoldInvoice tests settling of a hold invoice and related notifications.
func TestHoldInvoice(t *testing.T) {
	defer timeout(t)()

	cdb, cleanup, err := newDB()
	if err != nil {
		t.Fatalf("unable to create db: %v", err)
	}
	defer cleanup()

	// Instantiate and start the invoice registry.
	registry := NewRegistry(cdb, decodeExpiry)

	err = registry.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer registry.Stop()

	allSubscriptions := registry.SubscribeNotifications(0, 0)
	defer allSubscriptions.Cancel()

	// Subscribe to the not yet existing invoice.
	subscription := registry.SubscribeSingleInvoice(hash)
	defer subscription.Cancel()

	if subscription.hash != hash {
		t.Fatalf("expected subscription for provided hash")
	}

	// Give some time for the subscription to be processed.
	time.Sleep(time.Millisecond * 5)

	// Add the invoice.
	invoice := &channeldb.Invoice{
		Terms: channeldb.ContractTerm{
			PaymentPreimage: channeldb.UnknownPreimage,
			Value:           lnwire.MilliAtom(100000),
		},
	}

	_, err = registry.AddInvoice(invoice, hash)
	if err != nil {
		t.Fatal(err)
	}

	// We expect the open state to be sent to the single invoice subscriber.
	update := <-subscription.Updates
	if update.Terms.State != channeldb.ContractOpen {
		t.Fatalf("expected state ContractOpen, but got %v",
			update.Terms.State)
	}

	// We expect a new invoice notification to be sent out.
	newInvoice := <-allSubscriptions.NewInvoices
	if newInvoice.Terms.State != channeldb.ContractOpen {
		t.Fatalf("expected state ContractOpen, but got %v",
			newInvoice.Terms.State)
	}

	// Use slightly higher amount for accept/settle.
	amtPaid := lnwire.MilliAtom(100500)

	hodlChan := make(chan interface{}, 1)

	// NotifyExitHopHtlc without a preimage present in the invoice registry
	// should be possible.
	event, err := registry.NotifyExitHopHtlc(hash, amtPaid, hodlChan)
	if err != nil {
		t.Fatalf("expected settle to succeed but got %v", err)
	}
	if event != nil {
		t.Fatalf("unexpect direct settle")
	}

	// Test idempotency.
	event, err = registry.NotifyExitHopHtlc(hash, amtPaid, hodlChan)
	if err != nil {
		t.Fatalf("expected settle to succeed but got %v", err)
	}
	if event != nil {
		t.Fatalf("unexpect direct settle")
	}

	// We expect the accepted state to be sent to the single invoice
	// subscriber. For all invoice subscribers, we don't expect an update.
	// Those only get notified on settle.
	update = <-subscription.Updates
	if update.Terms.State != channeldb.ContractAccepted {
		t.Fatalf("expected state ContractAccepted, but got %v",
			update.Terms.State)
	}
	if update.AmtPaid != amtPaid {
		t.Fatal("invoice AmtPaid incorrect")
	}

	// Settling with preimage should succeed.
	err = registry.SettleHodlInvoice(preimage)
	if err != nil {
		t.Fatal("expected set preimage to succeed")
	}

	hodlEvent := (<-hodlChan).(HodlEvent)
	if *hodlEvent.Preimage != preimage {
		t.Fatal("unexpected preimage in hodl event")
	}

	// We expect a settled notification to be sent out for both all and
	// single invoice subscribers.
	settledInvoice := <-allSubscriptions.SettledInvoices
	if settledInvoice.Terms.State != channeldb.ContractSettled {
		t.Fatalf("expected state ContractSettled, but got %v",
			settledInvoice.Terms.State)
	}

	update = <-subscription.Updates
	if update.Terms.State != channeldb.ContractSettled {
		t.Fatalf("expected state ContractSettled, but got %v",
			update.Terms.State)
	}

	// Idempotency.
	err = registry.SettleHodlInvoice(preimage)
	if err != channeldb.ErrInvoiceAlreadySettled {
		t.Fatalf("expected ErrInvoiceAlreadySettled but got %v", err)
	}

	// Try to cancel.
	err = registry.CancelInvoice(hash)
	if err == nil {
		t.Fatal("expected cancelation of a settled invoice to fail")
	}
}

func newDB() (*channeldb.DB, func(), error) {
	// First, create a temporary directory to be used for the duration of
	// this test.
	tempDirName, err := ioutil.TempDir("", "channeldb")
	if err != nil {
		return nil, nil, err
	}

	// Next, create channeldb for the first time.
	cdb, err := channeldb.Open(tempDirName)
	if err != nil {
		os.RemoveAll(tempDirName)
		return nil, nil, err
	}

	cleanUp := func() {
		cdb.Close()
		os.RemoveAll(tempDirName)
	}

	return cdb, cleanUp, nil
}
