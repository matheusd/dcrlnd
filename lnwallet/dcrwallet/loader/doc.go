// Copyright (c) 2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

/*
Package loader provides a concurrent safe implementation of a wallet loader.

It is intended to allow creating and opening wallets as well as managing
services like ticket buyer by RPC servers as well other subsystems.

This is a copy of the original upstream dcrwallet loader package, which was made
internal for the 1.5 release cycle.
*/
package loader
