package input

import (
	"github.com/decred/dcrd/wire"
)

// Quick review of the serialized layout of decred transactions. This is
// applicable for version 1 serialization type, when full serialization is
// performed (ie: tx.Version == 1, tx.SerType: TxSerializeFull).
//
//		- Version+SerType                   ┐
//		- Input Count (varint)              │
//		- (in_count times) Input Prefix     ├  Prefix Serialization
//		- Output Count (varint)             │
//		- (out_count times) Output          │
//		- LockTime+Expiry                   ┘
//		- Input Count (varint)              ┬  Witness Serialization
//		- (in_count times) Input Witness    ┘

const (
	// baseTxSize is the size of all transaction-level data elements serialized,
	// stored and relayed for a transaction. When calculating the full serialized
	// size of a transaction, add the length of all the inputs, outputs and 3
	// varints (one for encoding the length of outputs and 2 for encoding the
	// length of inputs). It is calculated as:
	//
	//		- version + serialization type        4 bytes
	//		- locktime                            4 bytes
	//		- expiry                              4 bytes
	//
	// Total: 12 bytes
	baseTxSize int64 = 4 + 4 + 4

	// InputSize is the size of the fixed (always present) elements serialized,
	// stored and relayed for each transaction input. When calculating the full
	// serialized size of an input, add the length of the corresponding
	// sigScript and of the varint that encodes the length of the sigScript. It
	// is calculated as:
	//
	//		- PreviousOutPoint:                   ┐
	//		    - hash                32 bytes    │
	//		    - index                4 bytes    ├  Part of Prefix Serialization
	//		    - tree                 1 byte     │
	//		- Sequence                 4 bytes    │
	//		                                      ┘
	//		                                      ┐
	//		- ValueIn                8 bytes      │
	//		- Height                 4 bytes      ├  Part of Witness Serialization
	//		- Index                  4 bytes      │
	//		                                      ┘
	// Total: 57 bytes
	InputSize int64 = 32 + 4 + 1 + 4 + 8 + 4 + 4

	// OutputSize is the size of the fixed (always present) elements serialized,
	// stored and relayed for each transaction output. When calculating the full
	// serialized size of an output, add the length of the corresponding
	// pkscript and of the varint that encodes the length of the pkscript. It is
	// calculated as:
	//
	//		- Value                    8 bytes
	//		- ScriptVersion            2 bytes
	//
	// Total: 10 bytes
	OutputSize int64 = 8 + 2

	// The Following P2*PkScriptSize constants record the size of the standard
	// public key scripts used in decred transactions' outputs.

	// P2PKHPkScriptSize is the size of a transaction output script that
	// pays to a compressed pubkey hash.  It is calculated as:
	//
	//		- OP_DUP                  1 byte
	//		- OP_HASH160              1 byte
	//		- OP_DATA_20              1 byte
	//		- pubkey hash            20 bytes
	//		- OP_EQUALVERIFY          1 byte
	//		- OP_CHECKSIG             1 byte
	//
	// Total: 25 bytes
	P2PKHPkScriptSize int64 = 1 + 1 + 1 + 20 + 1 + 1

	// P2SHPkScriptSize is the size of a transaction output script that
	// pays to a script hash.  It is calculated as:
	//
	//		- OP_HASH160               1 byte
	//		- OP_DATA_20               1 byte
	//		- script hash             20 bytes
	//		- OP_EQUAL                 1 byte
	//
	// Total: 23 bytes
	P2SHPkScriptSize int64 = 1 + 1 + 20 + 1

	// The Following *SigScriptSize constants record the worst possible
	// size of the standard signature scripts used to redeem the corresponding
	// public key scripts in decred transactions' input.

	// P2PKHSigScriptSize is the worst case (largest) serialize size of a
	// transaction input script that redeems a compressed P2PKH output. It is
	// calculated as:
	//
	//      - OP_DATA_73                 1 byte
	//      - signature+hash_type       73 bytes
	//      - OP_DATA_33                 1 byte
	//      - compressed pubkey         33 bytes
	//
	// Total: 108 bytes
	P2PKHSigScriptSize int64 = 1 + 73 + 1 + 33

	// The following **RedeemScriptSize constants record sizes for LN-specific
	// redeem scripts that are pushed to SigScripts when redeeming LN-specific
	// P2SH outputs.

	// multiSig2Of2RedeemScriptSize is the size of a 2-of-2 multisig script. It is
	// calculated as:
	//
	//		- OP_2                     1 byte
	//		- OP_DATA_33               1 byte
	//		- pubkey_alice            33 bytes
	//		- OP_DATA_33               1 byte
	//		- pubkey_bob              33 bytes
	//		- OP_2                     1 byte
	//		- OP_CHECKMULTISIG         1 byte
	//
	// Total: 71 bytes
	multiSig2Of2RedeemScriptSize int64 = 1 + 1 + 33 + 1 + 33 + 1 + 1

	// toLocalRedeemScriptSize is the worst (largest) size of a redeemScript used in
	// RSMC outputs for the "local" node; in other words, it's the size of the
	// script for those outputs that may be redeemed by the local node after a
	// delay or by the counterparty by using a breach remedy key/transaction.
	// The size is calculated as:
	//
	//		- OP_IF                               1 byte
	//		    - OP_DATA_33                      1 byte
	//		    - revoke_key                     33 bytes
	//		- OP_ELSE                             1 byte
	//		    - OP_DATA_5                       1 byte
	//		    - csv_delay                       5 bytes
	//		    - OP_CHECKSEQUENCEVERIFY          1 byte
	//		    - OP_DROP                         1 byte
	//		    - OP_DATA_33                      1 byte
	//		    - delay_key                      33 bytes
	//		- OP_ENDIF                            1 byte
	//		- OP_CHECKSIG                         1 byte
	//
	// Total: 80 bytes
	//
	// TODO(decred) verify whether the maximum csv_delay can actually occupy the
	// full 5 bytes (which is the maximum used by OP_CHECKSEQUENCEVERIFY).
	toLocalRedeemScriptSize int64 = 1 + 1 + 33 + 1 + 1 + 5 + 1 + 1 + 1 + 33 + 1 + 1

	// acceptedHtlcRedeemScriptSize is the worst (largest) size of a
	// redeemScript used by the local node when receiving payment via an HTLC
	// output. In BOLT03 this is called a "Received HTLC Output".
	//
	// Currently generated by receiverHTLCScript().
	//
	// This is calculated as:
	//
	//      - OP_DUP                                         1 byte
	//      - OP_HASH160                                     1 byte
	//      - OP_DATA_20                                     1 byte
	//      - RIPEMD160(SHA256(revocationkey))              20 bytes
	//      - OP_EQUAL                                       1 byte
	//      - OP_IF                                          1 byte
	//              - OP_CHECKSIG                            1 byte
	//      - OP_ELSE                                        1 byte
	//              - OP_DATA_33                             1 byte
	//              - remotekey                             33 bytes
	//              - OP_SWAP                                1 byte
	//              - OP_SIZE                                1 byte
	//              - OP_DATA_32                             1 byte
	//              - OP_EQUAL                               1 byte
	//              - OP_IF                                  1 byte
	//                      - OP_HASH160                     1 byte
	//                      - OP_DATA_20                     1 byte
	//                      - RIPEMD160(payment_hash)       20 bytes
	//                      - OP_EQUALVERIFY                 1 byte
	//                      - OP_DATA_2                      1 byte
	//                      - OP_SWAP                        1 byte
	//                      - OP_DATA_33                     1 byte
	//                      - localkey                      33 bytes
	//                      - OP_DATA_2                      1 byte
	//                      - OP_CHECKMULTISIG               1 byte
	//              - OP_ELSE                                1 byte
	//                      - OP_DROP                        1 byte
	//                      - OP_DATA_5                      1 byte
	//                      - cltv_expiry                    5 bytes
	//                      - OP_CHECKLOCKTIMEVERIFY         1 byte
	//                      - OP_DROP                        1 byte
	//                      - OP_CHECKSIG                    1 byte
	//              - OP_ENDIF                               1 byte
	//      - OP_ENDIF                                       1 byte
	//
	// Total: 140 bytes
	//
	// TODO(decred) verify whether the maximum cltv_expirt can actually occupy
	// the full 5 bytes (which is the maximum used by OP_CHECKLOCKTIMEVERIFY).
	acceptedHtlcRedeemScriptSize int64 = 3*1 + 20 + 5*1 + 33 + 7*1 + 20 + 4*1 +
		33 + 5*1 + 5 + 5*1

	// offeredHtlcRedeemScriptSize is the worst (largest) size of a redeemScript used
	// by the local node when sending payment via an HTLC output.
	//
	// Currently generated by senderHTLCScript().
	//
	// This is calculated as:
	//
	//		- OP_DUP                                     1 byte
	//		- OP_HASH160                                 1 byte
	//		- OP_DATA_20                                 1 byte
	//		- RIPEMD160(SHA256(revocationkey))          20 bytes
	//		- OP_EQUAL                                   1 byte
	//		- OP_IF                                      1 byte
	//		        - OP_CHECKSIG                        1 byte
	//		- OP_ELSE                                    1 byte
	//		        - OP_DATA_33                         1 byte
	//		        - remotekey                         33 bytes
	//		        - OP_SWAP                            1 byte
	//		        - OP_SIZE                            1 byte
	//		        - OP_DATA_1                          1 byte
	//		        - OP_DATA_32                         1 byte
	//		        - OP_EQUAL                           1 byte
	//		        - OP_NOTIF                           1 byte
	//		                - OP_DROP                    1 byte
	//		                - OP_DATA_2                  1 byte
	//		                - OP_SWAP                    1 byte
	//		                - OP_DATA_33                 1 byte
	//		                - localkey                  33 bytes
	//		                - OP_DATA_2                  1 byte
	//		                - OP_CHECKMULTISIG           1 byte
	//		        - OP_ELSE                            1 byte
	//		                - OP_HASH160                 1 byte
	//		                - OP_DATA_20                 1 byte
	//		                - RIPEMD160(payment_hash)   20 bytes
	//		                - OP_EQUALVERIFY             1 byte
	//		                - OP_CHECKSIG                1 byte
	//		        - OP_ENDIF                           1 byte
	//		- OP_ENDIF                                   1 byte
	//
	// Total: 133 bytes
	offeredHtlcRedeemScriptSize int64 = 3*1 + 20 + 5*1 + 33 + 10*1 + 33 + 5*1 + 20 + 4*1

	// The following *SigScript constants record sizes for various types of
	// LN-specific sigScripts, spending outputs that use one of the custom
	// redeem scripts. These constants are the sum of the script data push plus
	// the actual sig script data required for redeeming one of the script's
	// code paths.
	//
	// All constants are named according to the schema
	// [tx-type][code-path]sigScriptSize. See the above *RedeemScriptSize
	// comments for explanations of each possible tx type/redeem script.

	// FundingOutputSigScriptSize is the size of a sigScript used when
	// redeeming a funding transaction output. This includes signatures for
	// both alice's and bob's keys plus the 2-of-2 multisig redeemScript. It
	// is calculated as:
	//
	//		- OP_DATA_73                     1 byte
	//		- alice_sig+hash_type           73 bytes
	//		- OP_DATA_73                     1 byte
	//		- bob_sig+hash_type             73 bytes
	//		- OP_DATA_71                     1 byte
	//		- multisig_2of2_script          71 bytes
	//
	// Total: 220 bytes
	FundingOutputSigScriptSize int64 = 1 + 73 + 1 + 73 + 1 +
		multiSig2Of2RedeemScriptSize

	// ToLocalTimeoutSigScriptSize is the size of sigScript used when
	// redeeming a toLocalScript using the "timeout" code path.
	//
	//		- OP_DATA_73                     1 byte
	//		- local_delay_sig+hash_type     73 bytes
	//		- OP_0                           1 byte
	//		- OP_PUSHDATA1                   1 byte
	//		- 80                             1 byte
	//		- to_local_timeout script       80 bytes
	//
	// Total: 157 bytes
	ToLocalTimeoutSigScriptSize int64 = 1 + 73 + 1 + 1 + 1 +
		toLocalRedeemScriptSize

	// ToLocalPenaltySigScriptSize is the size of a sigScript used when
	// redeeming a toLocalScript using the "penalty" code path.
	//
	//		- OP_DATA_73                      1 byte
	//		- revocation_sig+hash_type       73 bytes
	//		- OP_TRUE                         1 byte
	//		- OP_PUSHDATA1                    1 byte
	//		- 80                              1 byte
	//		- to_local_timeout script        80 bytes
	//
	// Total: 157 bytes
	// old ToLocalPenaltyWitnessSize
	ToLocalPenaltySigScriptSize int64 = 1 + 73 + 1 + 1 + 1 +
		toLocalRedeemScriptSize

	// AcceptedHtlcTimeoutSigScriptSize is the size of a sigScript used
	// when redeeming an acceptedHtlcScript using the "timeout" code path.
	//
	//		- OP_DATA_73                      1 byte
	//		- sender_sig+hash_type           73 bytes
	//		- OP_0                            1 byte
	//		- OP_PUSHDATA1                    1 byte
	//		- 140                             1 byte
	//		- accepted_htlc script          140 bytes
	//
	// Total: 217 bytes
	AcceptedHtlcTimeoutSigScriptSize int64 = 1 + 73 + 1 + 1 + 1 +
		acceptedHtlcRedeemScriptSize

	// AcceptedHtlcSuccessSigScriptSize is the size of a sigScript used
	// when redeeming an acceptedHtlcScript using the "success" code path.
	//
	//		- OP_DATA_73                         1 byte
	//		- sig_alice+hash_type               73 bytes
	//		- OP_DATA_73                         1 byte
	//		- sig_bob+hash_type                 73 bytes
	//		- OP_DATA_32                         1 byte
	//		- payment_preimage                  32 bytes
	//		- OP_PUSHDATA1                       1 byte
	//		- 140                                1 byte
	//		- accepted_htlc script             140 bytes
	//
	// Total: 323 bytes
	AcceptedHtlcSuccessSigScriptSize int64 = 1 + 73 + 1 + 73 + 1 + 32 +
		1 + 1 + acceptedHtlcRedeemScriptSize

	// AcceptedHtlcPenaltySigScriptSize is the size of a sigScript used
	// when redeeming an acceptedHtlcScript using the "penalty" code path.
	//
	//		- OP_DATA_73                        1 byte
	//		- revocation_sig+hash_type         73 bytes
	//		- OP_DATA_33                        1 byte
	//		- revocation_key                   33 bytes
	//		- OP_PUSHDATA1                      1 byte
	//		- 140                               1 byte
	//		- accepted_htlc script            140 bytes
	//
	// Total: 250 bytes
	AcceptedHtlcPenaltySigScriptSize int64 = 1 + 73 + 1 + 33 + 1 + 1 +
		acceptedHtlcRedeemScriptSize

	// OfferedHtlcTimeoutSigScriptSize is the size of a sigScript used
	// when redeeming an offeredHtlcScript using the "timeout" code path.
	//
	//		- OP_DATA_73                         1 byte
	//		- sig_alice+hash_type               73 bytes
	//		- OP_DATA_73                         1 byte
	//		- sig_bob+hash_type                 73 bytes
	//		- OP_0                               1 byte
	//		- OP_PUSHDATA1                       1 byte
	//		- 133                                1 byte
	//		- offered_htlc script              133 bytes
	//
	// Total: 284 bytes
	OfferedHtlcTimeoutSigScriptSize int64 = 1 + 73 + 1 + 73 + 1 + 1 +
		1 + offeredHtlcRedeemScriptSize

	// OfferedHtlcSuccessSigScriptSize is the size of a sigScript used
	// when redeeming an offeredHtlcScript using the "success" code path.
	//
	//		- OP_DATA_73                      1 byte
	//		- receiver_sig+hash_type         73 bytes
	//		- OP_DATA_73                      1 byte
	//		- sender_sig+hash_type           73 bytes
	//		- OP_DATA_32                      1 byte
	//		- payment_preimage               32 bytes
	//		- OP_PUSHDATA1                    1 byte
	//		- 133                             1 byte
	//		- offered_htlc script           133 bytes
	//
	// Total: 316 bytes
	OfferedHtlcSuccessSigScriptSize int64 = 1 + 73 + 1 + 73 + 1 + 32 +
		1 + 1 + offeredHtlcRedeemScriptSize

	// OfferedHtlcPenaltySigScriptSize is the size of a sigScript used
	// when redeeming an offeredHtlcScript using the "penalty" code path.
	//
	//		- OP_DATA_73                      1 byte
	//		- revocation_sig+hash_type       73 bytes
	//		- OP_DATA_33                      1 byte
	//		- revocation_key                 33 bytes
	//		- OP_PUSHDATA1                    1 byte
	//		- 133                             1 byte
	//		- offered_htlc script           133 bytes
	//
	// Total: 243 bytes
	OfferedHtlcPenaltySigScriptSize int64 = 1 + 73 + 1 + 33 + 1 + 1 +
		offeredHtlcRedeemScriptSize

	// The following constants record pre-calculated inputs, outputs and
	// transaction sizes for common transactions found in the LN ecosystem.

	// HTLCOutputSize is the size of an HTLC Output (a p2sh output) used in
	// commitment transactions.
	//
	//		- Output (value+version)        10 bytes
	//		- pkscript varint                1 byte
	//		- p2sh pkscript                 23 bytes
	//
	// Total: 34 bytes
	HTLCOutputSize int64 = OutputSize + 1 + P2SHPkScriptSize

	// CommitmentTxSize is the base size of a commitment transaction without any
	// HTLCs.
	//
	// Note: This uses 2 byte varints for output counts to account for the fact
	// that a full commitment transaction using the maximum allowed number of
	// HTLCs may use one extra byte for the output count varint.
	//
	// It is calculated as:
	//
	//		- base tx size                             12 bytes
	//		- input count prefix varint                 1 byte
	//		- input                                    57 bytes
	//		- output count prefix varint                2 bytes
	//		- remote output                            10 bytes
	//		- p2pkh remote varint                       1 byte
	//		- p2pkh remote pkscript                    25 bytes
	//		- local output                             10 bytes
	//		- p2sh local varint                         1 byte
	//		- p2sh local pkscript                      23 bytes
	//		- input count witness varint                1 byte
	//		- funding tx sigscript varint               1 byte
	//		- funding tx sigscript                    220 bytes
	//
	// Total: 364 bytes
	CommitmentTxSize int64 = baseTxSize + 1 + InputSize + 2 +
		OutputSize + 1 + P2PKHPkScriptSize + OutputSize + 1 + P2SHPkScriptSize +
		1 + 1 + FundingOutputSigScriptSize

	// HTLCTimeoutSize is the worst case (largest) size of the HTLC timeout
	// transaction which will transition an outgoing HTLC to the
	// delay-and-claim state. The worst case for a timeout transaction is
	// when redeeming an offered HTCL (which uses a larger sigScript). It
	// is calculated as:
	//
	//		- base tx size                                     12 bytes
	//		- input count prefix varint                         1 byte
	//		- input                                            57 bytes
	//		- output count prefix varint                        1 byte
	//		- output                                           10 bytes
	//		- p2sh pkscript varint                              1 byte
	//		- p2sh pkscript                                    23 bytes
	//		- input count witness varint                        1 byte
	//		- offered_htlc_timeout sigscript varint             2 bytes
	//		- offered_htlc_timeout sigscript                  284 bytes
	//
	// Total: 392 bytes
	// TODO(decred) Double check correctness of selected sigScript alternative
	HTLCTimeoutTxSize int64 = baseTxSize + 1 + InputSize + 1 + OutputSize + 1 +
		P2SHPkScriptSize + 1 + 2 + OfferedHtlcTimeoutSigScriptSize

	// HTLCSuccessSize is the worst case (largest) size of the HTLC success
	// transaction which will transition an HTLC tx to the delay-and-claim
	// state. The worst case for a success transaction is when redeeming an
	// accepted HTLC (which has a larger sigScript). It is calculated as:
	//
	//		- base tx Size                                   12 bytes
	//		- input count prefix varint                       1 byte
	//		- input                                          57 bytes
	//		- output count prefix varint                      1 byte
	//		- output                                         10 bytes
	//		- p2pkh pkscript varint                           1 byte
	//		- p2pkh pkscript                                 25 bytes
	//		- input count witness varint                      1 byte
	//		- accepted_htlc_success sigscript varint          2 bytes
	//		- accepted_htlc_timeout sigscript               323 bytes
	//
	// Total: 433 bytes
	// TODO(decred) Double check correctness of selected sigScript alternative
	HTLCSuccessTxSize int64 = baseTxSize + 1 + InputSize + 1 + OutputSize + 1 +
		P2PKHPkScriptSize + 1 + 2 + AcceptedHtlcSuccessSigScriptSize

	// MaxHTLCNumber is the maximum number HTLCs which can be included in a
	// commitment transaction. This limit was chosen such that, in the case
	// of a contract breach, the punishment transaction is able to sweep
	// all the HTLC's yet still remain below the widely used standard size
	// limits.
	//
	// This number is derived (as explained in BOLT-0005) by assuming a
	// penalty transaction will redeem the following elements (along with
	// their respective sizes):
	//
	// 		- base tx size				 12 bytes
	//		- output count varint			  1 byte
	//		- p2pkh output				 36 bytes
	//		- input count prefix varint		  2 bytes
	//		- input count witness varint		  2 bytes
	//		- to_remote commitment output
	//			- input 			 57 bytes
	//			- sigscript varint		  1 byte
	//			- 2-of-2 multisig sigscript 	220 bytes
	//		- to_local commitment output
	//			- input				 57 bytes
	//			- sigscript varint		  1 byte
	//			- to_local penalty sigscript	157 bytes
	//		- n accepted_htlc_penalty inputs
	//			- input				 57 bytes
	//			- sigscript varint		  1 byte
	//			- sigscript			250 bytes
	//
	// The "n" maximum number of redeemable htlcs can thus be calculated
	// (where static_data is everything _except_ the variable number of
	// htlc outputs):
	//
	//	= (max_tx_size - static_data) / accepted_htlc_penalty_size
	//      = (  100000    -     546   )  /      308
	//      = 322 htlcs
	MaxHTLCNumber = 322
)

// EstimateCommitmentTxSize estimates the size of a commitment transaction
// assuming that it has an additional 'count' HTLC outputs appended to it.
func EstimateCommitmentTxSize(count int) int64 {

	// Size of 'count' HTLC outputs.
	htlcsSize := int64(count) * HTLCOutputSize

	return CommitmentTxSize + htlcsSize
}

// TxSizeEstimator is able to calculate size estimates for transactions based on
// the input and output types. For purposes of estimation, all signatures are
// assumed to be of the maximum possible size, 73 bytes. Each method of the
// estimator returns an instance with the estimate applied. This allows callers
// to chain each of the methods
type TxSizeEstimator struct {
	inputCount  uint32
	outputCount uint32
	InputSize   int64
	OutputSize  int64
}

// AddP2PKHInput updates the size estimate to account for an additional input
// spending a P2PKH output.
func (twe *TxSizeEstimator) AddP2PKHInput() *TxSizeEstimator {
	scriptLenSerSize := int64(1) // varint for the following sigScript
	twe.InputSize += InputSize + scriptLenSerSize + P2PKHSigScriptSize
	twe.inputCount++

	return twe
}

// AddCustomInput updates the size estimate to account for an additional input,
// such that the caller is responsible for specifying the full estimated size of
// the sigScript.
//
// Note that the caller is entirely responsible for calculating the correct size
// of the sigScript. This function only adds the overhead of the fixed input
// data (prefix serialization) and of the varint for recording the sigScript
// size.
func (twe *TxSizeEstimator) AddCustomInput(sigScriptSize int64) *TxSizeEstimator {
	scriptLenSerSize := int64(wire.VarIntSerializeSize(uint64(sigScriptSize)))
	twe.InputSize += InputSize + scriptLenSerSize + sigScriptSize
	twe.inputCount++

	return twe
}

// AddP2PKHOutput updates the size estimate to account for an additional P2PKH
// output.
func (twe *TxSizeEstimator) AddP2PKHOutput() *TxSizeEstimator {
	scriptLenSerSize := int64(1) // varint for the following pkScript
	twe.OutputSize += OutputSize + scriptLenSerSize + P2PKHPkScriptSize
	twe.outputCount++

	return twe
}

// AddP2SHOutput updates the size estimate to account for an additional P2SH
// output.
func (twe *TxSizeEstimator) AddP2SHOutput() *TxSizeEstimator {
	scriptLenSerSize := int64(1) // varint for the following pkScript
	twe.OutputSize += OutputSize + scriptLenSerSize + P2SHPkScriptSize
	twe.outputCount++

	return twe
}

// Size gets the estimated size of the transaction.
func (twe *TxSizeEstimator) Size() int64 {
	return baseTxSize +
		int64(wire.VarIntSerializeSize(uint64(twe.inputCount))) + // prefix len([]TxIn) varint
		twe.InputSize + // prefix []TxIn + witness []TxIn
		int64(wire.VarIntSerializeSize(uint64(twe.outputCount))) + // prefix len([]TxOut) varint
		twe.OutputSize + // []TxOut prefix
		int64(wire.VarIntSerializeSize(uint64(twe.inputCount))) // witness len([]TxIn) varint
}
