package compat

import (
	chaincfg2 "github.com/decred/dcrd/chaincfg/v2"
	chaincfg3 "github.com/decred/dcrd/chaincfg/v3"
	dcrutil2 "github.com/decred/dcrd/dcrutil/v2"
	dcrutil3 "github.com/decred/dcrd/dcrutil/v3"
	hdkeychain2 "github.com/decred/dcrd/hdkeychain/v2"
	hdkeychain3 "github.com/decred/dcrd/hdkeychain/v3"
)

func Params2to3(v2 *chaincfg2.Params) *chaincfg3.Params {
	// Crappy way to do this, but dcrlnd doesn't change any params anywhere
	// (and this is temporary code) so whatever.
	switch v2.Name {
	case "mainnet":
		return chaincfg3.MainNetParams()
	case "testnet3":
		return chaincfg3.TestNet3Params()
	case "simnet":
		return chaincfg3.SimNetParams()
	case "regnet":
		return chaincfg3.RegNetParams()
	default:
		panic("oops!")
	}
}

func Params3to2(v3 *chaincfg3.Params) *chaincfg2.Params {
	// Crappy way to do this, but dcrlnd doesn't change any params anywhere
	// (and this is temporary code) so whatever.
	switch v3.Name {
	case "mainnet":
		return chaincfg2.MainNetParams()
	case "testnet3":
		return chaincfg2.TestNet3Params()
	case "simnet":
		return chaincfg2.SimNetParams()
	case "regnet":
		return chaincfg2.RegNetParams()
	default:
		panic("oops!")
	}
}

func Amount2to3(amt dcrutil2.Amount) dcrutil3.Amount {
	return dcrutil3.Amount(amt)
}

func Amount3to2(amt dcrutil3.Amount) dcrutil2.Amount {
	return dcrutil2.Amount(amt)
}

func Address3to2(addr dcrutil3.Address, net dcrutil2.AddressParams) (dcrutil2.Address, error) {
	s := addr.Address()
	return dcrutil2.DecodeAddress(s, net)
}

func ExtendedKey3to2(key *hdkeychain3.ExtendedKey, chain hdkeychain3.NetworkParams) *hdkeychain2.ExtendedKey {
	s := key.String()
	key2, err := hdkeychain2.NewKeyFromString(s, chain)
	if err != nil {
		panic(err)
	}
	return key2
}
