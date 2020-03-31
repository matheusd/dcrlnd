// +build remotewallet,spv

package lntest

func useRemoteWallet() bool {
	return true
}

func useDcrwNode() bool {
	return false
}
