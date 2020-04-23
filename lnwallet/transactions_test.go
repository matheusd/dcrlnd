package lnwallet

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/chaincfg/v2"
	"github.com/decred/dcrd/dcrec/secp256k1/v2"
	"github.com/decred/dcrd/dcrutil/v2"
	"github.com/decred/dcrd/txscript/v2"
	"github.com/decred/dcrd/wire"
	"github.com/decred/dcrlnd/channeldb"
	"github.com/decred/dcrlnd/input"
	"github.com/decred/dcrlnd/keychain"
	"github.com/decred/dcrlnd/lnwallet/chainfee"
	"github.com/decred/dcrlnd/lnwire"
	"github.com/decred/dcrlnd/shachain"
)

/**
* This file implements that different types of transactions used in the
* lightning protocol are created correctly. To do so, the tests use the test
* vectors defined in Appendix B & C of BOLT 03.
 */

// testContext contains the test parameters defined in Appendix B & C of the
// BOLT 03 spec.
type testContext struct {
	netParams *chaincfg.Params
	block1    *dcrutil.Block

	fundingInputPrivKey *secp256k1.PrivateKey
	localFundingPrivKey *secp256k1.PrivateKey
	localPaymentPrivKey *secp256k1.PrivateKey

	remoteFundingPubKey    *secp256k1.PublicKey
	localFundingPubKey     *secp256k1.PublicKey
	localRevocationPubKey  *secp256k1.PublicKey
	localPaymentPubKey     *secp256k1.PublicKey
	remotePaymentPubKey    *secp256k1.PublicKey
	localDelayPubKey       *secp256k1.PublicKey
	commitmentPoint        *secp256k1.PublicKey
	localPaymentBasePoint  *secp256k1.PublicKey
	remotePaymentBasePoint *secp256k1.PublicKey

	fundingChangeAddress dcrutil.Address
	fundingInputUtxo     *Utxo
	fundingInputTxOut    *wire.TxOut
	fundingTx            *dcrutil.Tx
	fundingOutpoint      wire.OutPoint
	shortChanID          lnwire.ShortChannelID

	htlcs []channeldb.HTLC

	localCsvDelay uint16
	fundingAmount dcrutil.Amount
	dustLimit     dcrutil.Amount
	FeePerKB      dcrutil.Amount
}

// htlcDesc is a description used to construct each HTLC in each test case.
type htlcDesc struct {
	index           int
	remoteSigHex    string
	resolutionTxHex string
}

// getHTLC constructs an HTLC based on a configured HTLC with auxiliary data
// such as the remote signature from the htlcDesc. The partially defined HTLCs
// originate from the BOLT 03 spec and are contained in the test context.
func (tc *testContext) getHTLC(index int, desc *htlcDesc) (channeldb.HTLC, error) {
	signature, err := hex.DecodeString(desc.remoteSigHex)
	if err != nil {
		return channeldb.HTLC{}, fmt.Errorf(
			"Failed to parse serialized signature: %v", err)
	}

	htlc := tc.htlcs[desc.index]
	return channeldb.HTLC{
		Signature:     signature,
		RHash:         htlc.RHash,
		RefundTimeout: htlc.RefundTimeout,
		Amt:           htlc.Amt,
		OutputIndex:   int32(index),
		Incoming:      htlc.Incoming,
	}, nil
}

// newTestContext populates a new testContext struct with the constant
// parameters defined in the BOLT 03 spec. This may return an error if any of
// the serialized parameters cannot be parsed.
func newTestContext() (tc *testContext, err error) {
	tc = new(testContext)

	const genesisHash = "2ced94b4ae95bba344cfa043268732d230649c640f92dce2d9518823d3057cb0"
	if tc.netParams, err = tc.createNetParams(genesisHash); err != nil {
		return
	}

	const block1Hex = "06000000b07c05d3238851d9e2dc920f649c6430d232872643a0cf44a3bb95aeb494ed2c841706fbd326fb3584ac6ed18f2899189c2d20bae5031932dd8a004300c2dabc000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000ffff7f20204e0000000000000100000073010000e94c485c01000000cf1e86eb445c26a0000000000000000000000000000000000000000000000000000000000101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff00ffffffff0300a0724e1809000000001976a9147e4765ae88ba9ad5c9e4715c484e90b34d358d5188ac00a0724e1809000000001976a91402fb1ac0137666d79165e13cecd403883615270788ac00a0724e1809000000001976a91469de627d3231b14228653dd09cba75eeb872754288ac00000000000000000100e057eb481b000000000000ffffffff0800002f646372642f00"
	if tc.block1, err = blockFromHex(block1Hex); err != nil {
		err = fmt.Errorf("Failed to parse serialized block: %v", err)
		return
	}

	// Key for decred's BlockOneLedgerRegNet address RsKrWb7Vny1jnzL1sDLgKTAteh9RZcRr5g6
	const fundingInputPrivKeyHex = "fd79250838efa1c142e182d012004c541df2668014cc1758027d70069b2ef474"
	tc.fundingInputPrivKey, err = privkeyFromHex(fundingInputPrivKeyHex)
	if err != nil {
		err = fmt.Errorf("Failed to parse serialized privkey: %v", err)
		return
	}

	const localFundingPrivKeyHex = "30ff4956bbdd3222d44cc5e8a1261dab1e07957bdac5ae88fe3261ef321f3749"
	tc.localFundingPrivKey, err = privkeyFromHex(localFundingPrivKeyHex)
	if err != nil {
		err = fmt.Errorf("Failed to parse serialized privkey: %v", err)
		return
	}

	const localPaymentPrivKeyHex = "bb13b121cdc357cd2e608b0aea294afca36e2b34cf958e2e6451a2f274694491"
	tc.localPaymentPrivKey, err = privkeyFromHex(localPaymentPrivKeyHex)
	if err != nil {
		err = fmt.Errorf("Failed to parse serialized privkey: %v", err)
		return
	}

	const localFundingPubKeyHex = "023da092f6980e58d2c037173180e9a465476026ee50f96695963e8efe436f54eb"
	tc.localFundingPubKey, err = pubkeyFromHex(localFundingPubKeyHex)
	if err != nil {
		err = fmt.Errorf("Failed to parse serialized pubkey: %v", err)
		return
	}

	// Sanity check.
	if !tc.localFundingPrivKey.PubKey().IsEqual(tc.localFundingPubKey) {
		err = fmt.Errorf("Pubkey of localFundingPrivKey not the same as encoded")
		return
	}

	// Corresponding private key: 11796dc04db0bd5858cfd9aa109e0b8f83039dbf2080520ea6df906802feb06f
	const remoteFundingPubKeyHex = "037f6fc2b0e0d63fab64424ef15991dfb76151ba0cf60dd2992015bf81b2eafdd5"
	tc.remoteFundingPubKey, err = pubkeyFromHex(remoteFundingPubKeyHex)
	if err != nil {
		err = fmt.Errorf("Failed to parse serialized pubkey: %v", err)
		return
	}

	const localRevocationPubKeyHex = "0212a140cd0c6539d07cd08dfe09984dec3251ea808b892efeac3ede9402bf2b19"
	tc.localRevocationPubKey, err = pubkeyFromHex(localRevocationPubKeyHex)
	if err != nil {
		err = fmt.Errorf("Failed to parse serialized pubkey: %v", err)
		return
	}

	const localPaymentPubKeyHex = "030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e7"
	tc.localPaymentPubKey, err = pubkeyFromHex(localPaymentPubKeyHex)
	if err != nil {
		err = fmt.Errorf("Failed to parse serialized pubkey: %v", err)
		return
	}

	// Corresponding private key: a914caedc83de8579ddc703de39b99b47fcb806f2e6987
	const remotePaymentPubKeyHex = "034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed9"
	tc.remotePaymentPubKey, err = pubkeyFromHex(remotePaymentPubKeyHex)
	if err != nil {
		err = fmt.Errorf("Failed to parse serialized pubkey: %v", err)
		return
	}

	const localDelayPubKeyHex = "03fd5960528dc152014952efdb702a88f71e3c1653b2314431701ec77e57fde83c"
	tc.localDelayPubKey, err = pubkeyFromHex(localDelayPubKeyHex)
	if err != nil {
		err = fmt.Errorf("Failed to parse serialized pubkey: %v", err)
		return
	}

	const commitmentPointHex = "025f7117a78150fe2ef97db7cfc83bd57b2e2c0d0dd25eaf467a4a1c2a45ce1486"
	tc.commitmentPoint, err = pubkeyFromHex(commitmentPointHex)
	if err != nil {
		err = fmt.Errorf("Failed to parse serialized pubkey: %v", err)
		return
	}

	const localPaymentBasePointHex = "034f355bdcb7cc0af728ef3cceb9615d90684bb5b2ca5f859ab0f0b704075871aa"
	tc.localPaymentBasePoint, err = pubkeyFromHex(localPaymentBasePointHex)
	if err != nil {
		err = fmt.Errorf("Failed to parse serialized pubkey: %v", err)
		return
	}

	const remotePaymentBasePointHex = "032c0b7cf95324a07d05398b240174dc0c2be444d96b159aa6c7f7b1e668680991"
	tc.remotePaymentBasePoint, err = pubkeyFromHex(remotePaymentBasePointHex)
	if err != nil {
		err = fmt.Errorf("Failed to parse serialized pubkey: %v", err)
		return
	}

	const fundingChangeAddressStr = "Rs8LovBHZfZmC4ShUicmExNWaivPm5cBtNN"
	tc.fundingChangeAddress, err = dcrutil.DecodeAddress(
		fundingChangeAddressStr, tc.netParams)
	if err != nil {
		err = fmt.Errorf("Failed to parse address: %v", err)
		return
	}

	tc.fundingInputUtxo, tc.fundingInputTxOut, err = tc.extractFundingInput()
	if err != nil {
		return
	}

	const fundingTxHex = "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff00ffffffff0300a0724e1809000000001976a9147e4765ae88ba9ad5c9e4715c484e90b34d358d5188ac00a0724e1809000000001976a91402fb1ac0137666d79165e13cecd403883615270788ac00a0724e1809000000001976a91469de627d3231b14228653dd09cba75eeb872754288ac00000000000000000100e057eb481b000000000000ffffffff0800002f646372642f"
	if tc.fundingTx, err = txFromHex(fundingTxHex); err != nil {
		err = fmt.Errorf("Failed to parse serialized tx: %v", err)
		return
	}

	tc.fundingOutpoint = wire.OutPoint{
		Hash:  *tc.fundingTx.Hash(),
		Index: 0,
	}

	tc.shortChanID = lnwire.ShortChannelID{
		BlockHeight: 1,
		TxIndex:     0,
		TxPosition:  0,
	}

	htlcData := []struct {
		incoming bool
		amount   lnwire.MilliAtom
		expiry   uint32
		preimage string
	}{
		{
			incoming: true,
			amount:   1000000,
			expiry:   500,
			preimage: "0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			incoming: true,
			amount:   2000000,
			expiry:   501,
			preimage: "0101010101010101010101010101010101010101010101010101010101010101",
		},
		{
			incoming: false,
			amount:   2000000,
			expiry:   502,
			preimage: "0202020202020202020202020202020202020202020202020202020202020202",
		},
		{
			incoming: false,
			amount:   3000000,
			expiry:   503,
			preimage: "0303030303030303030303030303030303030303030303030303030303030303",
		},
		{
			incoming: true,
			amount:   4000000,
			expiry:   504,
			preimage: "0404040404040404040404040404040404040404040404040404040404040404",
		},
	}

	tc.htlcs = make([]channeldb.HTLC, len(htlcData))
	for i, htlc := range htlcData {
		preimage, decodeErr := hex.DecodeString(htlc.preimage)
		if decodeErr != nil {
			err = fmt.Errorf("Failed to decode HTLC preimage: %v", decodeErr)
			return
		}

		tc.htlcs[i].RHash = sha256.Sum256(preimage)
		tc.htlcs[i].Amt = htlc.amount
		tc.htlcs[i].RefundTimeout = htlc.expiry
		tc.htlcs[i].Incoming = htlc.incoming
	}

	tc.localCsvDelay = 144
	tc.fundingAmount = 10000000
	tc.dustLimit = 546
	tc.FeePerKB = 15000

	return
}

// createNetParams is used by newTestContext to construct new chain parameters
// as required by the BOLT 03 spec.
func (tc *testContext) createNetParams(genesisHashStr string) (*chaincfg.Params, error) {
	params := chaincfg.RegNetParams()

	// Ensure regression net genesis block matches the one listed in BOLT spec.
	expectedGenesisHash, err := chainhash.NewHashFromStr(genesisHashStr)
	if err != nil {
		return nil, err
	}
	if !params.GenesisHash.IsEqual(expectedGenesisHash) {
		err = fmt.Errorf("Expected regression net genesis hash to be %s, "+
			"got %s", expectedGenesisHash, params.GenesisHash)
		return nil, err
	}

	return params, nil
}

// extractFundingInput returns references to the transaction output of the
// coinbase transaction which is used to fund the channel in the test vectors.
func (tc *testContext) extractFundingInput() (*Utxo, *wire.TxOut, error) {
	expectedTxHashHex := "f94a4d807512f45842661cc7fdefebafd54cbc1338655bccd8e9821d06e402c4"
	expectedTxHash, err := chainhash.NewHashFromStr(expectedTxHashHex)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to parse transaction hash: %v", err)
	}

	tx, err := tc.block1.Tx(0)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get coinbase transaction from "+
			"block 1: %v", err)
	}
	txout := tx.MsgTx().TxOut[0]

	var expectedAmount int64 = 10000000000000
	if txout.Value != expectedAmount {
		return nil, nil, fmt.Errorf("Coinbase transaction output amount from "+
			"block 1 does not match expected output amount: "+
			"expected %v, got %v", expectedAmount, txout.Value)
	}
	if !tx.Hash().IsEqual(expectedTxHash) {
		return nil, nil, fmt.Errorf("Coinbase transaction hash from block 1 "+
			"does not match expected hash: expected %v, got %v", expectedTxHash,
			tx.Hash())
	}

	block1Utxo := Utxo{
		AddressType: PubKeyHash,
		Value:       dcrutil.Amount(txout.Value),
		OutPoint: wire.OutPoint{
			Hash:  *tx.Hash(),
			Index: 0,
		},
		PkScript: txout.PkScript,
	}
	return &block1Utxo, txout, nil
}

// TestCommitmentAndHTLCTransactions checks the test vectors specified in
// BOLT 03, Appendix C. This deterministically generates commitment and second
// level HTLC transactions and checks that they match the expected values.
func TestCommitmentAndHTLCTransactions(t *testing.T) {
	t.Parallel()
	t.Skip("FIXME: re-enable after upstram #3821")

	tc, err := newTestContext()
	if err != nil {
		t.Fatal(err)
	}

	// Generate random some keys that don't actually matter but need to be set.
	var (
		identityKey         *secp256k1.PublicKey
		localDelayBasePoint *secp256k1.PublicKey
	)
	generateKeys := []**secp256k1.PublicKey{
		&identityKey,
		&localDelayBasePoint,
	}
	for _, keyRef := range generateKeys {
		privkey, err := secp256k1.GeneratePrivateKey()
		if err != nil {
			t.Fatalf("Failed to generate new key: %v", err)
		}
		*keyRef = privkey.PubKey()
	}

	// Manually construct a new LightningChannel.
	channelState := channeldb.OpenChannel{
		ChanType:        channeldb.SingleFunderTweaklessBit,
		ChainHash:       tc.netParams.GenesisHash,
		FundingOutpoint: tc.fundingOutpoint,
		ShortChannelID:  tc.shortChanID,
		IsInitiator:     true,
		IdentityPub:     identityKey,
		LocalChanCfg: channeldb.ChannelConfig{
			ChannelConstraints: channeldb.ChannelConstraints{
				DustLimit:        tc.dustLimit,
				MaxPendingAmount: lnwire.NewMAtomsFromAtoms(tc.fundingAmount),
				MaxAcceptedHtlcs: input.MaxHTLCNumber,
				CsvDelay:         tc.localCsvDelay,
			},
			MultiSigKey: keychain.KeyDescriptor{
				PubKey: tc.localFundingPubKey,
			},
			PaymentBasePoint: keychain.KeyDescriptor{
				PubKey: tc.localPaymentBasePoint,
			},
			HtlcBasePoint: keychain.KeyDescriptor{
				PubKey: tc.localPaymentBasePoint,
			},
			DelayBasePoint: keychain.KeyDescriptor{
				PubKey: localDelayBasePoint,
			},
		},
		RemoteChanCfg: channeldb.ChannelConfig{
			MultiSigKey: keychain.KeyDescriptor{
				PubKey: tc.remoteFundingPubKey,
			},
			PaymentBasePoint: keychain.KeyDescriptor{
				PubKey: tc.remotePaymentBasePoint,
			},
			HtlcBasePoint: keychain.KeyDescriptor{
				PubKey: tc.remotePaymentBasePoint,
			},
		},
		Capacity:           tc.fundingAmount,
		RevocationProducer: shachain.NewRevocationProducer(shachain.ShaHash(zeroHash)),
	}
	signer := &input.MockSigner{
		Privkeys: []*secp256k1.PrivateKey{
			tc.localFundingPrivKey, tc.localPaymentPrivKey,
		},
		NetParams: tc.netParams,
	}

	// Construct a LightningChannel manually because we don't have nor need all
	// of the dependencies.
	channel := LightningChannel{
		channelState:  &channelState,
		Signer:        signer,
		commitBuilder: NewCommitmentBuilder(&channelState, tc.netParams),
		netParams:     tc.netParams,
	}
	err = channel.createSignDesc()
	if err != nil {
		t.Fatalf("Failed to generate channel sign descriptor: %v", err)
	}

	// The commitmentPoint is technically hidden in the spec, but we need it to
	// generate the correct tweak.
	tweak := input.SingleTweakBytes(tc.commitmentPoint, tc.localPaymentBasePoint)
	keys := &CommitmentKeyRing{
		CommitPoint:         tc.commitmentPoint,
		LocalCommitKeyTweak: tweak,
		LocalHtlcKeyTweak:   tweak,
		LocalHtlcKey:        tc.localPaymentPubKey,
		RemoteHtlcKey:       tc.remotePaymentPubKey,
		ToLocalKey:          tc.localDelayPubKey,
		ToRemoteKey:         tc.remotePaymentPubKey,
		RevocationKey:       tc.localRevocationPubKey,
	}

	// testCases encode the raw test vectors specified in Appendix C of BOLT 03.
	// TODO(decred) The stored hex txs need to be reviewed and documented somewhere.
	testCases := []struct {
		commitment              channeldb.ChannelCommitment
		htlcDescs               []htlcDesc
		expectedCommitmentTxHex string
		remoteSigHex            string
	}{
		{ // 0
			commitment: channeldb.ChannelCommitment{
				CommitHeight:  42,
				LocalBalance:  7000000000,
				RemoteBalance: 3000000000,
				FeePerKB:      15000,
			},
			htlcDescs:               []htlcDesc{},
			expectedCommitmentTxHex: "0200000001c402e4061d82e9d8cc5b653813bc4cd5afebeffdc71c664258f41275804d4af9000000000038b02b8002c0c62d000000000000001976a914f682e6254058108032264237a8e62b4777400d4e88ac6cba6a0000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda0873e1952200000000001000000000000000000000000ffffffffd847304402204005c5116474be52f7f40192583d258e771727cf8a6f2f301bc53d7a12bfcd0b0220427ab29fc5807b938a3356f591e1f876ef65f18a2b713de1277f2156c5f3fe68014730440220774e8670eab7b7ded6e37fad1d3ce74ed8278b47a522ddd7a0f5de2aa74519aa02206f628efb5acb9242f6ee725ee3e14eb916f2c9294ff87f439575b1d79441891b01475221023da092f6980e58d2c037173180e9a465476026ee50f96695963e8efe436f54eb21037f6fc2b0e0d63fab64424ef15991dfb76151ba0cf60dd2992015bf81b2eafdd552ae",
			remoteSigHex:            "30440220774e8670eab7b7ded6e37fad1d3ce74ed8278b47a522ddd7a0f5de2aa74519aa02206f628efb5acb9242f6ee725ee3e14eb916f2c9294ff87f439575b1d79441891b",
		},
		{ // 1
			commitment: channeldb.ChannelCommitment{
				CommitHeight:  42,
				LocalBalance:  6988000000,
				RemoteBalance: 3000000000,
				FeePerKB:      0,
			},
			htlcDescs: []htlcDesc{
				{ // 1,0
					index:           0,
					remoteSigHex:    "304402203fe90dac888f202fbc5290c846a59682dda14593046c0e9cf2fc04028eacdbb40220382b1d6fe8eb6ef36d7311c26de864fdc529aee4e450f700d70e5a6dc20cbef9",
					resolutionTxHex: "02000000014c1170559172b82c493f1907d9f15683c0b744a830b5bd9e7846355a3bfe3e3b01000000000000000001f044000000000000000017a914d5e0af0295b6d842b7ec7abc5f99a3241285e3f18705000000000000000100000000000000000000000000000000fd1b01483045022100b34bc2b0822e793f2a5b11b1952b87b27530b63ba20c413e530bc231e95a9f3a022057d9d6f475d60d808ceb8fff342a44146bbc56c9722d67e3f4a7a286c13e3d2801483045022100d5d6b5385644092787a10883601cfa9e89d6eee9185ac2541f28a7818d73702d02205d4cd84123a4382ad32b20a602a5fb6aac1573040f4083973b7b7135eacf808701004c8676a914fbd88d3dd1237b3b0bbc3d310485d705ceb601cf8763ac6721031e4b7016ae2448401bb818f5685273f87a7e38a4f0278a550fd1596a370ad00d7c820120876475527c210375f3b13dac95e0625018e75ee130e2e3b74e7a0bf496b30913fe9ae154112d6e52ae67c0a614b8bcb07f6344b42ab04250c86a6e8b75d3fdbbc688ac6868",
				},
				{ // 1,2
					index:           1,
					remoteSigHex:    "304402207b2e085a7af0975dfbb1ff091bd0d7c28f2f201edafef2fa53d7ee4e7fbf72df022068e9fd876bf8e2a537943b6acd62e8b3668c5f12973ef30423d1e8796a414db8",
					resolutionTxHex: "02000000014c1170559172b82c493f1907d9f15683c0b744a830b5bd9e7846355a3bfe3e3b01000000000000000001f044000000000000000017a914d5e0af0295b6d842b7ec7abc5f99a3241285e3f18705000000000000000100000000000000000000000000000000fd1b01483045022100b34bc2b0822e793f2a5b11b1952b87b27530b63ba20c413e530bc231e95a9f3a022057d9d6f475d60d808ceb8fff342a44146bbc56c9722d67e3f4a7a286c13e3d2801483045022100d5d6b5385644092787a10883601cfa9e89d6eee9185ac2541f28a7818d73702d02205d4cd84123a4382ad32b20a602a5fb6aac1573040f4083973b7b7135eacf808701004c8676a914fbd88d3dd1237b3b0bbc3d310485d705ceb601cf8763ac6721031e4b7016ae2448401bb818f5685273f87a7e38a4f0278a550fd1596a370ad00d7c820120876475527c210375f3b13dac95e0625018e75ee130e2e3b74e7a0bf496b30913fe9ae154112d6e52ae67c0a614b8bcb07f6344b42ab04250c86a6e8b75d3fdbbc688ac6868",
				},
				{ // 1,1
					index:           2,
					remoteSigHex:    "304402207b2e085a7af0975dfbb1ff091bd0d7c28f2f201edafef2fa53d7ee4e7fbf72df022068e9fd876bf8e2a537943b6acd62e8b3668c5f12973ef30423d1e8796a414db8",
					resolutionTxHex: "02000000010ed8272a5f0f27af7a19edd6d21e3b177d47d995c1d246f206278891bd332e6d02000000000000000001d007000000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda087f6010000000000000100000000000000000000000000000000fd1a0147304402207b2e085a7af0975dfbb1ff091bd0d7c28f2f201edafef2fa53d7ee4e7fbf72df022068e9fd876bf8e2a537943b6acd62e8b3668c5f12973ef30423d1e8796a414db801483045022100a615f98d9e38309865286554bcfbdd27b192e8838b74e7bac6635b30c7a9f4d0022065248ccf5a13d5f9f72da1b6f47a2e169b17f8fe4f9d84fbff056d7834a9da6601004c8676a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac6721034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed97c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67c0a614b43e1b38138a41b37f7cd9a1d274bc63e3a9b5d188ac6868",
				},
				{ // 1,3
					index:           3,
					remoteSigHex:    "3044022066a641fecac3329596851cf4bce1152688d54c4ec52b8c87cdf2709d08817e940220498976fa5ac58c7bc77b63d10cf6ad66ff0818cf71d0851b67e78df3e46e5edb",
					resolutionTxHex: "02000000010ed8272a5f0f27af7a19edd6d21e3b177d47d995c1d246f206278891bd332e6d03000000000000000001b80b000000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda087f7010000000000000100000000000000000000000000000000fd1a01473044022066a641fecac3329596851cf4bce1152688d54c4ec52b8c87cdf2709d08817e940220498976fa5ac58c7bc77b63d10cf6ad66ff0818cf71d0851b67e78df3e46e5edb01483045022100d5d0753a3d927b67cc8b0c854066f358112c00fcc775d7230e1be4c22c6c7d2302200d9c332b2d55598e624dbb5a44d9970e096851d32417b978638bd0488a3f019801004c8676a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac6721034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed97c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67c0a6148a486ff2e31d6158bf39e2608864d63fefd09d5b88ac6868",
				},
				{ // 1,4
					index:           4,
					remoteSigHex:    "304402207e0410e45454b0978a623f36a10626ef17b27d9ad44e2760f98cfa3efb37924f0220220bd8acd43ecaa916a80bd4f919c495a2c58982ce7c8625153f8596692a801d",
					resolutionTxHex: "020000000001018154ecccf11a5fb56c39654c4deb4d2296f83c69268280b94d021370c94e219704000000000000000001a00f0000000000002200204adb4e2f00643db396dd120d4e7dc17625f5f2c11a40d857accc862d6b7dd80e050047304402207e0410e45454b0978a623f36a10626ef17b27d9ad44e2760f98cfa3efb37924f0220220bd8acd43ecaa916a80bd4f919c495a2c58982ce7c8625153f8596692a801d014730440220549e80b4496803cbc4a1d09d46df50109f546d43fbbf86cd90b174b1484acd5402205f12a4f995cb9bded597eabfee195a285986aa6d93ae5bb72507ebc6a4e2349e012004040404040404040404040404040404040404040404040404040404040404048a76a91414011f7254d96b819c76986c277d115efce6f7b58763ac67210394854aa6eab5b2a8122cc726e9dded053a2184d88256816826d6231c068d4a5b7c8201208763a91418bc1a114ccf9c052d3d23e28d3b0a9d1227434288527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae677502f801b175ac686800000000",
				},
			},
			expectedCommitmentTxHex: "0200000001c402e4061d82e9d8cc5b653813bc4cd5afebeffdc71c664258f41275804d4af9000000000038b02b8007e803000000000000000017a91466d30e8ad6532d2c2c39872452f68ee72db4293287d007000000000000000017a9143fbe957cc060bd7a34884fb2357648dd9e737e1787d007000000000000000017a914f4d6aadd25266f19107005e364cc9124b5009cca87b80b000000000000000017a914524799cd6e521eb575a038ed80bb03e03a91d2e187a00f000000000000000017a914cdbd3d056ef0e2c8f066a19ee85e52e70e07fe5b87c0c62d000000000000001976a914f682e6254058108032264237a8e62b4777400d4e88ace0a06a0000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda0873e1952200000000001000000000000000000000000ffffffffd8473044022035addd4509b976f54844cba64fb93e34260bf37a61231faa2fc0958d61944eaf0220240f44dbe4cc044ce2454a87763c4e74242d68514da4a02a703350bf3d2863d301473044022051216b75ef0104f9de22eeada13a43e65c37bbc8854860a2b0f151a066ac480302205a0152be3b401dc463f8a1c1d17ae74058c6e3b858a9c6a4063950071a3ebad001475221023da092f6980e58d2c037173180e9a465476026ee50f96695963e8efe436f54eb21037f6fc2b0e0d63fab64424ef15991dfb76151ba0cf60dd2992015bf81b2eafdd552ae",
			remoteSigHex:            "3044022051216b75ef0104f9de22eeada13a43e65c37bbc8854860a2b0f151a066ac480302205a0152be3b401dc463f8a1c1d17ae74058c6e3b858a9c6a4063950071a3ebad0",
		},
		{ // 2
			commitment: channeldb.ChannelCommitment{
				CommitHeight:  42,
				LocalBalance:  6988000000,
				RemoteBalance: 3000000000,
				FeePerKB:      647,
			},
			htlcDescs: []htlcDesc{
				{ // 2,0
					index:           0,
					remoteSigHex:    "30440220385a5afe75632f50128cbb029ee95c80156b5b4744beddc729ad339c9ca432c802202ba5f48550cad3379ac75b9b4fedb86a35baa6947f16ba5037fb8b11ab343740",
					resolutionTxHex: "020000000001018323148ce2419f21ca3d6780053747715832e18ac780931a514b187768882bb60000000000000000000122020000000000002200204adb4e2f00643db396dd120d4e7dc17625f5f2c11a40d857accc862d6b7dd80e05004730440220385a5afe75632f50128cbb029ee95c80156b5b4744beddc729ad339c9ca432c802202ba5f48550cad3379ac75b9b4fedb86a35baa6947f16ba5037fb8b11ab3437400147304402205999590b8a79fa346e003a68fd40366397119b2b0cdf37b149968d6bc6fbcc4702202b1e1fb5ab7864931caed4e732c359e0fe3d86a548b557be2246efb1708d579a012000000000000000000000000000000000000000000000000000000000000000008a76a91414011f7254d96b819c76986c277d115efce6f7b58763ac67210394854aa6eab5b2a8122cc726e9dded053a2184d88256816826d6231c068d4a5b7c8201208763a914b8bcb07f6344b42ab04250c86a6e8b75d3fdbbc688527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae677502f401b175ac686800000000",
				},
				{ // 2,1
					index:           1,
					remoteSigHex:    "3045022100e96f299ee8e605687ef7348b6cf9295d4e9ef2e20fb2473c14cab44f311742cc022053a24341b0856f5c467e560ef7ed83e0aa689d47aedbc24a04fc2636c553e796",
					resolutionTxHex: "0200000001db809e7bcd6e6a41ed4a74cc6a746fb3aec7b89a3a581b3a1a30fee4562313c501000000000000000001d306000000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda087f6010000000000000100000000000000000000000000000000fd1a01483045022100e96f299ee8e605687ef7348b6cf9295d4e9ef2e20fb2473c14cab44f311742cc022053a24341b0856f5c467e560ef7ed83e0aa689d47aedbc24a04fc2636c553e7960148304502210081407b282fc0979a669bdb98729c44fdeeabda8eb34d50cd4b5061b128b485f702206a26d87ef32b930a6ad0ddb67d9c6f9981abd9a25a0a143c8561fdde2746838b01004c8576a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac6721034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed97c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67a914dfd1c066155d9a82067f199d093e96903abee2ec88ac6868",
				},
				{ // 2,2
					index:           2,
					remoteSigHex:    "3045022100bd6b404539a2ce2c3bc2de8046207d46f78616902a4935a2a89a924d79055917022022aed9b70904de3026ce9fb926f2c9412e9731522303fe0ce9e3de05329fe792",
					resolutionTxHex: "0200000001162d4a6ab03ea621078b492bae9d0de14e026bc5733f5867646a6613c915a5a002000000000000000001d206000000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda087f6010000000000000100000000000000000000000000000000fd1a01483045022100bd6b404539a2ce2c3bc2de8046207d46f78616902a4935a2a89a924d79055917022022aed9b70904de3026ce9fb926f2c9412e9731522303fe0ce9e3de05329fe7920147304402202b2173786803162aeb4547e188dcb8957dd2f40513be801871488ea02036543702203f757754b92d5c8203b9ae28efbff67cb1d6850f858ae5f02f66f013ab409ba801004c8676a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac6721034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed97c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67c0a614b43e1b38138a41b37f7cd9a1d274bc63e3a9b5d188ac6868",
				},
				{ // 2,3
					index:           3,
					remoteSigHex:    "3045022100e2c39eebbde7791a85180cbf4ca81ac84b479e3fb213e59259b2c6048d41da49022048368d8669c6b5f07f74858e2d372b55f41ef108f8f5dd1bd35412a78fcbe62f",
					resolutionTxHex: "0200000001162d4a6ab03ea621078b492bae9d0de14e026bc5733f5867646a6613c915a5a003000000000000000001ba0a000000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda087f7010000000000000100000000000000000000000000000000fd1b01483045022100e2c39eebbde7791a85180cbf4ca81ac84b479e3fb213e59259b2c6048d41da49022048368d8669c6b5f07f74858e2d372b55f41ef108f8f5dd1bd35412a78fcbe62f014830450221009270083aba211660d97dde905b76050583fece447241d5281fb0dd3413c87c00022068f0fb5a780ca25a74d349474d829e5ce0c762ed9129691d3173e44f3b3fc13c01004c8676a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac6721034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed97c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67c0a6148a486ff2e31d6158bf39e2608864d63fefd09d5b88ac6868",
				},
				{ // 2,4
					index:           4,
					remoteSigHex:    "3045022100cc28030b59f0914f45b84caa983b6f8effa900c952310708c2b5b00781117022022027ba2ccdf94d03c6d48b327f183f6e28c8a214d089b9227f94ac4f85315274f0",
					resolutionTxHex: "020000000001018323148ce2419f21ca3d6780053747715832e18ac780931a514b187768882bb604000000000000000001da0d0000000000002200204adb4e2f00643db396dd120d4e7dc17625f5f2c11a40d857accc862d6b7dd80e0500483045022100cc28030b59f0914f45b84caa983b6f8effa900c952310708c2b5b00781117022022027ba2ccdf94d03c6d48b327f183f6e28c8a214d089b9227f94ac4f85315274f00147304402202d1a3c0d31200265d2a2def2753ead4959ae20b4083e19553acfffa5dfab60bf022020ede134149504e15b88ab261a066de49848411e15e70f9e6a5462aec2949f8f012004040404040404040404040404040404040404040404040404040404040404048a76a91414011f7254d96b819c76986c277d115efce6f7b58763ac67210394854aa6eab5b2a8122cc726e9dded053a2184d88256816826d6231c068d4a5b7c8201208763a91418bc1a114ccf9c052d3d23e28d3b0a9d1227434288527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae677502f801b175ac686800000000",
				},
			},
			expectedCommitmentTxHex: "0200000001c402e4061d82e9d8cc5b653813bc4cd5afebeffdc71c664258f41275804d4af9000000000038b02b8007e803000000000000000017a91466d30e8ad6532d2c2c39872452f68ee72db4293287d007000000000000000017a9143fbe957cc060bd7a34884fb2357648dd9e737e1787d007000000000000000017a914f4d6aadd25266f19107005e364cc9124b5009cca87b80b000000000000000017a914524799cd6e521eb575a038ed80bb03e03a91d2e187a00f000000000000000017a914cdbd3d056ef0e2c8f066a19ee85e52e70e07fe5b87c0c62d000000000000001976a914f682e6254058108032264237a8e62b4777400d4e88ac879f6a0000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda0873e1952200000000001000000000000000000000000ffffffffd9483045022100ae216321ba047aa96a0df56c0ff450b21674bd271d76c5ee8aedd690f859c8c4022043b7c2ae9118b23e83ca1ed0055a02932c5e3f4a994b880ef89f064a2f045637014730440220172d9ffaf5c4bb2348124402e7ff4deb013c70039e184ea4d845f7a8fda5dac402200ca255329170e1ee633670837019ab2cd36098f7f91d4bc7de1ccd6ec23e3e1d01475221023da092f6980e58d2c037173180e9a465476026ee50f96695963e8efe436f54eb21037f6fc2b0e0d63fab64424ef15991dfb76151ba0cf60dd2992015bf81b2eafdd552ae",
			remoteSigHex:            "30440220172d9ffaf5c4bb2348124402e7ff4deb013c70039e184ea4d845f7a8fda5dac402200ca255329170e1ee633670837019ab2cd36098f7f91d4bc7de1ccd6ec23e3e1d",
		},
		{ // 3
			commitment: channeldb.ChannelCommitment{
				CommitHeight:  42,
				LocalBalance:  6988000000,
				RemoteBalance: 3000000000,
				FeePerKB:      648,
			},
			htlcDescs: []htlcDesc{
				{ // 3,0
					index:           1,
					remoteSigHex:    "3045022100d1f772a0eb536d4acf3afec9f6ec4e06f14a4e104b9bac6822790fa6c290b13a02207f7f35e3f1441e45d448d84ca27c3cf39632d6d1c979758f16b5d5c310bfe032",
					resolutionTxHex: "0200000001227c54c58b9cd185a35d9c8cfab40d27e9641669d6166c4365822d8c319fd6ef01000000000000000001d206000000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda087f6010000000000000100000000000000000000000000000000fd1b01483045022100d1f772a0eb536d4acf3afec9f6ec4e06f14a4e104b9bac6822790fa6c290b13a02207f7f35e3f1441e45d448d84ca27c3cf39632d6d1c979758f16b5d5c310bfe0320148304502210099370b1dfe73c9431972b8d0be685ba2984fd5f52c1e7c1502250b2797d5a8e402200fef13b9b3fafc71386a5444a39518e9c646e6165b03aba530bc943872c5669601004c8676a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac6721034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed97c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67c0a614b43e1b38138a41b37f7cd9a1d274bc63e3a9b5d188ac6868",
				},
				{ // 3,1
					index:           2,
					remoteSigHex:    "3045022100d1f772a0eb536d4acf3afec9f6ec4e06f14a4e104b9bac6822790fa6c290b13a02207f7f35e3f1441e45d448d84ca27c3cf39632d6d1c979758f16b5d5c310bfe032",
					resolutionTxHex: "0200000001227c54c58b9cd185a35d9c8cfab40d27e9641669d6166c4365822d8c319fd6ef01000000000000000001d206000000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda087f6010000000000000100000000000000000000000000000000fd1b01483045022100d1f772a0eb536d4acf3afec9f6ec4e06f14a4e104b9bac6822790fa6c290b13a02207f7f35e3f1441e45d448d84ca27c3cf39632d6d1c979758f16b5d5c310bfe0320148304502210099370b1dfe73c9431972b8d0be685ba2984fd5f52c1e7c1502250b2797d5a8e402200fef13b9b3fafc71386a5444a39518e9c646e6165b03aba530bc943872c5669601004c8676a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac6721034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed97c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67c0a614b43e1b38138a41b37f7cd9a1d274bc63e3a9b5d188ac6868",
				},
				{ // 3,2
					index:           3,
					remoteSigHex:    "3045022100c031f8909f95c341946adb43831ebed4ef65fb41c7fdbe9d7e2a5b2316a5254e0220633182e63c2cb762a11e7de70618603b6e64d675d260f0ce798af3977a53edbc",
					resolutionTxHex: "0200000001227c54c58b9cd185a35d9c8cfab40d27e9641669d6166c4365822d8c319fd6ef02000000000000000001ba0a000000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda087f7010000000000000100000000000000000000000000000000fd1a01483045022100c031f8909f95c341946adb43831ebed4ef65fb41c7fdbe9d7e2a5b2316a5254e0220633182e63c2cb762a11e7de70618603b6e64d675d260f0ce798af3977a53edbc01473044022037c212785e66e81948cb8e5bd829b5d1a8b9ca89ce33cfe3de765901b99ccdbb02207fb703e06f322c8d66513be0cb53d81aabe36ed2dae6c5463d08436b0aaf3d9c01004c8676a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac6721034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed97c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67c0a6148a486ff2e31d6158bf39e2608864d63fefd09d5b88ac6868",
				},
				{ // 3,3
					index:           4,
					remoteSigHex:    "3044022035cac88040a5bba420b1c4257235d5015309113460bc33f2853cd81ca36e632402202fc94fd3e81e9d34a9d01782a0284f3044370d03d60f3fc041e2da088d2de58f",
					resolutionTxHex: "02000000000101579c183eca9e8236a5d7f5dcd79cfec32c497fdc0ec61533cde99ecd436cadd103000000000000000001d90d0000000000002200204adb4e2f00643db396dd120d4e7dc17625f5f2c11a40d857accc862d6b7dd80e0500473044022035cac88040a5bba420b1c4257235d5015309113460bc33f2853cd81ca36e632402202fc94fd3e81e9d34a9d01782a0284f3044370d03d60f3fc041e2da088d2de58f0147304402200daf2eb7afd355b4caf6fb08387b5f031940ea29d1a9f35071288a839c9039e4022067201b562456e7948616c13acb876b386b511599b58ac1d94d127f91c50463a6012004040404040404040404040404040404040404040404040404040404040404048a76a91414011f7254d96b819c76986c277d115efce6f7b58763ac67210394854aa6eab5b2a8122cc726e9dded053a2184d88256816826d6231c068d4a5b7c8201208763a91418bc1a114ccf9c052d3d23e28d3b0a9d1227434288527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae677502f801b175ac686800000000",
				},
			},
			expectedCommitmentTxHex: "0200000001c402e4061d82e9d8cc5b653813bc4cd5afebeffdc71c664258f41275804d4af9000000000038b02b8006d007000000000000000017a9143fbe957cc060bd7a34884fb2357648dd9e737e1787d007000000000000000017a914f4d6aadd25266f19107005e364cc9124b5009cca87b80b000000000000000017a914524799cd6e521eb575a038ed80bb03e03a91d2e187a00f000000000000000017a914cdbd3d056ef0e2c8f066a19ee85e52e70e07fe5b87c0c62d000000000000001976a914f682e6254058108032264237a8e62b4777400d4e88ac9c9f6a0000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda0873e1952200000000001000000000000000000000000ffffffffda483045022100f509a959b0f6ab95ba0ce159621d67a7b4c392e00b6ac2229e437c3a7450c56e02202df68104320f4abf23354e33a8e219230625ed011fad6a1165aa07d7e203645101483045022100da36e9418230cf09fe307efd15d512bb3d8b8967a813cd59dd0db06757be88350220456375f3bca91d9e23ca28adec9b87a2669656041f486595e1d94cc882fc572901475221023da092f6980e58d2c037173180e9a465476026ee50f96695963e8efe436f54eb21037f6fc2b0e0d63fab64424ef15991dfb76151ba0cf60dd2992015bf81b2eafdd552ae",
			remoteSigHex:            "3045022100da36e9418230cf09fe307efd15d512bb3d8b8967a813cd59dd0db06757be88350220456375f3bca91d9e23ca28adec9b87a2669656041f486595e1d94cc882fc5729",
		},
		{ // 4
			commitment: channeldb.ChannelCommitment{
				CommitHeight:  42,
				LocalBalance:  6988000000,
				RemoteBalance: 3000000000,
				FeePerKB:      2069,
			},
			htlcDescs: []htlcDesc{
				{ // 4,0
					index:           1,
					remoteSigHex:    "304402201862a6d4e965ea0163b6b863a59b56bc4e8afc75914ac03e8dd21b2a5077049c02204b53b2c23a9d99ff8b150ed44d9f0c9257d01cfaaa35420436a4b9229742e90d",
					resolutionTxHex: "020000000108724b58ec47dd8bdf1c2bbeb631506676c444c379453a69364a7029a879b88500000000000000000001a504000000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda087f6010000000000000100000000000000000000000000000000fd180147304402201862a6d4e965ea0163b6b863a59b56bc4e8afc75914ac03e8dd21b2a5077049c02204b53b2c23a9d99ff8b150ed44d9f0c9257d01cfaaa35420436a4b9229742e90d0147304402206edc444a6d7616321cdc3652d2567477f7358c11c438ae063111f90b47bfef7a02205ab52360caacadc003e78e48ff3ce2b215b666f842f96c857b89996f1e62f93901004c8576a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac6721034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed97c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67a914dfd1c066155d9a82067f199d093e96903abee2ec88ac6868",
				},
				{ // 4,1
					index:           2,
					remoteSigHex:    "304402205ff6844933eb9635f22151e60faae5d075d3eccad096a54c32edbc177bab37c6022012649a24289b878055af38d26b03a7bc46703fbdfde81d0011e14eeb0596f978",
					resolutionTxHex: "02000000017cf51e2837dc108f405575c1b6c7be9f1d3f808c6b4a349436c06b5bfc5460ac01000000000000000001a304000000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda087f6010000000000000100000000000000000000000000000000fd1a0147304402205ff6844933eb9635f22151e60faae5d075d3eccad096a54c32edbc177bab37c6022012649a24289b878055af38d26b03a7bc46703fbdfde81d0011e14eeb0596f9780148304502210085e33ca252f46ff6aa4271ccc2dfd36c930386728b0aacb65e8c259caa5d1ca802207fb07582017d57c52aee5e01e4639acfbf8c2af3d675b1e4bb8ae9a1242ad7a601004c8676a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac6721034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed97c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67c0a614b43e1b38138a41b37f7cd9a1d274bc63e3a9b5d188ac6868",
				},
				{ // 4,2
					index:           3,
					remoteSigHex:    "3044022044a2b66745adddbdc56dc93e330977c9d8d205d25b5fec91dc9a8fa37f16be7902202705ab46f2d835f86e80fc5538e5b9f4c824078acf6a511fba52f3c71669955e",
					resolutionTxHex: "02000000017cf51e2837dc108f405575c1b6c7be9f1d3f808c6b4a349436c06b5bfc5460ac020000000000000000018b08000000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda087f7010000000000000100000000000000000000000000000000fd1901473044022044a2b66745adddbdc56dc93e330977c9d8d205d25b5fec91dc9a8fa37f16be7902202705ab46f2d835f86e80fc5538e5b9f4c824078acf6a511fba52f3c71669955e014730440220763502af34387d8fc46532370426ad6b29ba0e37994032f83c19d379b9574e9a022007ee250924724beca25b20e597dd18a065942f1052e1757dc74fcc69c564949d01004c8676a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac6721034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed97c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67c0a6148a486ff2e31d6158bf39e2608864d63fefd09d5b88ac6868",
				},
				{ // 4,3
					index:           4,
					remoteSigHex:    "30450221008ec888e36e4a4b3dc2ed6b823319855b2ae03006ca6ae0d9aa7e24bfc1d6f07102203b0f78885472a67ff4fe5916c0bb669487d659527509516fc3a08e87a2cc0a7c",
					resolutionTxHex: "02000000000101ca94a9ad516ebc0c4bdd7b6254871babfa978d5accafb554214137d398bfcf6a03000000000000000001f2090000000000002200204adb4e2f00643db396dd120d4e7dc17625f5f2c11a40d857accc862d6b7dd80e05004830450221008ec888e36e4a4b3dc2ed6b823319855b2ae03006ca6ae0d9aa7e24bfc1d6f07102203b0f78885472a67ff4fe5916c0bb669487d659527509516fc3a08e87a2cc0a7c0147304402202c3e14282b84b02705dfd00a6da396c9fe8a8bcb1d3fdb4b20a4feba09440e8b02202b058b39aa9b0c865b22095edcd9ff1f71bbfe20aa4993755e54d042755ed0d5012004040404040404040404040404040404040404040404040404040404040404048a76a91414011f7254d96b819c76986c277d115efce6f7b58763ac67210394854aa6eab5b2a8122cc726e9dded053a2184d88256816826d6231c068d4a5b7c8201208763a91418bc1a114ccf9c052d3d23e28d3b0a9d1227434288527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae677502f801b175ac686800000000",
				},
			},
			expectedCommitmentTxHex: "0200000001c402e4061d82e9d8cc5b653813bc4cd5afebeffdc71c664258f41275804d4af9000000000038b02b8006d007000000000000000017a9143fbe957cc060bd7a34884fb2357648dd9e737e1787d007000000000000000017a914f4d6aadd25266f19107005e364cc9124b5009cca87b80b000000000000000017a914524799cd6e521eb575a038ed80bb03e03a91d2e187a00f000000000000000017a914cdbd3d056ef0e2c8f066a19ee85e52e70e07fe5b87c0c62d000000000000001976a914f682e6254058108032264237a8e62b4777400d4e88acd69c6a0000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda0873e1952200000000001000000000000000000000000ffffffffd947304402200b2ad0c10db06a076767f70e08b9700c7b6c2bab03d74b9a466acb0d6925dcba02206279aaadce6512c4b95490942f0aa65bc0cec9746996fb6576cf15a05187e3c801483045022100e18a13454fdd8e3fe1b4118cd0795a9bcedd31add1e61f34cec66860fa04dcc10220058257355375214a3e332e83e1d3bd37277d0ca1ac96e4cfc16f7dec8c7437dd01475221023da092f6980e58d2c037173180e9a465476026ee50f96695963e8efe436f54eb21037f6fc2b0e0d63fab64424ef15991dfb76151ba0cf60dd2992015bf81b2eafdd552ae",
			remoteSigHex:            "3045022100e18a13454fdd8e3fe1b4118cd0795a9bcedd31add1e61f34cec66860fa04dcc10220058257355375214a3e332e83e1d3bd37277d0ca1ac96e4cfc16f7dec8c7437dd",
		},
		{ // 5
			commitment: channeldb.ChannelCommitment{
				CommitHeight:  42,
				LocalBalance:  6988000000,
				RemoteBalance: 3000000000,
				FeePerKB:      2070,
			},
			htlcDescs: []htlcDesc{
				{ // 5,0
					index:           2,
					remoteSigHex:    "3045022100d91ee565fa843b65c19e6aa0e079cda9d354f1d4d7d3a840504d8b4db54e2f3e0220645a70819fc1affa9f7eeab01626f0452269085b0b27211405f08484fe1a2d51",
					resolutionTxHex: "020000000147cc9b8ab9a47070d655cf67a8f691d84fc4ce168aff0e8888645c8f8d5d59d900000000000000000001a304000000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda087f6010000000000000100000000000000000000000000000000fd1a01483045022100d91ee565fa843b65c19e6aa0e079cda9d354f1d4d7d3a840504d8b4db54e2f3e0220645a70819fc1affa9f7eeab01626f0452269085b0b27211405f08484fe1a2d5101473044022069e49f7f7f5e5c229425179deef10b8ad54e2ae09c03e53c6886b4df9e20a76c02207ddd8a8d9ff9309e6291cb7e34f204f2ed80f88dd266787b6feebbe39962d85501004c8676a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac6721034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed97c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67c0a614b43e1b38138a41b37f7cd9a1d274bc63e3a9b5d188ac6868",
				},
				{ // 5,1
					index:           3,
					remoteSigHex:    "304402200cb2a5569642cf6d70c8442df3c11e473d9519f054eb09a8b9c0051b3a063be6022076a969b294a16d163b861f1985958d37a05093175ca7da3de115b022ac611fa0",
					resolutionTxHex: "020000000147cc9b8ab9a47070d655cf67a8f691d84fc4ce168aff0e8888645c8f8d5d59d9010000000000000000018b08000000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda087f7010000000000000100000000000000000000000000000000fd190147304402200cb2a5569642cf6d70c8442df3c11e473d9519f054eb09a8b9c0051b3a063be6022076a969b294a16d163b861f1985958d37a05093175ca7da3de115b022ac611fa0014730440220173d28053059b02f932b1a03cdac5fc01566b7fe175f563f9081c8c9ae8184ab022050e2a8787f1d7af67ebdd951d57b3fe10660fc1d7d6d2c0a46b62e0dfacb826901004c8676a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac6721034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed97c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67c0a6148a486ff2e31d6158bf39e2608864d63fefd09d5b88ac6868",
				},
				{ // 5,2
					index:           4,
					remoteSigHex:    "3045022100c9458a4d2cbb741705577deb0a890e5cb90ee141be0400d3162e533727c9cb2102206edcf765c5dc5e5f9b976ea8149bf8607b5a0efb30691138e1231302b640d2a4",
					resolutionTxHex: "020000000112b0f746ca4c74cd579424e3110a8e75e1e00a24f4f37eb5cbc3bfdb9a83944d020000000000000000018b0800000000000000002200204adb4e2f00643db396dd120d4e7dc17625f5f2c11a40d857accc862d6b7dd80ef7010000000000000100000000000000000000000000000000fd1901483045022100d4e69d363de993684eae7b37853c40722a4c1b4a7b588ad7b5d8a9b5006137a102207a069c628170ee34be5612747051bdcc087466dbaa68d5756ea81c10155aef180147304402200405a71afc1e023e501c390770e4859e84601986975293fbdae165224a2a2f4802206a88efb0fa687a0372155bc1d1c9feddf11c1fc9b25751326973079090e29cc401004c8576a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac67210394854aa6eab5b2a8122cc726e9dded053a2184d88256816826d6231c068d4a5b7c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67a9148a486ff2e31d6158bf39e2608864d63fefd09d5b88ac6868",
				},
			},
			expectedCommitmentTxHex: "0200000001c402e4061d82e9d8cc5b653813bc4cd5afebeffdc71c664258f41275804d4af9000000000038b02b8005d007000000000000000017a914f4d6aadd25266f19107005e364cc9124b5009cca87b80b000000000000000017a914524799cd6e521eb575a038ed80bb03e03a91d2e187a00f000000000000000017a914cdbd3d056ef0e2c8f066a19ee85e52e70e07fe5b87c0c62d000000000000001976a914f682e6254058108032264237a8e62b4777400d4e88ac1c9d6a0000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda0873e1952200000000001000000000000000000000000ffffffffd94730440220613335df1d1542bcefebd19f1094827af18efc5255bc6e17a82939f17b2ec5cd02201d7d1e9a5abb924a0727a61bd5080edd3dc1ef5258fa23e0b638c5ec16c08c6101483045022100a52122c999f526cf1d13f1c053fabdd37690da5cd7e01a7eaf3f9f935707005c02203da5494fb1f4b02dac52f7a4f3cc95c00bb87a18439f9e5c1ac217873ef70cd501475221023da092f6980e58d2c037173180e9a465476026ee50f96695963e8efe436f54eb21037f6fc2b0e0d63fab64424ef15991dfb76151ba0cf60dd2992015bf81b2eafdd552ae",
			remoteSigHex:            "3045022100a52122c999f526cf1d13f1c053fabdd37690da5cd7e01a7eaf3f9f935707005c02203da5494fb1f4b02dac52f7a4f3cc95c00bb87a18439f9e5c1ac217873ef70cd5",
		},
		{ // 6
			commitment: channeldb.ChannelCommitment{
				CommitHeight:  42,
				LocalBalance:  6988000000,
				RemoteBalance: 3000000000,
				FeePerKB:      2194,
			},
			htlcDescs: []htlcDesc{
				{ // 6,0
					index:           2,
					remoteSigHex:    "30450221009b338ebf687d72f9c6339394acc1348f11a0ed14ba6274f97b130bde77c4800c02206b72c21ec16e840743e0deced2b95d2b80e23d478ad9b73bc0d7cd8fb8b2f663",
					resolutionTxHex: "0200000001e27a42bd777ada604f4593df41c22b187bb05f829a150c5f2b5f997d5c9677a6000000000000000000017204000000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda087f6010000000000000100000000000000000000000000000000fd1a014830450221009b338ebf687d72f9c6339394acc1348f11a0ed14ba6274f97b130bde77c4800c02206b72c21ec16e840743e0deced2b95d2b80e23d478ad9b73bc0d7cd8fb8b2f663014730440220254483ecd330d345a860802c4c6a1d9e70dc31629537f5fa199bb49d2e04df2402201390cf4ea197c9ac7f936521e0f926e4e43c8bd5eb65dbafb3f9a93fdc209fc301004c8676a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac6721034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed97c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67c0a614b43e1b38138a41b37f7cd9a1d274bc63e3a9b5d188ac6868",
				},
				{ // 6,1
					index:           3,
					remoteSigHex:    "3044022047acbd17627df3e530d06fa142e6472adb476113785f732b6bad532eccc31db6022005c2e525cf1907e53ccb7bc45a6f1c824237ed7cecb7afa55e0aa9ee6ac23916",
					resolutionTxHex: "0200000001e27a42bd777ada604f4593df41c22b187bb05f829a150c5f2b5f997d5c9677a6010000000000000000015a08000000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda087f7010000000000000100000000000000000000000000000000fd1a01473044022047acbd17627df3e530d06fa142e6472adb476113785f732b6bad532eccc31db6022005c2e525cf1907e53ccb7bc45a6f1c824237ed7cecb7afa55e0aa9ee6ac23916014830450221009e3f309d7f7aa4aa6258245ba4a350b90a8aac345dfa47ab9d589048b5fb1cb90220722a529c4a5367cb67a0eb841ea73a9668741746df597cfd3326f07be859904e01004c8676a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac6721034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed97c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67c0a6148a486ff2e31d6158bf39e2608864d63fefd09d5b88ac6868",
				},
				{ // 6,2
					index:           4,
					remoteSigHex:    "3045022100a12a9a473ece548584aabdd051779025a5ed4077c4b7aa376ec7a0b1645e5a48022039490b333f53b5b3e2ddde1d809e492cba2b3e5fc3a436cd3ffb4cd3d500fa5a",
					resolutionTxHex: "02000000000101fb824d4e4dafc0f567789dee3a6bce8d411fe80f5563d8cdfdcc7d7e4447d43a020000000000000000019a090000000000002200204adb4e2f00643db396dd120d4e7dc17625f5f2c11a40d857accc862d6b7dd80e0500483045022100a12a9a473ece548584aabdd051779025a5ed4077c4b7aa376ec7a0b1645e5a48022039490b333f53b5b3e2ddde1d809e492cba2b3e5fc3a436cd3ffb4cd3d500fa5a01483045022100ff200bc934ab26ce9a559e998ceb0aee53bc40368e114ab9d3054d9960546e2802202496856ca163ac12c143110b6b3ac9d598df7254f2e17b3b94c3ab5301f4c3b0012004040404040404040404040404040404040404040404040404040404040404048a76a91414011f7254d96b819c76986c277d115efce6f7b58763ac67210394854aa6eab5b2a8122cc726e9dded053a2184d88256816826d6231c068d4a5b7c8201208763a91418bc1a114ccf9c052d3d23e28d3b0a9d1227434288527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae677502f801b175ac686800000000",
				},
			},
			expectedCommitmentTxHex: "0200000001c402e4061d82e9d8cc5b653813bc4cd5afebeffdc71c664258f41275804d4af9000000000038b02b8005d007000000000000000017a914f4d6aadd25266f19107005e364cc9124b5009cca87b80b000000000000000017a914524799cd6e521eb575a038ed80bb03e03a91d2e187a00f000000000000000017a914cdbd3d056ef0e2c8f066a19ee85e52e70e07fe5b87c0c62d000000000000001976a914f682e6254058108032264237a8e62b4777400d4e88ace29c6a0000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda0873e1952200000000001000000000000000000000000ffffffffda483045022100df41b4cf7b7ad3d4b21fcca335085dc514427fc7be136a00f1ad6f3ca7f396c802200a48ddecc181ab1643ef26566200ba0413d300b4fc847a6c825a5782699ae52101483045022100c842c7321b7f98c616b8bb09359309968f9e05be40924210fe1bee51d8dc4664022018e99ab8f988bce0f9a8b00ddd7c2743067c09c63982d105534a1156dbc1aedb01475221023da092f6980e58d2c037173180e9a465476026ee50f96695963e8efe436f54eb21037f6fc2b0e0d63fab64424ef15991dfb76151ba0cf60dd2992015bf81b2eafdd552ae",
			remoteSigHex:            "3045022100c842c7321b7f98c616b8bb09359309968f9e05be40924210fe1bee51d8dc4664022018e99ab8f988bce0f9a8b00ddd7c2743067c09c63982d105534a1156dbc1aedb",
		},
		{ // 7
			commitment: channeldb.ChannelCommitment{
				CommitHeight:  42,
				LocalBalance:  6988000000,
				RemoteBalance: 3000000000,
				FeePerKB:      2195,
			},
			htlcDescs: []htlcDesc{
				{ // 7,0
					index:           3,
					remoteSigHex:    "304402201501f93fa469f0bd3541ff490f3d7a21fc364cb2f697b3f09236166e9cc5f2f602206b534bf7ede9c10da54539c7e6c654e77f0cabf1e54f8a8b04e6b0341ae7c1d2",
					resolutionTxHex: "0200000001d09dc33b92acc8340f4cd1feca38455a8dd4e19160eea5eebe3a45ad136333d7000000000000000000015a08000000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda087f7010000000000000100000000000000000000000000000000fd1a0147304402201501f93fa469f0bd3541ff490f3d7a21fc364cb2f697b3f09236166e9cc5f2f602206b534bf7ede9c10da54539c7e6c654e77f0cabf1e54f8a8b04e6b0341ae7c1d201483045022100a2f44ac49144beb9bf2c221c97f9c21e87090f855b5211cd9a638742371e02b902204e427955ccc637cf037fbe3842036fa7817a1d7cd0231954cae2527c8395178c01004c8676a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac6721034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed97c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67c0a6148a486ff2e31d6158bf39e2608864d63fefd09d5b88ac6868",
				},
				{ // 7,1
					index:           4,
					remoteSigHex:    "3045022100e769cb156aa2f7515d126cef7a69968629620ce82afcaa9e210969de6850df4602200b16b3f3486a229a48aadde520dbee31ae340dbadaffae74fbb56681fef27b92",
					resolutionTxHex: "020000000001014e16c488fa158431c1a82e8f661240ec0a71ba0ce92f2721a6538c510226ad5c0100000000000000000199090000000000002200204adb4e2f00643db396dd120d4e7dc17625f5f2c11a40d857accc862d6b7dd80e0500483045022100e769cb156aa2f7515d126cef7a69968629620ce82afcaa9e210969de6850df4602200b16b3f3486a229a48aadde520dbee31ae340dbadaffae74fbb56681fef27b92014730440220665b9cb4a978c09d1ca8977a534999bc8a49da624d0c5439451dd69cde1a003d022070eae0620f01f3c1bd029cc1488da13fb40fdab76f396ccd335479a11c5276d8012004040404040404040404040404040404040404040404040404040404040404048a76a91414011f7254d96b819c76986c277d115efce6f7b58763ac67210394854aa6eab5b2a8122cc726e9dded053a2184d88256816826d6231c068d4a5b7c8201208763a91418bc1a114ccf9c052d3d23e28d3b0a9d1227434288527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae677502f801b175ac686800000000",
				},
			},
			expectedCommitmentTxHex: "0200000001c402e4061d82e9d8cc5b653813bc4cd5afebeffdc71c664258f41275804d4af9000000000038b02b8004b80b000000000000000017a914524799cd6e521eb575a038ed80bb03e03a91d2e187a00f000000000000000017a914cdbd3d056ef0e2c8f066a19ee85e52e70e07fe5b87c0c62d000000000000001976a914f682e6254058108032264237a8e62b4777400d4e88ac2c9d6a0000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda0873e1952200000000001000000000000000000000000ffffffffda483045022100a24f521bf04f55b86976fda3c0b44951f56ecbdc6d9d5e478318736bc2a4c268022054bbf68831fb5de1c4a2b3087c96b8bc46ebe12dbc05444e574ee9d4b50ea44401483045022100d7a814abfc9d8976bc3dd6f44c0013a58e413e6cb3ee74f1f2894c2e4e0f31e8022037b78f8543c3c883402bd24c01b0d00e1de85f83bb507cc3d3a8d4b2bc65d07101475221023da092f6980e58d2c037173180e9a465476026ee50f96695963e8efe436f54eb21037f6fc2b0e0d63fab64424ef15991dfb76151ba0cf60dd2992015bf81b2eafdd552ae",
			remoteSigHex:            "3045022100d7a814abfc9d8976bc3dd6f44c0013a58e413e6cb3ee74f1f2894c2e4e0f31e8022037b78f8543c3c883402bd24c01b0d00e1de85f83bb507cc3d3a8d4b2bc65d071",
		},
		{ // 8
			commitment: channeldb.ChannelCommitment{
				CommitHeight:  42,
				LocalBalance:  6988000000,
				RemoteBalance: 3000000000,
				FeePerKB:      3702,
			},
			htlcDescs: []htlcDesc{
				{ // 8,0
					index:           3,
					remoteSigHex:    "3045022100c91d5b8ab25f212a9f7054e1eaa4367ff91e66191d662f521d7c4685f9ba25080220329f8074419dd3a0a7670c2be531e4b7659f4bc2bb1af6190acf54c75ca5568a",
					resolutionTxHex: "020000000156dbd426d58317f6f542ffc72bcb223fee8ccf0cabb1811cae09942713d3c8dd000000000000000000010a06000000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda087f7010000000000000100000000000000000000000000000000fd1a01483045022100c91d5b8ab25f212a9f7054e1eaa4367ff91e66191d662f521d7c4685f9ba25080220329f8074419dd3a0a7670c2be531e4b7659f4bc2bb1af6190acf54c75ca5568a0147304402200fe315377691c35c8c0392efa500d754c2e6fdbfa241b0a628fd5fb3a27f5cea0220062befd580eeb587fed6581dd669a46f4d01f6b6096e550b7acf590c83536b7e01004c8676a914f06a2ee4f3cc96a8b6963e14c601fef5ee3de3ce8763ac6721034ac695f3269d836cd00072a088a9109d83609e6e69955bcf415acda0179b6ed97c820120876475527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae67c0a6148a486ff2e31d6158bf39e2608864d63fefd09d5b88ac6868",
				},
				{ // 8,1
					index:           4,
					remoteSigHex:    "3045022100ea9dc2a7c3c3640334dab733bb4e036e32a3106dc707b24227874fa4f7da746802204d672f7ac0fe765931a8df10b81e53a3242dd32bd9dc9331eb4a596da87954e9",
					resolutionTxHex: "02000000000101b8de11eb51c22498fe39722c7227b6e55ff1a94146cf638458cb9bc6a060d3a30100000000000000000176050000000000002200204adb4e2f00643db396dd120d4e7dc17625f5f2c11a40d857accc862d6b7dd80e0500483045022100ea9dc2a7c3c3640334dab733bb4e036e32a3106dc707b24227874fa4f7da746802204d672f7ac0fe765931a8df10b81e53a3242dd32bd9dc9331eb4a596da87954e9014730440220048a41c660c4841693de037d00a407810389f4574b3286afb7bc392a438fa3f802200401d71fa87c64fe621b49ac07e3bf85157ac680acb977124da28652cc7f1a5c012004040404040404040404040404040404040404040404040404040404040404048a76a91414011f7254d96b819c76986c277d115efce6f7b58763ac67210394854aa6eab5b2a8122cc726e9dded053a2184d88256816826d6231c068d4a5b7c8201208763a91418bc1a114ccf9c052d3d23e28d3b0a9d1227434288527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae677502f801b175ac686800000000",
				},
			},
			expectedCommitmentTxHex: "0200000001c402e4061d82e9d8cc5b653813bc4cd5afebeffdc71c664258f41275804d4af9000000000038b02b8004b80b000000000000000017a914524799cd6e521eb575a038ed80bb03e03a91d2e187a00f000000000000000017a914cdbd3d056ef0e2c8f066a19ee85e52e70e07fe5b87c0c62d000000000000001976a914f682e6254058108032264237a8e62b4777400d4e88aca19a6a0000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda0873e1952200000000001000000000000000000000000ffffffffd9483045022100ac1d440438173e6fd778d4a34ac6e21ec183bbc2d3459bf882b994519747b10402202a91c5378bc48eefd85b669308f83dd55f9ca6dca5c15dad7b6941b7bac0302b01473044022072471915a1ea5ddf1938b17bdb3b861119bc762dcbaca3d51665164972f0e57302205123433283257b637388f011f4474366560d72989d13e097da347e766d7aa16501475221023da092f6980e58d2c037173180e9a465476026ee50f96695963e8efe436f54eb21037f6fc2b0e0d63fab64424ef15991dfb76151ba0cf60dd2992015bf81b2eafdd552ae",
			remoteSigHex:            "3044022072471915a1ea5ddf1938b17bdb3b861119bc762dcbaca3d51665164972f0e57302205123433283257b637388f011f4474366560d72989d13e097da347e766d7aa165",
		},
		{ // 9
			commitment: channeldb.ChannelCommitment{
				CommitHeight:  42,
				LocalBalance:  6988000000,
				RemoteBalance: 3000000000,
				FeePerKB:      3703,
			},
			htlcDescs: []htlcDesc{
				{ // 9,0
					index:           4,
					remoteSigHex:    "3044022044f65cf833afdcb9d18795ca93f7230005777662539815b8a601eeb3e57129a902206a4bf3e53392affbba52640627defa8dc8af61c958c9e827b2798ab45828abdd",
					resolutionTxHex: "020000000001011c076aa7fb3d7460d10df69432c904227ea84bbf3134d4ceee5fb0f135ef206d0000000000000000000175050000000000002200204adb4e2f00643db396dd120d4e7dc17625f5f2c11a40d857accc862d6b7dd80e0500473044022044f65cf833afdcb9d18795ca93f7230005777662539815b8a601eeb3e57129a902206a4bf3e53392affbba52640627defa8dc8af61c958c9e827b2798ab45828abdd01483045022100b94d931a811b32eeb885c28ddcf999ae1981893b21dd1329929543fe87ce793002206370107fdd151c5f2384f9ceb71b3107c69c74c8ed5a28a94a4ab2d27d3b0724012004040404040404040404040404040404040404040404040404040404040404048a76a91414011f7254d96b819c76986c277d115efce6f7b58763ac67210394854aa6eab5b2a8122cc726e9dded053a2184d88256816826d6231c068d4a5b7c8201208763a91418bc1a114ccf9c052d3d23e28d3b0a9d1227434288527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae677502f801b175ac686800000000",
				},
			},
			expectedCommitmentTxHex: "0200000001c402e4061d82e9d8cc5b653813bc4cd5afebeffdc71c664258f41275804d4af9000000000038b02b8003a00f000000000000000017a914cdbd3d056ef0e2c8f066a19ee85e52e70e07fe5b87c0c62d000000000000001976a914f682e6254058108032264237a8e62b4777400d4e88ac1f9b6a0000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda0873e1952200000000001000000000000000000000000ffffffffda483045022100bc4c34c8a7f119db1c7d8c8a1d30857bdb1ff26fc399c58b42b7b891b00248520220031696003f70829ab9577494864e90717959cbc625ff725782aa2b14af7ca93a01483045022100cc5c0b911d295b686f5ef8e984997061fe52af1d53300d1017349cbf7e76a1fd02207fb45734273358663722e1314329513b96cdde7fea35eec375b8f2ef0ff8466801475221023da092f6980e58d2c037173180e9a465476026ee50f96695963e8efe436f54eb21037f6fc2b0e0d63fab64424ef15991dfb76151ba0cf60dd2992015bf81b2eafdd552ae",
			remoteSigHex:            "3045022100cc5c0b911d295b686f5ef8e984997061fe52af1d53300d1017349cbf7e76a1fd02207fb45734273358663722e1314329513b96cdde7fea35eec375b8f2ef0ff84668",
		},
		{ // 10
			commitment: channeldb.ChannelCommitment{
				CommitHeight:  42,
				LocalBalance:  6988000000,
				RemoteBalance: 3000000000,
				FeePerKB:      4914,
			},
			htlcDescs: []htlcDesc{
				{ // 10,0
					index:           4,
					remoteSigHex:    "3045022100fcb38506bfa11c02874092a843d0cc0a8613c23b639832564a5f69020cb0f6ba02206508b9e91eaa001425c190c68ee5f887e1ad5b1b314002e74db9dbd9e42dbecf",
					resolutionTxHex: "0200000000010110a3fdcbcd5db477cd3ad465e7f501ffa8c437e8301f00a6061138590add757f0000000000000000000122020000000000002200204adb4e2f00643db396dd120d4e7dc17625f5f2c11a40d857accc862d6b7dd80e0500483045022100fcb38506bfa11c02874092a843d0cc0a8613c23b639832564a5f69020cb0f6ba02206508b9e91eaa001425c190c68ee5f887e1ad5b1b314002e74db9dbd9e42dbecf0148304502210086e76b460ddd3cea10525fba298405d3fe11383e56966a5091811368362f689a02200f72ee75657915e0ede89c28709acd113ede9e1b7be520e3bc5cda425ecd6e68012004040404040404040404040404040404040404040404040404040404040404048a76a91414011f7254d96b819c76986c277d115efce6f7b58763ac67210394854aa6eab5b2a8122cc726e9dded053a2184d88256816826d6231c068d4a5b7c8201208763a91418bc1a114ccf9c052d3d23e28d3b0a9d1227434288527c21030d417a46946384f88d5f3337267c5e579765875dc4daca813e21734b140639e752ae677502f801b175ac686800000000",
				},
			},
			expectedCommitmentTxHex: "0200000001c402e4061d82e9d8cc5b653813bc4cd5afebeffdc71c664258f41275804d4af9000000000038b02b8003a00f000000000000000017a914cdbd3d056ef0e2c8f066a19ee85e52e70e07fe5b87c0c62d000000000000001976a914f682e6254058108032264237a8e62b4777400d4e88ac3d996a0000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda0873e1952200000000001000000000000000000000000ffffffffd94730440220392a2d6c88f6b41f8005437863ff273e3998f2876b9f45bdafb754f194a8671802205cf447f67b80c8e2ba5e779568b1e27043afd55049d269557b4cd7cdf8ef3250014830450221008bbeefed6835aedceaeefa38f3fef0569534e8ce8bf6d8a3072526d05345c53202207fbb04d4476096cae365b528031b41fcd3090d8eefdca8916c688571d7f3602801475221023da092f6980e58d2c037173180e9a465476026ee50f96695963e8efe436f54eb21037f6fc2b0e0d63fab64424ef15991dfb76151ba0cf60dd2992015bf81b2eafdd552ae",
			remoteSigHex:            "30450221008bbeefed6835aedceaeefa38f3fef0569534e8ce8bf6d8a3072526d05345c53202207fbb04d4476096cae365b528031b41fcd3090d8eefdca8916c688571d7f36028",
		},
		{ // 11
			commitment: channeldb.ChannelCommitment{
				CommitHeight:  42,
				LocalBalance:  6988000000,
				RemoteBalance: 3000000000,
				FeePerKB:      4915,
			},
			htlcDescs:               []htlcDesc{},
			expectedCommitmentTxHex: "0200000001c402e4061d82e9d8cc5b653813bc4cd5afebeffdc71c664258f41275804d4af9000000000038b02b8002c0c62d000000000000001976a914f682e6254058108032264237a8e62b4777400d4e88ace3996a0000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda0873e1952200000000001000000000000000000000000ffffffffda483045022100e7fdcfb6a28ea2191e4c5147cc8f127a34915eef373d785ed691bb4acecbd9c70220323bdb4aa9ab743f76d4fa41a6266c0097674a67395358070c4ed2f758e080aa01483045022100ab3723ed562027a17f33f3a2c3bd1a958c7dd8e3e7476f7cef4416a7dc592ee2022021c9c97b998719be84f520415e9905a5cd7dbcad54c0b620ab5afd19e660ae6101475221023da092f6980e58d2c037173180e9a465476026ee50f96695963e8efe436f54eb21037f6fc2b0e0d63fab64424ef15991dfb76151ba0cf60dd2992015bf81b2eafdd552ae",
			remoteSigHex:            "3045022100ab3723ed562027a17f33f3a2c3bd1a958c7dd8e3e7476f7cef4416a7dc592ee2022021c9c97b998719be84f520415e9905a5cd7dbcad54c0b620ab5afd19e660ae61",
		},
		{ // 12
			commitment: channeldb.ChannelCommitment{
				CommitHeight:  42,
				LocalBalance:  6988000000,
				RemoteBalance: 3000000000,
				FeePerKB:      9651180,
			},
			htlcDescs:               []htlcDesc{},
			expectedCommitmentTxHex: "0200000001c402e4061d82e9d8cc5b653813bc4cd5afebeffdc71c664258f41275804d4af9000000000038b02b8002c0c62d000000000000001976a914f682e6254058108032264237a8e62b4777400d4e88ac1b06350000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda0873e1952200000000001000000000000000000000000ffffffffd9473044022057d1b2bd4375e82a07147157e2ce5315d59aacd36750b75e0a135052c7ffec2902206c6fba9f71bc9c8a92c7d65238d687bda31285b343ad65ce8784e2d589f3004501483045022100c7b4ee1c16739013fb0274838a3a4b80b6e01887e20dfc6629c291bcf5fbcf5c02204bba54deaedd44b06783561fe61f7ea261d341b3ea577efdc53ce99c1bd3824901475221023da092f6980e58d2c037173180e9a465476026ee50f96695963e8efe436f54eb21037f6fc2b0e0d63fab64424ef15991dfb76151ba0cf60dd2992015bf81b2eafdd552ae",
			remoteSigHex:            "3045022100c7b4ee1c16739013fb0274838a3a4b80b6e01887e20dfc6629c291bcf5fbcf5c02204bba54deaedd44b06783561fe61f7ea261d341b3ea577efdc53ce99c1bd38249",
		},
		{ // 13
			commitment: channeldb.ChannelCommitment{
				CommitHeight:  42,
				LocalBalance:  6988000000,
				RemoteBalance: 3000000000,
				FeePerKB:      9651181,
			},
			htlcDescs:               []htlcDesc{},
			expectedCommitmentTxHex: "0200000001c402e4061d82e9d8cc5b653813bc4cd5afebeffdc71c664258f41275804d4af9000000000038b02b8002c0c62d000000000000001976a914f682e6254058108032264237a8e62b4777400d4e88ac1b06350000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda0873e1952200000000001000000000000000000000000ffffffffd9473044022057d1b2bd4375e82a07147157e2ce5315d59aacd36750b75e0a135052c7ffec2902206c6fba9f71bc9c8a92c7d65238d687bda31285b343ad65ce8784e2d589f3004501483045022100c7b4ee1c16739013fb0274838a3a4b80b6e01887e20dfc6629c291bcf5fbcf5c02204bba54deaedd44b06783561fe61f7ea261d341b3ea577efdc53ce99c1bd3824901475221023da092f6980e58d2c037173180e9a465476026ee50f96695963e8efe436f54eb21037f6fc2b0e0d63fab64424ef15991dfb76151ba0cf60dd2992015bf81b2eafdd552ae",
			remoteSigHex:            "3045022100c7b4ee1c16739013fb0274838a3a4b80b6e01887e20dfc6629c291bcf5fbcf5c02204bba54deaedd44b06783561fe61f7ea261d341b3ea577efdc53ce99c1bd38249",
		},
		{ // 14
			commitment: channeldb.ChannelCommitment{
				CommitHeight:  42,
				LocalBalance:  6988000000,
				RemoteBalance: 3000000000,
				FeePerKB:      9651936,
			},
			htlcDescs:               []htlcDesc{},
			expectedCommitmentTxHex: "0200000001c402e4061d82e9d8cc5b653813bc4cd5afebeffdc71c664258f41275804d4af9000000000038b02b8002c0c62d000000000000001976a914f682e6254058108032264237a8e62b4777400d4e88ac0805350000000000000017a91496fbabc4e9b687c24342b316586fc2cc8f70dda0873e1952200000000001000000000000000000000000ffffffffd94830450221008e8e2f2d046e8349873dff6636823c1aebf12f2247ec1fdd95d1dd30d5a150380220123ab652499bd4e874d407fb22b981eaf00b1bb60831771cf3b1fc779d52c2a501473044022020ee4a24358d5f06bac9e1595f415fb4c700ea4e5f9e86963aa433a9b2ce3465022012186c1308621a7e400b13d8f8e96b7894a594de174093ac526b5776ef5827f201475221023da092f6980e58d2c037173180e9a465476026ee50f96695963e8efe436f54eb21037f6fc2b0e0d63fab64424ef15991dfb76151ba0cf60dd2992015bf81b2eafdd552ae",
			remoteSigHex:            "3044022020ee4a24358d5f06bac9e1595f415fb4c700ea4e5f9e86963aa433a9b2ce3465022012186c1308621a7e400b13d8f8e96b7894a594de174093ac526b5776ef5827f2",
		},
	}

	fundingTxOut := channel.signDesc.Output

	for i, test := range testCases {
		expectedCommitmentTx, err := txFromHex(test.expectedCommitmentTxHex)
		if err != nil {
			t.Fatalf("Case %d: Failed to parse serialized tx: %v", i, err)
		}

		// Build required HTLC structs from raw test vector data.
		htlcs := make([]channeldb.HTLC, len(test.htlcDescs))
		for i, htlcDesc := range test.htlcDescs {
			htlcs[i], err = tc.getHTLC(i, &htlcDesc)
			if err != nil {
				t.Fatal(err)
			}
		}
		theHTLCView := htlcViewFromHTLCs(htlcs)

		feePerKB := chainfee.AtomPerKByte(test.commitment.FeePerKB)
		isOurs := true
		height := test.commitment.CommitHeight

		// Create unsigned commitment transaction.
		view, err := channel.commitBuilder.createUnsignedCommitmentTx(
			test.commitment.LocalBalance,
			test.commitment.RemoteBalance, isOurs, feePerKB,
			height, theHTLCView, keys,
		)
		if err != nil {
			t.Errorf("Case %d: Failed to create new commitment tx: %v", i, err)
			continue
		}

		commitmentView := &commitment{
			ourBalance:   view.ourBalance,
			theirBalance: view.theirBalance,
			txn:          view.txn,
			fee:          view.fee,
			height:       height,
			feePerKB:     feePerKB,
			dustLimit:    tc.dustLimit,
			isOurs:       isOurs,
		}

		// Initialize LocalCommit, which is used in getSignedCommitTx.
		channelState.LocalCommitment = test.commitment
		channelState.LocalCommitment.Htlcs = htlcs
		channelState.LocalCommitment.CommitTx = commitmentView.txn

		// This is the remote party's signature over the commitment
		// transaction which is included in the commitment tx's witness
		// data.
		channelState.LocalCommitment.CommitSig, err = hex.DecodeString(test.remoteSigHex)
		if err != nil {
			t.Fatalf("Case %d: Failed to parse serialized signature: %v",
				i, err)
		}

		commitTx, err := channel.getSignedCommitTx()
		if err != nil {
			t.Errorf("Case %d: Failed to sign commitment tx: %v", i, err)
			continue
		}

		// Sanity check the commitment to ensure it has a chance of
		// being valid.
		if err = checkSignedCommitmentTxSanity(commitTx, fundingTxOut, tc.netParams); err != nil {
			t.Errorf("Case %d: Failed commitment tx sanity check: %v", i, err)
		}

		// Check that commitment transaction was created correctly.
		if commitTx.TxHashWitness() != expectedCommitmentTx.MsgTx().TxHashWitness() {
			t.Errorf("Case %d: Generated unexpected commitment tx: "+
				"expected %s, got %s", i, spew.Sdump(expectedCommitmentTx),
				spew.Sdump(commitTx))
			continue
		}

		// Generate second-level HTLC transactions for HTLCs in
		// commitment tx.
		htlcResolutions, err := extractHtlcResolutions(
			chainfee.AtomPerKByte(test.commitment.FeePerKB), true, signer,
			htlcs, keys, &channel.channelState.LocalChanCfg,
			&channel.channelState.RemoteChanCfg, commitTx.TxHash(),
		)
		if err != nil {
			t.Errorf("Case %d: Failed to extract HTLC resolutions: %v", i, err)
			continue
		}

		resolutionIdx := 0
		for j, htlcDesc := range test.htlcDescs {
			// TODO: Check HTLC success transactions; currently not implemented.
			// resolutionIdx can be replaced by j when this is handled.
			if htlcs[j].Incoming {
				continue
			}

			expectedTx, err := txFromHex(htlcDesc.resolutionTxHex)
			if err != nil {
				t.Fatalf("Failed to parse serialized tx: %d %d %v", i, j, err)
			}

			htlcResolution := htlcResolutions.OutgoingHTLCs[resolutionIdx]
			resolutionIdx++

			actualTx := htlcResolution.SignedTimeoutTx
			if actualTx == nil {
				t.Errorf("Case %d: Failed to generate second level tx: "+
					"output %d, %v", i, j,
					htlcResolution)
				continue
			}

			// Sanity check the resulting tx to ensure it has a
			// chance of being mined.
			if err = checkSignedCommitmentSpendingTxSanity(actualTx, commitTx, tc.netParams); err != nil {
				t.Errorf("Case %d: Failed htlc resolution tx sanity check: "+
					"output %d, %v", i, j, err)
			}

			// Check that second-level HTLC transaction was created correctly.
			if actualTx.TxHashWitness() != expectedTx.MsgTx().TxHashWitness() {
				t.Fatalf("Case %d: Generated unexpected second level tx: "+
					"output %d, expected %s, got %s", i, j,
					expectedTx.MsgTx().TxHashWitness(), actualTx.TxHashWitness())
				continue
			}
		}

	}
}

// htlcViewFromHTLCs constructs an htlcView of PaymentDescriptors from a slice
// of channeldb.HTLC structs.
func htlcViewFromHTLCs(htlcs []channeldb.HTLC) *htlcView {
	var theHTLCView htlcView
	for _, htlc := range htlcs {
		paymentDesc := &PaymentDescriptor{
			RHash:   htlc.RHash,
			Timeout: htlc.RefundTimeout,
			Amount:  htlc.Amt,
		}
		if htlc.Incoming {
			theHTLCView.theirUpdates =
				append(theHTLCView.theirUpdates, paymentDesc)
		} else {
			theHTLCView.ourUpdates =
				append(theHTLCView.ourUpdates, paymentDesc)
		}
	}
	return &theHTLCView
}

func TestCommitTxStateHint(t *testing.T) {
	t.Parallel()

	stateHintTests := []struct {
		name       string
		from       uint64
		to         uint64
		inputs     int
		shouldFail bool
	}{
		{
			name:       "states 0 to 1000",
			from:       0,
			to:         1000,
			inputs:     1,
			shouldFail: false,
		},
		{
			name:       "states 'maxStateHint-1000' to 'maxStateHint'",
			from:       maxStateHint - 1000,
			to:         maxStateHint,
			inputs:     1,
			shouldFail: false,
		},
		{
			name:       "state 'maxStateHint+1'",
			from:       maxStateHint + 1,
			to:         maxStateHint + 10,
			inputs:     1,
			shouldFail: true,
		},
		{
			name:       "commit transaction with two inputs",
			inputs:     2,
			shouldFail: true,
		},
	}

	var obfuscator [StateHintSize]byte
	copy(obfuscator[:], testHdSeed[:StateHintSize])
	timeYesterday := uint32(time.Now().Unix() - 24*60*60)

	for _, test := range stateHintTests {
		commitTx := wire.NewMsgTx()
		commitTx.Version = input.LNTxVersion

		// Add supplied number of inputs to the commitment transaction.
		for i := 0; i < test.inputs; i++ {
			commitTx.AddTxIn(&wire.TxIn{})
		}

		for i := test.from; i <= test.to; i++ {
			stateNum := i

			err := SetStateNumHint(commitTx, stateNum, obfuscator)
			if err != nil && !test.shouldFail {
				t.Fatalf("unable to set state num %v: %v", i, err)
			} else if err == nil && test.shouldFail {
				t.Fatalf("Failed(%v): test should fail but did not", test.name)
			}

			locktime := commitTx.LockTime
			sequence := commitTx.TxIn[0].Sequence

			// Locktime should not be less than 500,000,000 and not larger
			// than the time 24 hours ago. One day should provide a good
			// enough buffer for the tests.
			if locktime < 5e8 || locktime > timeYesterday {
				if !test.shouldFail {
					t.Fatalf("The value of locktime (%v) may cause the commitment "+
						"transaction to be unspendable", locktime)
				}
			}

			if sequence&wire.SequenceLockTimeDisabled == 0 {
				if !test.shouldFail {
					t.Fatalf("Sequence locktime is NOT disabled when it should be")
				}
			}

			extractedStateNum := GetStateNumHint(commitTx, obfuscator)
			if extractedStateNum != stateNum && !test.shouldFail {
				t.Fatalf("state number mismatched, expected %v, got %v",
					stateNum, extractedStateNum)
			} else if extractedStateNum == stateNum && test.shouldFail {
				t.Fatalf("Failed(%v): test should fail but did not", test.name)
			}
		}
		t.Logf("Passed: %v", test.name)
	}
}

// testSpendValidation ensures that we're able to spend all outputs in the
// commitment transaction that we create.
func testSpendValidation(t *testing.T, tweakless bool) {
	// We generate a fake output, and the corresponding txin. This output
	// doesn't need to exist, as we'll only be validating spending from the
	// transaction that references this.
	txid, err := chainhash.NewHash(testHdSeed.CloneBytes())
	if err != nil {
		t.Fatalf("unable to create txid: %v", err)
	}
	fundingOut := &wire.OutPoint{
		Hash:  *txid,
		Index: 50,
		Tree:  wire.TxTreeRegular,
	}

	const channelBalance = dcrutil.Amount(1 * 10e8)
	const csvTimeout = 5
	fakeFundingTxIn := wire.NewTxIn(fundingOut, int64(channelBalance), nil)

	// We also set up set some resources for the commitment transaction.
	// Each side currently has 1 DCR within the channel, with a total
	// channel capacity of 2 DCR.
	aliceKeyPriv, aliceKeyPub := secp256k1.PrivKeyFromBytes(testWalletPrivKey)
	bobKeyPriv, bobKeyPub := secp256k1.PrivKeyFromBytes(bobsPrivKey)

	revocationPreimage := testHdSeed.CloneBytes()
	commitSecret, commitPoint := secp256k1.PrivKeyFromBytes(revocationPreimage)
	revokePubKey := input.DeriveRevocationPubkey(bobKeyPub, commitPoint)

	aliceDelayKey := input.TweakPubKey(aliceKeyPub, commitPoint)

	// Bob will have the channel "force closed" on him, so for the sake of
	// our commitments, if it's tweakless, his key will just be his regular
	// pubkey.
	bobPayKey := input.TweakPubKey(bobKeyPub, commitPoint)
	channelType := channeldb.SingleFunderBit
	if tweakless {
		bobPayKey = bobKeyPub
		channelType = channeldb.SingleFunderTweaklessBit
	}

	aliceCommitTweak := input.SingleTweakBytes(commitPoint, aliceKeyPub)
	bobCommitTweak := input.SingleTweakBytes(commitPoint, bobKeyPub)

	aliceSelfOutputSigner := &input.MockSigner{
		Privkeys: []*secp256k1.PrivateKey{aliceKeyPriv},
	}

	aliceChanCfg := &channeldb.ChannelConfig{
		ChannelConstraints: channeldb.ChannelConstraints{
			DustLimit: DefaultDustLimit(),
			CsvDelay:  csvTimeout,
		},
	}

	bobChanCfg := &channeldb.ChannelConfig{
		ChannelConstraints: channeldb.ChannelConstraints{
			DustLimit: DefaultDustLimit(),
			CsvDelay:  csvTimeout,
		},
	}

	// With all the test data set up, we create the commitment transaction.
	// We only focus on a single party's transactions, as the scripts are
	// identical with the roles reversed.
	//
	// This is Alice's commitment transaction, so she must wait a CSV delay
	// of 5 blocks before sweeping the output, while bob can spend
	// immediately with either the revocation key, or his regular key.
	keyRing := &CommitmentKeyRing{
		ToLocalKey:    aliceDelayKey,
		RevocationKey: revokePubKey,
		ToRemoteKey:   bobPayKey,
	}
	commitmentTx, err := CreateCommitTx(
		channelType, *fakeFundingTxIn, keyRing, aliceChanCfg,
		bobChanCfg, channelBalance, channelBalance,
	)
	if err != nil {
		t.Fatalf("unable to create commitment transaction: %v", nil)
	}

	delayOutput := commitmentTx.TxOut[0]
	regularOutput := commitmentTx.TxOut[1]

	// We're testing an uncooperative close, output sweep, so construct a
	// transaction which sweeps the funds to a random address.
	targetOutput, err := input.CommitScriptUnencumbered(aliceKeyPub)
	if err != nil {
		t.Fatalf("unable to create target output: %v", err)
	}
	sweepTx := wire.NewMsgTx()
	sweepTx.Version = input.LNTxVersion
	sweepTx.AddTxIn(wire.NewTxIn(&wire.OutPoint{
		Hash:  commitmentTx.TxHash(),
		Index: 0,
	}, delayOutput.Value, nil))
	sweepTx.AddTxOut(&wire.TxOut{
		PkScript: targetOutput,
		Value:    0.5 * 10e8,
	})

	// First, we'll test spending with Alice's key after the timeout.
	delayScript, err := input.CommitScriptToSelf(
		csvTimeout, aliceDelayKey, revokePubKey,
	)
	if err != nil {
		t.Fatalf("unable to generate alice delay script: %v", err)
	}
	sweepTx.TxIn[0].Sequence = input.LockTimeToSequence(false, csvTimeout)
	signDesc := &input.SignDescriptor{
		WitnessScript: delayScript,
		KeyDesc: keychain.KeyDescriptor{
			PubKey: aliceKeyPub,
		},
		SingleTweak: aliceCommitTweak,
		Output: &wire.TxOut{
			Value: int64(channelBalance),
		},
		HashType:   txscript.SigHashAll,
		InputIndex: 0,
	}
	aliceWitnessSpend, err := input.CommitSpendTimeout(
		aliceSelfOutputSigner, signDesc, sweepTx,
	)
	if err != nil {
		t.Fatalf("unable to generate delay commit spend witness: %v", err)
	}
	sweepTx.TxIn[0].SignatureScript, err = input.WitnessStackToSigScript(aliceWitnessSpend)
	if err != nil {
		t.Fatalf("unable to convert witness stack to sigScript: %v", err)
	}

	vm, err := txscript.NewEngine(delayOutput.PkScript,
		sweepTx, 0, input.ScriptVerifyFlags, delayOutput.Version, nil)
	if err != nil {
		t.Fatalf("unable to create engine: %v", err)
	}
	if err := vm.Execute(); err != nil {
		t.Fatalf("spend from delay output is invalid: %v", err)
	}

	bobSigner := &input.MockSigner{Privkeys: []*secp256k1.PrivateKey{bobKeyPriv}}

	// Next, we'll test bob spending with the derived revocation key to
	// simulate the scenario when Alice broadcasts this commitment
	// transaction after it's been revoked.
	signDesc = &input.SignDescriptor{
		KeyDesc: keychain.KeyDescriptor{
			PubKey: bobKeyPub,
		},
		DoubleTweak:   commitSecret,
		WitnessScript: delayScript,
		Output: &wire.TxOut{
			Value: int64(channelBalance),
		},
		HashType:   txscript.SigHashAll,
		InputIndex: 0,
	}
	bobWitnessSpend, err := input.CommitSpendRevoke(bobSigner, signDesc,
		sweepTx)
	if err != nil {
		t.Fatalf("unable to generate revocation witness: %v", err)
	}
	sweepTx.TxIn[0].SignatureScript, err = input.WitnessStackToSigScript(bobWitnessSpend)
	if err != nil {
		t.Fatalf("unable to convert witness stack to sigScript: %v", err)
	}

	vm, err = txscript.NewEngine(delayOutput.PkScript,
		sweepTx, 0, input.ScriptVerifyFlags, delayOutput.Version, nil)
	if err != nil {
		t.Fatalf("unable to create engine: %v", err)
	}
	if err := vm.Execute(); err != nil {
		t.Fatalf("revocation spend is invalid: %v", err)
	}

	// In order to test the final scenario, we modify the TxIn of the sweep
	// transaction to instead point to the regular output (non delay)
	// within the commitment transaction.
	sweepTx.TxIn[0] = &wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			Hash:  commitmentTx.TxHash(),
			Index: 1,
		},
	}

	// Finally, we test bob sweeping his output as normal in the case that
	// Alice broadcasts this commitment transaction.
	bobScriptP2PKH, err := input.CommitScriptUnencumbered(bobPayKey)
	if err != nil {
		t.Fatalf("unable to create bob p2wkh script: %v", err)
	}
	signDesc = &input.SignDescriptor{
		KeyDesc: keychain.KeyDescriptor{
			PubKey: bobKeyPub,
		},
		WitnessScript: bobScriptP2PKH,
		Output: &wire.TxOut{
			Value:    int64(channelBalance),
			PkScript: bobScriptP2PKH,
		},
		HashType:   txscript.SigHashAll,
		InputIndex: 0,
	}
	if !tweakless {
		signDesc.SingleTweak = bobCommitTweak
	}
	bobRegularSpend, err := input.CommitSpendNoDelay(
		bobSigner, signDesc, sweepTx, tweakless,
	)
	if err != nil {
		t.Fatalf("unable to create bob regular spend: %v", err)
	}
	sweepTx.TxIn[0].SignatureScript, err = input.WitnessStackToSigScript(bobRegularSpend)
	if err != nil {
		t.Fatalf("unable to convert witness stack to sigScript: %v", err)
	}

	vm, err = txscript.NewEngine(regularOutput.PkScript,
		sweepTx, 0, input.ScriptVerifyFlags, regularOutput.Version, nil)
	if err != nil {
		t.Fatalf("unable to create engine: %v", err)
	}
	if err := vm.Execute(); err != nil {
		t.Fatalf("bob p2wkh spend is invalid: %v", err)
	}
}

// TestCommitmentSpendValidation test the spendability of both outputs within
// the commitment transaction.
//
// The following spending cases are covered by this test:
//   * Alice's spend from the delayed output on her commitment transaction.
//   * Bob's spend from Alice's delayed output when she broadcasts a revoked
//     commitment transaction.
//   * Bob's spend from his unencumbered output within Alice's commitment
//     transaction.
func TestCommitmentSpendValidation(t *testing.T) {
	t.Parallel()

	// In the modern network, all channels use the new tweakless format,
	// but we also need to support older nodes that want to open channels
	// with the legacy format, so we'll test spending in both scenarios.
	for _, tweakless := range []bool{true, false} {
		tweakless := tweakless
		t.Run(fmt.Sprintf("tweak=%v", tweakless), func(t *testing.T) {
			testSpendValidation(t, tweakless)
		})
	}
}
