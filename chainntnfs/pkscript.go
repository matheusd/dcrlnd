// Adapted from the upstream decred/dcrd file contained in the txscript
// package.

package chainntnfs

import (
	"errors"
	"fmt"

	"github.com/decred/dcrd/chaincfg/v2"
	"github.com/decred/dcrd/dcrec"
	"github.com/decred/dcrd/dcrutil/v2"
	"github.com/decred/dcrd/txscript/v2"
)

const (
	// pubKeyHashSigScriptLen is the length of a signature script attempting
	// to spend a P2PKH script. The only other possible length value is 107
	// bytes, due to the signature within it. This length is determined by
	// the following:
	//   0x47 or 0x48 (71 or 72 byte data push) | <71 or 72 byte sig> |
	//   0x21 (33 byte data push) | <33 byte compressed pubkey>
	pubKeyHashSigScriptLen = 106

	// sigLen is the maximum length of a signature data push in a p2pkh
	// sigScript.
	sigLen = 72

	// compressedPubKeyLen is the length in bytes of a compressed public
	// key.
	compressedPubKeyLen = 33

	// pubKeyHashLen is the length of a P2PKH script.
	pubKeyHashLen = 25

	// scriptHashLen is the length of a P2SH script.
	scriptHashLen = 23

	// maxLen is the maximum script length supported by ParsePkScript.
	maxLen = pubKeyHashLen
)

var (
	// ErrUnsupportedScriptType is an error returned when we attempt to
	// parse/re-compute an output script into a PkScript struct.
	ErrUnsupportedScriptType = errors.New("unsupported script type")
)

// PkScript is a wrapper struct around a byte array, allowing it to be used
// as a map index.
type PkScript struct {
	// class is the type of the script encoded within the byte array. This
	// is used to determine the correct length of the script within the byte
	// array.
	class txscript.ScriptClass

	// script is the script contained within a byte array. If the script is
	// smaller than the length of the byte array, it will be padded with 0s
	// at the end.
	script [maxLen]byte

	// scriptVersion is the script version of the given pkscript. Given
	// this is _not_ embedded in the pkscript itself, it must be provided
	// externally.
	scriptVersion uint16
}

// ParsePkScript parses an output script into the PkScript struct.
// ErrUnsupportedScriptType is returned when attempting to parse an unsupported
// script type.
func ParsePkScript(scriptVersion uint16, pkScript []byte) (PkScript, error) {
	if scriptVersion != 0 {
		return PkScript{}, fmt.Errorf("unsupported script version %d "+
			"(only supports version 0)", scriptVersion)
	}

	outputScript := PkScript{scriptVersion: scriptVersion}
	scriptClass, _, _, err := txscript.ExtractPkScriptAddrs(
		scriptVersion, pkScript, chaincfg.MainNetParams(),
	)
	if err != nil {
		return outputScript, fmt.Errorf("unable to parse script type: "+
			"%v", err)
	}

	if !isSupportedScriptType(scriptClass) {
		return outputScript, ErrUnsupportedScriptType
	}

	outputScript.class = scriptClass
	copy(outputScript.script[:], pkScript)

	return outputScript, nil
}

// isSupportedScriptType determines whether the script type is supported by the
// PkScript struct.
func isSupportedScriptType(class txscript.ScriptClass) bool {
	switch class {
	case txscript.PubKeyHashTy, txscript.ScriptHashTy:
		return true
	default:
		return false
	}
}

// Class returns the script type.
func (s PkScript) Class() txscript.ScriptClass {
	return s.class
}

// Script returns the script as a byte slice without any padding. This is a
// copy of the original script, therefore it's safe for modification.
func (s PkScript) Script() []byte {
	var script []byte

	switch s.class {
	case txscript.PubKeyHashTy:
		script = make([]byte, pubKeyHashLen)
		copy(script, s.script[:pubKeyHashLen])

	case txscript.ScriptHashTy:
		script = make([]byte, scriptHashLen)
		copy(script, s.script[:scriptHashLen])

	default:
		// Unsupported script type.
		return nil
	}

	return script
}

// Address encodes the script into an address for the given chain.
func (s PkScript) Address(chainParams *chaincfg.Params) (dcrutil.Address, error) {
	var (
		address dcrutil.Address
		err     error
	)

	switch s.class {
	case txscript.PubKeyHashTy:
		scriptHash := s.script[3:23]
		address, err = dcrutil.NewAddressPubKeyHash(
			scriptHash, chainParams, dcrec.STEcdsaSecp256k1,
		)
	case txscript.ScriptHashTy:
		scriptHash := s.script[1:21]
		address, err = dcrutil.NewAddressScriptHashFromHash(
			scriptHash, chainParams,
		)
	default:
		err = ErrUnsupportedScriptType
	}

	if err != nil {
		return nil, err
	}
	return address, nil
}

// String returns a hex-encoded string representation of the script.
func (s PkScript) String() string {
	str, _ := txscript.DisasmString(s.Script())
	return str
}

// ComputePkScript computes the pkScript of an transaction output by looking at
// the transaction input's signature script.
//
// NOTE: Only P2PKH and P2SH redeem scripts are supported. Only the standard
// secp256k1 keys are supported (alternative suites are not).
func ComputePkScript(scriptVersion uint16, sigScript []byte) (PkScript, error) {

	var pkScript PkScript

	if scriptVersion != 0 {
		return pkScript, fmt.Errorf("unsupported script version %d "+
			"(only supports version 0)", scriptVersion)
	}

	// Ensure that either an input's signature script or a witness was
	// provided.
	if len(sigScript) == 0 {
		return pkScript, ErrUnsupportedScriptType
	}

	// Create a tokenizer and decode up to the last opcode. Store the first
	// data as well, to check for the correct p2kh sig script style.
	tokenizer := txscript.MakeScriptTokenizer(
		scriptVersion, sigScript,
	)
	var opcodeCount int
	var firstData []byte
	for tokenizer.Next() {
		if tokenizer.Opcode() > txscript.OP_16 {
			return pkScript, ErrUnsupportedScriptType
		}
		if opcodeCount == 0 {
			firstData = tokenizer.Data()
		}
		opcodeCount++
	}
	if tokenizer.Err() != nil {
		return pkScript, tokenizer.Err()
	}

	var scriptClass txscript.ScriptClass
	var address dcrutil.Address
	var err error

	// The last opcode of a sigscript will either be a pubkey (for p2kh
	// pkscripts) or a redeem script (for p2sh pkscripts). Further, a
	// standard p2pkh will only have an extra signature data push.
	lastData := tokenizer.Data()
	lastDataHash := dcrutil.Hash160(lastData)
	firstDataIsSigLen := len(firstData) == sigLen ||
		len(firstData) == sigLen-1
	lastDataIsPubkeyLen := len(lastData) == compressedPubKeyLen
	if opcodeCount == 2 && firstDataIsSigLen && lastDataIsPubkeyLen {
		// The sigScript has the correct structure for spending a
		// p2pkh, therefore assume it is one.
		scriptClass = txscript.PubKeyHashTy
		address, err = dcrutil.NewAddressPubKeyHash(
			lastDataHash, chaincfg.MainNetParams(),
			dcrec.STEcdsaSecp256k1,
		)
	} else {
		// Assume it's a p2sh.
		scriptClass = txscript.ScriptHashTy
		address, err = dcrutil.NewAddressScriptHashFromHash(
			lastDataHash, chaincfg.MainNetParams(),
		)
	}
	if err != nil {
		return pkScript, err
	}

	scriptSlice, err := txscript.PayToAddrScript(address)
	if err != nil {
		return pkScript, err
	}
	var script [maxLen]byte
	copy(script[:], scriptSlice)

	return PkScript{
		class:         scriptClass,
		scriptVersion: scriptVersion,
		script:        script,
	}, nil
}
