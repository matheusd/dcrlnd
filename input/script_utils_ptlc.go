package input

import (
	"fmt"

	"github.com/decred/dcrd/dcrec"
	"github.com/decred/dcrd/dcrec/secp256k1/v3"
	"github.com/decred/dcrd/dcrutil/v3"
	"github.com/decred/dcrd/txscript/v3"
	"github.com/decred/dcrd/wire"
	"github.com/matheusd/dcr_adaptor_sigs"
)

// SenderHTLCScript constructs the public key script for an outgoing PTLC
// output payment for the sender's version of the commitment transaction. The
// possible script paths from this output include:
//
//    * The sender timing out the PTLC using the second level PTLC timeout
//      transaction.
//    * The receiver of the PTLC claiming the output on-chain with a sig that
//      reveals the PTLC secret scalar (to anyone also holding a corresponding
//	adaptor sig).
//    * The receiver of the PTLC sweeping all the funds in the case that a
//      revoked commitment transaction bearing this PTLC was broadcast.
//
// If confirmedSpend=true, a 1 OP_CSV check will be added to the non-revocation
// cases, to allow sweeping only after confirmation.
//
// TODO: Improve the script and protocol so that redundant keys aren't needed
// (this needs to correctly integrate muSig steps higher up the LN protocol
// stack).
//
// Possible Input Scripts:
//    SENDR: <sendr sig> <recvr sig> <0> (spend using PTLC timeout transaction)
//    RECVR: <recvr sig> <ptlc s sig> (spend using PTLC success no delay tx)
//    REVOK: <revoke sig> <revoke key>
//     * receiver revoke
//
// OP_DUP OP_HASH160 <revocation key hash160> OP_EQUAL
// OP_IF
//     OP_CHECKSIG
// OP_ELSE
//     <recv htlc key>
//     OP_SWAP OP_SIZE 33 OP_EQUAL
//     OP_NOTIF
//         OP_DROP 2 OP_SWAP <sender htlc key> 2 OP_CHECKMULTISIG
//     OP_ELSE
//         <ptlc r-sig> OP_SWAP OP_CAT <ptlc key> 2 OP_CHECKSIGALTVERIFY
//         OP_CHECKSIG
//     OP_ENDIF
//     [1 OP_CHECKSEQUENCEVERIFY OP_DROP] <- if allowing confirmed spend only.
// OP_ENDIF
func SenderPTLCScript(senderHtlcKey, receiverHtlcKey,
	revocationKey, ptlcR, ptlcKey *secp256k1.PublicKey,
	confirmedSpend bool) ([]byte, error) {

	fmt.Printf("YYYYYYYYYY (SenderPTLCScript) R: %x\n", ptlcR.SerializeCompressed())
	fmt.Printf("YYYYYYYYYY (SenderPTLCScript) Key: %x\n", ptlcKey.SerializeCompressed())

	builder := txscript.NewScriptBuilder()

	// The opening operations are used to determine if this is the receiver
	// of the PTLC attempting to sweep all the funds due to a contract
	// breach. In this case, they'll place the revocation key at the top of
	// the stack.
	builder.AddOp(txscript.OP_DUP)
	builder.AddOp(txscript.OP_HASH160)
	builder.AddData(dcrutil.Hash160(revocationKey.SerializeCompressed()))
	builder.AddOp(txscript.OP_EQUAL)

	// If the hash matches, then this is the revocation clause. The output
	// can be spent if the check sig operation passes.
	builder.AddOp(txscript.OP_IF)
	builder.AddOp(txscript.OP_CHECKSIG)

	// Otherwise, this may either be the receiver of the PTLC claiming with
	// the payment scalar, or the sender of the PTLC sweeping the output
	// after it has timed out.
	builder.AddOp(txscript.OP_ELSE)

	// We'll do a bit of set up by pushing the receiver's key on the top of
	// the stack. This will be needed later if we decide that this is the
	// sender activating the time out clause with the PTLC timeout
	// transaction.
	builder.AddData(receiverHtlcKey.SerializeCompressed())

	// Atm, the top item of the stack is the receiverKey's so we use a swap
	// to expose what is either the ptlc sig or the 0 flag.
	builder.AddOp(txscript.OP_SWAP)

	// With the top item swapped, check if it's 65 bytes. If so, then this
	// *may* be a signature for the ptlc key.
	builder.AddOp(txscript.OP_SIZE)
	builder.AddInt64(33)
	builder.AddOp(txscript.OP_EQUAL)

	// If it isn't then this might be the sender of the PTLC activating the
	// time out clause.
	builder.AddOp(txscript.OP_NOTIF)

	// We'll drop the flag value off the top of the stack so we can
	// reconstruct the multi-sig script used as an off-chain covenant. If
	// two valid signatures are provided, ten then output will be deemed as
	// spendable.
	builder.AddOp(txscript.OP_DROP)
	builder.AddOp(txscript.OP_2)
	builder.AddOp(txscript.OP_SWAP)
	builder.AddData(senderHtlcKey.SerializeCompressed())
	builder.AddOp(txscript.OP_2)
	builder.AddOp(txscript.OP_CHECKMULTISIG)

	// Otherwise, then the only other case is that this is the receiver of
	// the PTLC sweeping it on-chain with a signature that reveals the
	// secret payment scalar (for holders of an appropriate adaptor sig).
	builder.AddOp(txscript.OP_ELSE)

	// Push the PTLC R point, which is part of the full signature and
	// ordinarily a random point, but is deterministically generated in
	// order to allow the receiver to generate and send the adaptor sig to
	// the sender.
	//
	// The R point ends up concatenated to the S scalar sent in the sig
	// script to form the full signature as required by consensus rules.
	//
	// Since the consensus rules require the R point to be on a positive Y
	// coordinate, strip the first byte in compact serialization.
	builder.AddData(ptlcR.SerializeCompressed()[1:])
	builder.AddOp(txscript.OP_SWAP)
	builder.AddOp(txscript.OP_CAT)

	// We'll push the ptlc public key and the signature algo flag (schnorr)
	// to the stack and verify that the provided sig is valid.
	builder.AddData(ptlcKey.SerializeCompressed())
	builder.AddData([]byte{byte(dcrec.STSchnorrSecp256k1)})
	builder.AddOp(txscript.OP_CHECKSIGALTVERIFY)

	// This checks the receiver's signature so that a third party with
	// knowledge of the payment preimage still cannot steal the output.
	builder.AddOp(txscript.OP_CHECKSIG)

	// Close out the OP_IF statement above.
	builder.AddOp(txscript.OP_ENDIF)

	// Add 1 block CSV delay if a confirmation is required for the
	// non-revocation clauses.
	if confirmedSpend {
		builder.AddOp(txscript.OP_1)
		builder.AddOp(txscript.OP_CHECKSEQUENCEVERIFY)
		builder.AddOp(txscript.OP_DROP)
	}

	// Close out the OP_IF statement at the top of the script.
	builder.AddOp(txscript.OP_ENDIF)

	return builder.Script()
}

// SenderPtlcSpendRedeem spends the senderPtlcScript using the success branch,
// revealing the preimage via the adaptor sig. This exercises the following
// branch of the sender script:
//
//    RECVR: <recvr sig> <ptlc s sig> (spend using PTLC success no delay tx)
func SenderPtlcSpendRedeem(signer Signer, signDesc *SignDescriptor,
	ptlcSuccessNoDelayTx *wire.MsgTx,
	ptlcSignDesc *SignDescriptor,
	paymentScalar *dcr_adaptor_sigs.SecretScalar) (TxWitness, error) {

	sweepSig, err := signer.SignOutputRaw(ptlcSuccessNoDelayTx, signDesc)
	if err != nil {
		return nil, err
	}

	adaptorSigSig, err := signer.SignOutputRaw(ptlcSuccessNoDelayTx, ptlcSignDesc)
	if err != nil {
		return nil, err
	}
	adaptorSig := adaptorSigSig.(*dcr_adaptor_sigs.AdaptorSignature)

	var ptlcFullSigSBytes []byte
	switch {
	case paymentScalar != nil:
		schnorrSig, err := dcr_adaptor_sigs.AssembleFullSig(adaptorSig, paymentScalar)
		if err != nil {
			return nil, err
		}

		// Serialize the full Schnorr sig and strip the R bytes since
		// that is already included in the corresponding redeem script.
		ptlcFullSigSBytes = schnorrSig.Serialize()[32:]
		ptlcFullSigSBytes = append(ptlcFullSigSBytes, byte(txscript.SigHashAll))

	case adaptorSig != nil:
		// Without the payment scalar we serialize the entire adaptor
		// sig, assuming the scalar will be filled later by
		// ReplaceSenderPtlcSpendRedeemScalar.
		ptlcFullSigSBytes = adaptorSig.Serialize()
	}

	// The final witness stack is used the provide the script with the
	// payment pre-image, and also execute the multi-sig clause after the
	// pre-images matches.
	witnessStack := TxWitness(make([][]byte, 3))
	witnessStack[0] = append(sweepSig.Serialize(), byte(signDesc.HashType))
	witnessStack[1] = ptlcFullSigSBytes
	witnessStack[2] = signDesc.WitnessScript

	return witnessStack, nil
}

// ReplaceSenderPtlcSpendRedeemScalar replaces the ptlc sig in a sigScript
// generated by senderPtlcSpendRedeem with a new, valid, full sig derived from
// an adaptor sig.  This will generate an undefined result script if the
// sigScript was *not* generated by the provided function (ie: this must be the
// sigScript of a SignedSuccessNoDelayTx, used to redeem a PTLC).
//
// If the adaptor sig is empty but the paymentScalar is specified, then an
// attempt is made at deserializing the adaptor script from the passed
// sigScript and the final sig is then assembled.
//
// This function returns the new sigScript to replace the older sigScript.
func ReplaceSenderPtlcSpendRedeemScalar(sigScript []byte,
	adaptorSig *dcr_adaptor_sigs.AdaptorSignature,
	paymentScalar *dcr_adaptor_sigs.SecretScalar) ([]byte, error) {

	sigScriptPushes, err := SigScriptToWitnessStack(sigScript)
	if err != nil {
		return nil, err
	}
	switch {
	case adaptorSig == nil && len(sigScriptPushes[1]) != dcr_adaptor_sigs.AdaptorSignatureSerializeLen:
		return nil, fmt.Errorf("adaptor signature is not serialized in " +
			"the existing sigScript")

	case adaptorSig == nil:
		adaptorSig, err = dcr_adaptor_sigs.ParseAdaptorSignature(sigScriptPushes[1])
		if err != nil {
			return nil, fmt.Errorf("invalid adaptor signature "+
				"serialized in sigScript: %v", err)
		}
	}

	ptlcSig, err := dcr_adaptor_sigs.AssembleFullSig(adaptorSig, paymentScalar)
	if err != nil {
		return nil, err
	}

	// Strip the R point from the serialization since that's already
	// included in the pk script.
	ptlcFullSigSBytes := ptlcSig.Serialize()[32:]
	ptlcFullSigSBytes = append(ptlcFullSigSBytes, byte(txscript.SigHashAll))

	return txscript.NewScriptBuilder().
		AddData(sigScriptPushes[0]). // receiver sig
		AddData(ptlcFullSigSBytes).  // ptlc sig
		AddData(sigScriptPushes[2]). // redeem script
		Script()
}

// ReceiverPTLCScript
//
// If confirmedSpend=true, a 1 OP_CSV check will be added to the non-revocation
// cases, to allow sweeping only after confirmation.
//
// TODO: Instead of including the ptlc payment point directly in the script, it
// could be the HASH160 of it which would make the script smaller for the other
// redeeming cases. Using the key directly for the moment to make the api as
// close as possible to the HTLC case for the moment.
//
// The PTLC case also is ideally realized by a muSig instead of the simple sig+
// standard multisig but that's a larger protocol change as well.
//
// Possible Input Scripts:
//    RECVR: <sender sig> <recvr sig> <ptlc full sig> 1 (spend using HTLC success transaction)
//    REVOK: <sig> <key>
//    SENDR: <sig> 0
//
//
// OP_DUP OP_HASH160 <revocation key hash160> OP_EQUAL
// OP_IF
//     OP_CHECKSIG
// OP_ELSE
//     <sendr htlc key>
//     OP_SWAP
//     OP_IF
//         OP_SWAP <ptlc key> 2 OP_CHECKSIGALTVERIFY
//         2 OP_SWAP <recvr htlc key> 2 OP_CHECKMULTISIG
//     OP_ELSE
//         OP_DROP <cltv expiry> OP_CHECKLOCKTIMEVERIFY OP_DROP
//         OP_CHECKSIG
//     OP_ENDIF
//     [1 OP_CHECKSEQUENCEVERIFY OP_DROP] <- if allowing confirmed spend only.
// OP_ENDIF
func ReceiverPTLCScript(cltvExpiry uint32, senderHtlcKey,
	receiverHtlcKey, revocationKey, ptlcKey *secp256k1.PublicKey,
	confirmedSpend bool) ([]byte, error) {

	builder := txscript.NewScriptBuilder()

	// The opening operations are used to determine if this is the sender
	// of the HTLC attempting to sweep all the funds due to a contract
	// breach. In this case, they'll place the revocation key at the top of
	// the stack.
	builder.AddOp(txscript.OP_DUP)
	builder.AddOp(txscript.OP_HASH160)
	builder.AddData(dcrutil.Hash160(revocationKey.SerializeCompressed()))
	builder.AddOp(txscript.OP_EQUAL)

	// If the hash matches, then this is the revocation clause. The output
	// can be spent if the check sig operation passes.
	builder.AddOp(txscript.OP_IF)
	builder.AddOp(txscript.OP_CHECKSIG)

	// Otherwise, this may either be the receiver of the HTLC starting the
	// claiming process via the second level HTLC success transaction and
	// the pre-image, or the sender of the HTLC sweeping the output after
	// it has timed out.
	builder.AddOp(txscript.OP_ELSE)

	// We'll do a bit of set up by pushing the sender's key on the top of
	// the stack. This will be needed later if we decide that this is the
	// receiver transitioning the output to the claim state using their
	// second-level HTLC success transaction.
	builder.AddData(senderHtlcKey.SerializeCompressed())

	// Atm, the top item of the stack is the sender's key so we use a swap
	// to expose the flag that will select which branch to test next.
	builder.AddOp(txscript.OP_SWAP)

	// If the item on the top of the stack is 1 then the PTLC output should
	// be redeemed with a sig that reveals the ptlc secret scalar (to
	// parties that have a corresponding adaptor sig). This indicates that
	// the receiver of the PTLC is attempting to claim the output on-chain
	// by transitioning the state of the PTLC to delay+claim.
	builder.AddOp(txscript.OP_IF)

	// We'll push the ptlc public point and the signature algo flag
	// (schnorr) to the stack and verify that the provided sig is valid.
	builder.AddOp(txscript.OP_SWAP)
	builder.AddData(ptlcKey.SerializeCompressed())
	builder.AddData([]byte{byte(dcrec.STSchnorrSecp256k1)})
	builder.AddOp(txscript.OP_CHECKSIGALTVERIFY)

	// If the sig is for the ptlc point, then we'll also need to satisfy
	// the multi-sig covenant by providing both signatures of the sender
	// and receiver. If the convenient is met, then we'll allow the
	// spending of this output, but only by the PTLC success transaction.
	builder.AddOp(txscript.OP_2)
	builder.AddOp(txscript.OP_SWAP)
	builder.AddData(receiverHtlcKey.SerializeCompressed())
	builder.AddOp(txscript.OP_2)
	builder.AddOp(txscript.OP_CHECKMULTISIG)

	// Otherwise, this might be the sender of the HTLC attempting to sweep
	// it on-chain after the timeout.
	builder.AddOp(txscript.OP_ELSE)

	// We'll drop the extra item (which is the output from evaluating the
	// OP_EQUAL) above from the stack.
	builder.AddOp(txscript.OP_DROP)

	// With that item dropped off, we can now enforce the absolute
	// lock-time required to timeout the PTLC. If the time has passed, then
	// we'll proceed with a checksig to ensure that this is actually the
	// sender of he original PTLC.
	builder.AddInt64(int64(cltvExpiry))
	builder.AddOp(txscript.OP_CHECKLOCKTIMEVERIFY)
	builder.AddOp(txscript.OP_DROP)
	builder.AddOp(txscript.OP_CHECKSIG)

	// Close out the inner if statement.
	builder.AddOp(txscript.OP_ENDIF)

	// Add 1 block CSV delay for non-revocation clauses if confirmation is
	// required.
	if confirmedSpend {
		builder.AddOp(txscript.OP_1)
		builder.AddOp(txscript.OP_CHECKSEQUENCEVERIFY)
		builder.AddOp(txscript.OP_DROP)
	}

	// Close out the outer if statement.
	builder.AddOp(txscript.OP_ENDIF)

	return builder.Script()
}

// ReceiverPtlcSpendRedeem constructs a valid witness allowing the receiver of
// an PTLC to redeem the conditional payment in the event that their commitment
// transaction is broadcast. This clause transitions the state of the PLTC
// output into the delay+claim state by activating the off-chain covenant bound
// by the 2-of-2 multi-sig output. The PTLC success timeout transaction being
// signed has a relative timelock delay enforced by its sequence number. This
// delay give the sender of the PTLC enough time to revoke the output if this
// is a breach commitment transaction.
//
// The corresponding clause from ReceiverPTLCScript() exercised by this function
// is:
//
//    RECVR: <sender sig> <recvr sig> <ptlc full sig> 1 (spend using HTLC success transaction)
//
// If both the adaptor sig and payment scalar are specified, the full ptlc sig
// is filled in. If only the adaptor sig is specified, then a data push for the
// serialized adaptor sig is added in place of the full sig. Note that such
// script is not consensus-valid and callers are responsible to assemble a full
// ptlc sig to make it so. If the adaptor sig is not specified, then an empty
// data push is added.
func ReceiverPtlcSpendRedeem(senderSig Signature,
	senderSigHash txscript.SigHashType, adaptorSig *dcr_adaptor_sigs.AdaptorSignature,
	paymentScalar *dcr_adaptor_sigs.SecretScalar,
	signer Signer, signDesc *SignDescriptor, htlcSuccessTx *wire.MsgTx) (
	TxWitness, error) {

	// First, we'll generate a signature for the HTLC success transaction.
	// The signDesc should be signing with the public key used as the
	// receiver's public key and also the correct single tweak.
	sweepSig, err := signer.SignOutputRaw(htlcSuccessTx, signDesc)
	if err != nil {
		return nil, err
	}

	var ptlcFullSigBytes []byte
	switch {
	case adaptorSig != nil && paymentScalar != nil:
		ptlcSig, err := dcr_adaptor_sigs.AssembleFullSig(adaptorSig, paymentScalar)
		if err != nil {
			return nil, err
		}
		ptlcFullSigBytes = ptlcSig.Serialize()

	case adaptorSig != nil:
		ptlcFullSigBytes = adaptorSig.Serialize()
	}

	// The final witness stack is used the provide the script with the
	// payment pre-image, and also execute the multi-sig clause after the
	// pre-images matches.
	witnessStack := TxWitness(make([][]byte, 5))
	witnessStack[0] = append(senderSig.Serialize(), byte(senderSigHash))
	witnessStack[1] = append(sweepSig.Serialize(), byte(signDesc.HashType))
	witnessStack[2] = ptlcFullSigBytes
	witnessStack[3] = []byte{txscript.OP_1}
	witnessStack[4] = signDesc.WitnessScript

	return witnessStack, nil
}

// ReplaceReceiverPtlcSpendRedeemPreimage replaces the ptlc sig in a sigScript
// generated by receiverPtlcSpendRedeemPreimage with a new, valid, full sig
// derived from an adaptor sig.  This will generate an undefined result script
// if the sigScript was *not* generated by the provided function (ie: this must
// be the sigScript of a SignedSuccessTx, used to redeem an PTLC).
//
// If the adaptor sig is empty but the paymentScalar is specified, then an
// attempt is made at deserializing the adaptor script from the passed
// sigScript and the final sig is then assembled.
//
// This function returns the new sigScript to replace the older sigScript.
func ReplaceReceiverPtlcSpendRedeemPreimage(sigScript []byte,
	adaptorSig *dcr_adaptor_sigs.AdaptorSignature,
	paymentScalar *dcr_adaptor_sigs.SecretScalar) ([]byte, error) {

	sigScriptPushes, err := SigScriptToWitnessStack(sigScript)
	if err != nil {
		return nil, err
	}
	if len(sigScriptPushes) != 5 {
		return nil, fmt.Errorf("the provided sigScript does not have 5 elements")
	}

	switch {
	case adaptorSig == nil && len(sigScriptPushes[2]) != dcr_adaptor_sigs.AdaptorSignatureSerializeLen:
		return nil, fmt.Errorf("adaptor signature is not serialized in " +
			"the existing sigScript")

	case adaptorSig == nil:
		adaptorSig, err = dcr_adaptor_sigs.ParseAdaptorSignature(sigScriptPushes[2])
		if err != nil {
			return nil, fmt.Errorf("invalid adaptor signature "+
				"serialized in sigScript: %v", err)
		}

	}

	ptlcSig, err := dcr_adaptor_sigs.AssembleFullSig(adaptorSig, paymentScalar)
	if err != nil {
		return nil, err
	}
	ptlcFullSigBytes := ptlcSig.Serialize()
	ptlcFullSigBytes = append(ptlcFullSigBytes, byte(txscript.SigHashAll))

	return txscript.NewScriptBuilder().
		AddData(sigScriptPushes[0]). // sender sig
		AddData(sigScriptPushes[1]). // receiver sig
		AddData(ptlcFullSigBytes).   // ptlc sig
		AddData(sigScriptPushes[3]). // 1
		AddData(sigScriptPushes[4]). // redeem script
		Script()
}

func PayToPubkeyHashScript(pubKey *secp256k1.PublicKey) ([]byte, error) {
	pkh := dcrutil.Hash160(pubKey.SerializeCompressed())
	return txscript.NewScriptBuilder().
		AddOp(txscript.OP_DUP).
		AddOp(txscript.OP_HASH160).
		AddData(pkh).
		AddOp(txscript.OP_EQUALVERIFY).
		AddOp(txscript.OP_CHECKSIG).
		Script()
}
