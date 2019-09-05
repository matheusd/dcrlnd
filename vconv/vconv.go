package vconv

import (
	hdv1 "github.com/decred/dcrd/hdkeychain"
	hdv2 "github.com/decred/dcrd/hdkeychain/v2"

	chaincfg1 "github.com/decred/dcrd/chaincfg"
	chaincfg2 "github.com/decred/dcrd/chaincfg/v2"

	dcrutil1 "github.com/decred/dcrd/dcrutil"
	dcrutil2 "github.com/decred/dcrd/dcrutil/v2"

	rpcclient2 "github.com/decred/dcrd/rpcclient/v2"
	rpcclient3 "github.com/decred/dcrd/rpcclient/v3"
)

// This file contains functions to convert to and from v1 and v2 versions of
// several packages.
//
// This will be removed once conversion is complete.

// Hd2to1 converts a hdkeychain/v2 extended key to the v1 API. An error check
// during string conversion is intentionally dropped for brevity.
func Hd2to1(k2 *hdv2.ExtendedKey) *hdv1.ExtendedKey {
	k, _ := hdv1.NewKeyFromString(k2.String())
	return k
}

// NetParams1to2 converts a chaincfg/v1 *Params to a chaincfg/v2 *Params. It
// panics if the provided params is not recognized.
func NetParams1to2(netParams1 *chaincfg1.Params) *chaincfg2.Params {
	switch {
	case netParams1 == &chaincfg1.RegNetParams:
		return chaincfg2.RegNetParams()
	case netParams1 == &chaincfg1.SimNetParams:
		return chaincfg2.SimNetParams()
	case netParams1 == &chaincfg1.TestNet3Params:
		return chaincfg2.TestNet3Params()
	case netParams1 == &chaincfg1.MainNetParams:
		return chaincfg2.MainNetParams()
	default:
		panic("Unknown chaincfg v1 parameters")
	}
}

func Addr2to1(addr2 dcrutil2.Address) dcrutil1.Address {
	if addr2 == nil {
		return nil
	}
	addr1, err := dcrutil1.DecodeAddress(addr2.String())
	if err != nil {
		panic(err)
	}
	return addr1
}

func Addr1to2(addr1 dcrutil1.Address, netParams *chaincfg2.Params) dcrutil2.Address {
	if addr1 == nil {
		return nil
	}
	addr2, err := dcrutil2.DecodeAddress(addr1.String(), netParams)
	if err != nil {
		panic(err)
	}
	return addr2
}

func RPCConfig3to2(config3 rpcclient3.ConnConfig) rpcclient2.ConnConfig {
	return rpcclient2.ConnConfig{
		Host:         config3.Host,
		Endpoint:     config3.Endpoint,
		User:         config3.User,
		Pass:         config3.Pass,
		DisableTLS:   config3.DisableTLS,
		Certificates: config3.Certificates,
	}
}

func RPCConfig2to3(config2 rpcclient2.ConnConfig) rpcclient3.ConnConfig {
	return rpcclient3.ConnConfig{
		Host:         config2.Host,
		Endpoint:     config2.Endpoint,
		User:         config2.User,
		Pass:         config2.Pass,
		DisableTLS:   config2.DisableTLS,
		Certificates: config2.Certificates,
	}
}
