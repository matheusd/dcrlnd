package channeldb

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/decred/dcrlnd/lntypes"
	bolt "go.etcd.io/bbolt"
)

var (
	// ErrAlreadyPaid signals we have already paid this payment hash.
	ErrAlreadyPaid = errors.New("invoice is already paid")

	// ErrPaymentInFlight signals that payment for this payment hash is
	// already "in flight" on the network.
	ErrPaymentInFlight = errors.New("payment is in transition")

	// ErrPaymentNotInitiated is returned  if payment wasn't initiated in
	// switch.
	ErrPaymentNotInitiated = errors.New("payment isn't initiated")

	// ErrPaymentAlreadyCompleted is returned in the event we attempt to
	// recomplete a completed payment.
	ErrPaymentAlreadyCompleted = errors.New("payment is already completed")

	// ErrPaymentAlreadyFailed is returned in the event we attempt to
	// re-fail a failed payment.
	ErrPaymentAlreadyFailed = errors.New("payment has already failed")

	// ErrUnknownPaymentStatus is returned when we do not recognize the
	// existing state of a payment.
	ErrUnknownPaymentStatus = errors.New("unknown payment status")
)

// ControlTower tracks all outgoing payments made, whose primary purpose is to
// prevent duplicate payments to the same payment hash. In production, a
// persistent implementation is preferred so that tracking can survive across
// restarts. Payments are transitioned through various payment states, and the
// ControlTower interface provides access to driving the state transitions.
type ControlTower interface {
	// InitPayment atomically moves the payment into the InFlight state.
	// This method checks that no completed payment exist for this payment
	// hash.
	InitPayment(lntypes.Hash, *PaymentCreationInfo) error

	// RegisterAttempt atomically records the provided PaymentAttemptInfo.
	RegisterAttempt(lntypes.Hash, *PaymentAttemptInfo) error

	// Success transitions a payment into the Completed state. After
	// invoking this method, InitPayment should always return an error to
	// prevent us from making duplicate payments to the same payment hash.
	// The provided preimage is atomically saved to the DB for record
	// keeping.
	Success(lntypes.Hash, lntypes.Preimage) error

	// Fail transitions a payment into the Failed state. After invoking
	// this method, InitPayment should return nil on its next call for this
	// payment hash, allowing the switch to make a subsequent payment.
	Fail(lntypes.Hash) error
}

// paymentControl is persistent implementation of ControlTower to restrict
// double payment sending.
type paymentControl struct {
	db *DB
}

// NewPaymentControl creates a new instance of the paymentControl.
func NewPaymentControl(db *DB) ControlTower {
	return &paymentControl{
		db: db,
	}
}

// InitPayment checks or records the given PaymentCreationInfo with the DB,
// making sure it does not already exist as an in-flight payment. Then this
// method returns successfully, the payment is guranteeed to be in the InFlight
// state.
func (p *paymentControl) InitPayment(paymentHash lntypes.Hash,
	info *PaymentCreationInfo) error {

	var b bytes.Buffer
	if err := serializePaymentCreationInfo(&b, info); err != nil {
		return err
	}
	infoBytes := b.Bytes()

	var takeoffErr error
	err := p.db.Batch(func(tx *bolt.Tx) error {
		bucket, err := fetchPaymentBucket(tx, paymentHash)
		if err != nil {
			return err
		}

		// Get the existing status of this payment, if any.
		paymentStatus := fetchPaymentStatus(bucket)

		// Reset the takeoff error, to avoid carrying over an error
		// from a previous execution of the batched db transaction.
		takeoffErr = nil

		switch paymentStatus {

		// We allow retrying failed payments.
		case StatusFailed:

		// This is a new payment that is being initialized for the
		// first time.
		case StatusGrounded:

		// We already have an InFlight payment on the network. We will
		// disallow any new payments.
		case StatusInFlight:
			takeoffErr = ErrPaymentInFlight
			return nil

		// We've already completed a payment to this payment hash,
		// forbid the switch from sending another.
		case StatusCompleted:
			takeoffErr = ErrAlreadyPaid
			return nil

		default:
			takeoffErr = ErrUnknownPaymentStatus
			return nil
		}

		// Obtain a new sequence number for this payment. This is used
		// to sort the payments in order of creation, and also acts as
		// a unique identifier for each payment.
		sequenceNum, err := nextPaymentSequence(tx)
		if err != nil {
			return err
		}

		err = bucket.Put(paymentSequenceKey, sequenceNum)
		if err != nil {
			return err
		}

		// We'll move it into the InFlight state.
		err = bucket.Put(paymentStatusKey, StatusInFlight.Bytes())
		if err != nil {
			return err
		}

		// Add the payment info to the bucket, which contains the
		// static information for this payment
		err = bucket.Put(paymentCreationInfoKey, infoBytes)
		if err != nil {
			return err
		}

		// We'll delete any lingering attempt info to start with, in
		// case we are initializing a payment that was attempted
		// earlier, but left in a state where we could retry.
		err = bucket.Delete(paymentAttemptInfoKey)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil
	}

	return takeoffErr
}

// RegisterAttempt atomically records the provided PaymentAttemptInfo to the
// DB.
func (p *paymentControl) RegisterAttempt(paymentHash lntypes.Hash,
	attempt *PaymentAttemptInfo) error {

	// Serialize the information before opening the db transaction.
	var a bytes.Buffer
	if err := serializePaymentAttemptInfo(&a, attempt); err != nil {
		return err
	}
	attemptBytes := a.Bytes()

	var updateErr error
	err := p.db.Batch(func(tx *bolt.Tx) error {
		// Reset the update error, to avoid carrying over an error
		// from a previous execution of the batched db transaction.
		updateErr = nil

		bucket, err := fetchPaymentBucket(tx, paymentHash)
		if err != nil {
			return err
		}

		// We can only register attempts for payments that are
		// in-flight.
		if err := ensureInFlight(bucket); err != nil {
			updateErr = err
			return nil
		}

		// Add the payment attempt to the payments bucket.
		return bucket.Put(paymentAttemptInfoKey, attemptBytes)
	})
	if err != nil {
		return err
	}

	return updateErr
}

// Success transitions a payment into the Completed state. After invoking this
// method, InitPayment should always return an error to prevent us from making
// duplicate payments to the same payment hash. The provided preimage is
// atomically saved to the DB for record keeping.
func (p *paymentControl) Success(paymentHash lntypes.Hash,
	preimage lntypes.Preimage) error {

	var updateErr error
	err := p.db.Batch(func(tx *bolt.Tx) error {
		// Reset the update error, to avoid carrying over an error from
		// a previous execution of the batched db transaction.
		updateErr = nil

		bucket, err := fetchPaymentBucket(tx, paymentHash)
		if err != nil {
			return err
		}

		// We can only mark in-flight payments as succeeded.
		if err := ensureInFlight(bucket); err != nil {
			updateErr = err
			return nil
		}

		// Record the successful payment info atomically to the
		// payments record.
		err = bucket.Put(paymentSettleInfoKey, preimage[:])
		if err != nil {
			return err
		}

		return bucket.Put(paymentStatusKey, StatusCompleted.Bytes())
	})
	if err != nil {
		return err
	}

	return updateErr

}

// Fail transitions a payment into the Failed state. After invoking this
// method, InitPayment should return nil on its next call for this payment
// hash, allowing the switch to make a subsequent payment.
func (p *paymentControl) Fail(paymentHash lntypes.Hash) error {
	var updateErr error
	err := p.db.Batch(func(tx *bolt.Tx) error {
		// Reset the update error, to avoid carrying over an error from
		// a previous execution of the batched db transaction.
		updateErr = nil

		bucket, err := fetchPaymentBucket(tx, paymentHash)
		if err != nil {
			return err
		}

		// We can only mark in-flight payments as failed.
		if err := ensureInFlight(bucket); err != nil {
			updateErr = err
			return nil
		}

		// A failed response was received for an InFlight payment, mark
		// it as Failed to allow subsequent attempts.
		return bucket.Put(paymentStatusKey, StatusFailed.Bytes())
	})
	if err != nil {
		return err
	}

	return updateErr
}

// fetchPaymentBucket fetches or creates the sub-bucket assigned to this
// payment hash.
func fetchPaymentBucket(tx *bolt.Tx, paymentHash lntypes.Hash) (
	*bolt.Bucket, error) {

	payments, err := tx.CreateBucketIfNotExists(paymentsRootBucket)
	if err != nil {
		return nil, err
	}

	return payments.CreateBucketIfNotExists(paymentHash[:])
}

// nextPaymentSequence returns the next sequence number to store for a new
// payment.
func nextPaymentSequence(tx *bolt.Tx) ([]byte, error) {
	payments, err := tx.CreateBucketIfNotExists(paymentsRootBucket)
	if err != nil {
		return nil, err
	}

	seq, err := payments.NextSequence()
	if err != nil {
		return nil, err
	}

	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, seq)
	return b, nil
}

// fetchPaymentStatus fetches the payment status from the bucket.  If the
// status isn't found, it will default to "StatusGrounded".
func fetchPaymentStatus(bucket *bolt.Bucket) PaymentStatus {
	// The default status for all payments that aren't recorded in
	// database.
	var paymentStatus = StatusGrounded

	paymentStatusBytes := bucket.Get(paymentStatusKey)
	if paymentStatusBytes != nil {
		paymentStatus.FromBytes(paymentStatusBytes)
	}

	return paymentStatus
}

// ensureInFlight checks whether the payment found in the given bucket has
// status InFlight, and returns an error otherwise. This should be used to
// ensure we only mark in-flight payments as succeeded or failed.
func ensureInFlight(bucket *bolt.Bucket) error {
	paymentStatus := fetchPaymentStatus(bucket)

	switch {

	// The payment was indeed InFlight, return.
	case paymentStatus == StatusInFlight:
		return nil

	// Our records show the payment as unknown, meaning it never
	// should have left the switch.
	case paymentStatus == StatusGrounded:
		return ErrPaymentNotInitiated

	// The payment succeeded previously.
	case paymentStatus == StatusCompleted:
		return ErrPaymentAlreadyCompleted

	// The payment was already failed.
	case paymentStatus == StatusFailed:
		return ErrPaymentAlreadyFailed

	default:
		return ErrUnknownPaymentStatus
	}
}