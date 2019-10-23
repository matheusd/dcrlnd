package netann

import (
	"fmt"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrec/secp256k1/v2"
	"github.com/decred/dcrlnd/lnwallet"
)

// NodeSigner is an implementation of the MessageSigner interface backed by the
// identity private key of running lnd node.
type NodeSigner struct {
	privKey *secp256k1.PrivateKey
}

// NewNodeSigner creates a new instance of the NodeSigner backed by the target
// private key.
func NewNodeSigner(key *secp256k1.PrivateKey) *NodeSigner {
	priv := &secp256k1.PrivateKey{}
	priv.Curve = secp256k1.S256()
	priv.PublicKey.X = key.X
	priv.PublicKey.Y = key.Y
	priv.D = key.D
	return &NodeSigner{
		privKey: priv,
	}
}

// SignMessage signs a chainhash digest of the passed msg under the
// resident node's private key. If the target public key is _not_ the node's
// private key, then an error will be returned.
func (n *NodeSigner) SignMessage(pubKey *secp256k1.PublicKey,
	msg []byte) (*secp256k1.Signature, error) {

	// If this isn't our identity public key, then we'll exit early with an
	// error as we can't sign with this key.
	if !pubKey.IsEqual((*secp256k1.PublicKey)(&n.privKey.PublicKey)) {
		return nil, fmt.Errorf("unknown public key")
	}

	// Otherwise, we'll sign the chainhash of the target message.
	digest := chainhash.HashB(msg)
	sign, err := n.privKey.Sign(digest)
	if err != nil {
		return nil, fmt.Errorf("can't sign the message: %v", err)
	}

	return sign, nil
}

// SignCompact signs a chainhash digest of the msg parameter under the
// resident node's private key. The returned signature is a pubkey-recoverable
// signature.
func (n *NodeSigner) SignCompact(msg []byte) ([]byte, error) {
	// We'll sign the chainhash of the target message.
	digest := chainhash.HashB(msg)

	return n.SignDigestCompact(digest)
}

// SignDigestCompact signs the provided message digest under the resident
// node's private key. The returned signature is a pubkey-recoverable signature.
func (n *NodeSigner) SignDigestCompact(hash []byte) ([]byte, error) {

	// Should the signature reference a compressed public key or not.
	isCompressedKey := true

	// secp256k1.SignCompact returns a pubkey-recoverable signature
	sig, err := secp256k1.SignCompact(n.privKey, hash, isCompressedKey)
	if err != nil {
		return nil, fmt.Errorf("can't sign the hash: %v", err)
	}

	return sig, nil
}

// A compile time check to ensure that NodeSigner implements the MessageSigner
// interface.
var _ lnwallet.MessageSigner = (*NodeSigner)(nil)
