package walletunlocker

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/decred/dcrd/chaincfg/v2"
	"github.com/decred/dcrlnd/aezeed"
	"github.com/decred/dcrlnd/chanbackup"
	"github.com/decred/dcrlnd/keychain"
	"github.com/decred/dcrlnd/lnrpc"
	"github.com/decred/dcrlnd/lnwallet"

	pb "decred.org/dcrwallet/rpc/walletrpc"
	"decred.org/dcrwallet/wallet"
	"decred.org/dcrwallet/wallet/txrules"
	"github.com/decred/dcrlnd/lnwallet/dcrwallet"
	walletloader "github.com/decred/dcrlnd/lnwallet/dcrwallet/loader"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ChannelsToRecover wraps any set of packed (serialized+encrypted) channel
// back ups together. These can be passed in when unlocking the wallet, or
// creating a new wallet for the first time with an existing seed.
type ChannelsToRecover struct {
	// PackedMultiChanBackup is an encrypted and serialized multi-channel
	// backup.
	PackedMultiChanBackup chanbackup.PackedMulti

	// PackedSingleChanBackups is a series of encrypted and serialized
	// single-channel backup for one or more channels.
	PackedSingleChanBackups chanbackup.PackedSingles
}

// WalletInitMsg is a message sent by the UnlockerService when a user wishes to
// set up the internal wallet for the first time. The user MUST provide a
// passphrase, but is also able to provide their own source of entropy. If
// provided, then this source of entropy will be used to generate the wallet's
// HD seed. Otherwise, the wallet will generate one itself.
type WalletInitMsg struct {
	// Passphrase is the passphrase that will be used to encrypt the wallet
	// itself. This MUST be at least 8 characters.
	Passphrase []byte

	// WalletSeed is the deciphered cipher seed that the wallet should use
	// to initialize itself.
	WalletSeed *aezeed.CipherSeed

	// RecoveryWindow is the address look-ahead used when restoring a seed
	// with existing funds. A recovery window zero indicates that no
	// recovery should be attempted, such as after the wallet's initial
	// creation.
	RecoveryWindow uint32

	// ChanBackups a set of static channel backups that should be received
	// after the wallet has been initialized.
	ChanBackups ChannelsToRecover
}

// WalletUnlockMsg is a message sent by the UnlockerService when a user wishes
// to unlock the internal wallet after initial setup. The user can optionally
// specify a recovery window, which will resume an interrupted rescan for used
// addresses.
type WalletUnlockMsg struct {
	// Passphrase is the passphrase that will be used to encrypt the wallet
	// itself. This MUST be at least 8 characters.
	Passphrase []byte

	// RecoveryWindow is the address look-ahead used when restoring a seed
	// with existing funds. A recovery window zero indicates that no
	// recovery should be attempted, such as after the wallet's initial
	// creation, but before any addresses have been created.
	RecoveryWindow uint32

	// Wallet is the loaded and unlocked Wallet. This is returned through
	// the channel to avoid it being unlocked twice (once to check if the
	// password is correct, here in the WalletUnlocker and again later when
	// lnd actually uses it). Because unlocking involves scrypt which is
	// resource intensive, we want to avoid doing it twice.
	Wallet *wallet.Wallet

	Loader *walletloader.Loader

	// Conn is the connection to a remote wallet when the daemon has been
	// configured to connect to a wallet instead of using the embedded one.
	Conn *grpc.ClientConn

	// ChanBackups a set of static channel backups that should be received
	// after the wallet has been unlocked.
	ChanBackups ChannelsToRecover
}

// UnlockerService implements the WalletUnlocker service used to provide lnd
// with a password for wallet encryption at startup. Additionally, during
// initial setup, users can provide their own source of entropy which will be
// used to generate the seed that's ultimately used within the wallet.
type UnlockerService struct {
	// InitMsgs is a channel that carries all wallet init messages.
	InitMsgs chan *WalletInitMsg

	// UnlockMsgs is a channel where unlock parameters provided by the rpc
	// client to be used to unlock and decrypt an existing wallet will be
	// sent.
	UnlockMsgs chan *WalletUnlockMsg

	chainDir       string
	noFreelistSync bool
	netParams      *chaincfg.Params
	macaroonFiles  []string

	dcrwHost    string
	dcrwCert    string
	dcrwAccount int32
}

// New creates and returns a new UnlockerService.
func New(chainDir string, params *chaincfg.Params, noFreelistSync bool,
	macaroonFiles []string, dcrwHost, dcrwCert string, dcrwAccount int32) *UnlockerService {

	return &UnlockerService{
		InitMsgs:       make(chan *WalletInitMsg, 1),
		UnlockMsgs:     make(chan *WalletUnlockMsg, 1),
		chainDir:       chainDir,
		noFreelistSync: noFreelistSync,
		netParams:      params,
		macaroonFiles:  macaroonFiles,
		dcrwHost:       dcrwHost,
		dcrwCert:       dcrwCert,
		dcrwAccount:    dcrwAccount,
	}
}

// GenSeed is the first method that should be used to instantiate a new lnd
// instance. This method allows a caller to generate a new aezeed cipher seed
// given an optional passphrase. If provided, the passphrase will be necessary
// to decrypt the cipherseed to expose the internal wallet seed.
//
// Once the cipherseed is obtained and verified by the user, the InitWallet
// method should be used to commit the newly generated seed, and create the
// wallet.
func (u *UnlockerService) GenSeed(ctx context.Context,
	in *lnrpc.GenSeedRequest) (*lnrpc.GenSeedResponse, error) {

	// Before we start, we'll ensure that the wallet hasn't already created
	// so we don't show a *new* seed to the user if one already exists.
	netDir := dcrwallet.NetworkDir(u.chainDir, u.netParams)
	loader := walletloader.NewLoader(u.netParams, netDir,
		&walletloader.StakeOptions{}, wallet.DefaultGapLimit, false,
		txrules.DefaultRelayFeePerKb, wallet.DefaultAccountGapLimit,
		false)
	walletExists, err := loader.WalletExists()
	if err != nil {
		return nil, err
	}
	if walletExists {
		return nil, fmt.Errorf("wallet already exists")
	}

	var entropy [aezeed.EntropySize]byte

	switch {
	// If the user provided any entropy, then we'll make sure it's sized
	// properly.
	case len(in.SeedEntropy) != 0 && len(in.SeedEntropy) != aezeed.EntropySize:
		return nil, fmt.Errorf("incorrect entropy length: expected "+
			"16 bytes, instead got %v bytes", len(in.SeedEntropy))

	// If the user provided the correct number of bytes, then we'll copy it
	// over into our buffer for usage.
	case len(in.SeedEntropy) == aezeed.EntropySize:
		copy(entropy[:], in.SeedEntropy)

	// Otherwise, we'll generate a fresh new set of bytes to use as entropy
	// to generate the seed.
	default:
		if _, err := rand.Read(entropy[:]); err != nil {
			return nil, err
		}
	}

	// Now that we have our set of entropy, we'll create a new cipher seed
	// instance.
	//
	cipherSeed, err := aezeed.New(
		keychain.KeyDerivationVersion, &entropy, time.Now(),
	)
	if err != nil {
		return nil, err
	}

	// With our raw cipher seed obtained, we'll convert it into an encoded
	// mnemonic using the user specified pass phrase.
	mnemonic, err := cipherSeed.ToMnemonic(in.AezeedPassphrase)
	if err != nil {
		return nil, err
	}

	// Additionally, we'll also obtain the raw enciphered cipher seed as
	// well to return to the user.
	encipheredSeed, err := cipherSeed.Encipher(in.AezeedPassphrase)
	if err != nil {
		return nil, err
	}

	return &lnrpc.GenSeedResponse{
		CipherSeedMnemonic: mnemonic[:],
		EncipheredSeed:     encipheredSeed[:],
	}, nil
}

// extractChanBackups is a helper function that extracts the set of channel
// backups from the proto into a format that we'll pass to higher level
// sub-systems.
func extractChanBackups(chanBackups *lnrpc.ChanBackupSnapshot) *ChannelsToRecover {
	// If there aren't any populated channel backups, then we can exit
	// early as there's nothing to extract.
	if chanBackups == nil || (chanBackups.SingleChanBackups == nil &&
		chanBackups.MultiChanBackup == nil) {
		return nil
	}

	// Now that we know there's at least a single back up populated, we'll
	// extract the multi-chan backup (if it's there).
	var backups ChannelsToRecover
	if chanBackups.MultiChanBackup != nil {
		multiBackup := chanBackups.MultiChanBackup
		backups.PackedMultiChanBackup = chanbackup.PackedMulti(
			multiBackup.MultiChanBackup,
		)
	}

	if chanBackups.SingleChanBackups == nil {
		return &backups
	}

	// Finally, we can extract all the single chan backups as well.
	for _, backup := range chanBackups.SingleChanBackups.ChanBackups {
		singleChanBackup := backup.ChanBackup

		backups.PackedSingleChanBackups = append(
			backups.PackedSingleChanBackups, singleChanBackup,
		)
	}

	return &backups
}

// InitWallet is used when lnd is starting up for the first time to fully
// initialize the daemon and its internal wallet. At the very least a wallet
// password must be provided. This will be used to encrypt sensitive material
// on disk.
//
// In the case of a recovery scenario, the user can also specify their aezeed
// mnemonic and passphrase. If set, then the daemon will use this prior state
// to initialize its internal wallet.
//
// Alternatively, this can be used along with the GenSeed RPC to obtain a
// seed, then present it to the user. Once it has been verified by the user,
// the seed can be fed into this RPC in order to commit the new wallet.
func (u *UnlockerService) InitWallet(ctx context.Context,
	in *lnrpc.InitWalletRequest) (*lnrpc.InitWalletResponse, error) {

	// Make sure the password meets our constraints.
	password := in.WalletPassword
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}

	// Require that the recovery window be non-negative.
	recoveryWindow := in.RecoveryWindow
	if recoveryWindow < 0 {
		return nil, fmt.Errorf("recovery window %d must be "+
			"non-negative", recoveryWindow)
	}

	gapLimit := wallet.DefaultGapLimit
	if int(recoveryWindow) > gapLimit {
		gapLimit = int(recoveryWindow)
	}

	// We'll then open up the directory that will be used to store the
	// wallet's files so we can check if the wallet already exists. This
	// loader is only used for this check and should not leak to the
	// outside.
	netDir := dcrwallet.NetworkDir(u.chainDir, u.netParams)
	loader := walletloader.NewLoader(u.netParams, netDir,
		&walletloader.StakeOptions{}, gapLimit, false,
		txrules.DefaultRelayFeePerKb, wallet.DefaultAccountGapLimit,
		false)

	walletExists, err := loader.WalletExists()
	if err != nil {
		return nil, err
	}

	// If the wallet already exists, then we'll exit early as we can't
	// create the wallet if it already exists!
	if walletExists {
		return nil, fmt.Errorf("wallet already exists")
	}

	// At this point, we know that the wallet doesn't already exist. So
	// we'll map the user provided aezeed and passphrase into a decoded
	// cipher seed instance.
	var mnemonic aezeed.Mnemonic
	copy(mnemonic[:], in.CipherSeedMnemonic)

	// If we're unable to map it back into the ciphertext, then either the
	// mnemonic is wrong, or the passphrase is wrong.
	cipherSeed, err := mnemonic.ToCipherSeed(in.AezeedPassphrase)
	if err != nil {
		return nil, err
	}

	// With the cipher seed deciphered, and the auth service created, we'll
	// now send over the wallet password and the seed. This will allow the
	// daemon to initialize itself and startup.
	initMsg := &WalletInitMsg{
		Passphrase:     password,
		WalletSeed:     cipherSeed,
		RecoveryWindow: uint32(gapLimit),
	}

	// Before we return the unlock payload, we'll check if we can extract
	// any channel backups to pass up to the higher level sub-system.
	chansToRestore := extractChanBackups(in.ChannelBackups)
	if chansToRestore != nil {
		initMsg.ChanBackups = *chansToRestore
	}

	u.InitMsgs <- initMsg

	return &lnrpc.InitWalletResponse{}, nil
}

// UnlockRemoteWallet sends the password provided by the incoming
// UnlockRemoteWalletRequest over the UnlockMsgs channel in case it
// successfully decrypts an existing remote wallet.
func (u *UnlockerService) unlockRemoteWallet(password []byte,
	chanBackups *lnrpc.ChanBackupSnapshot) (*lnrpc.UnlockWalletResponse, error) {

	ctxb := context.Background()

	creds, err := credentials.NewClientTLSFromFile(u.dcrwCert, "localhost")
	if err != nil {
		return nil, err

	}
	conn, err := grpc.Dial(u.dcrwHost, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	wallet := pb.NewWalletServiceClient(conn)

	// Ensure we can grab the privkey for the given account using the
	// provided password.
	getAcctReq := &pb.GetAccountExtendedPrivKeyRequest{
		AccountNumber: uint32(u.dcrwAccount),
		Passphrase:    password,
	}
	_, err = wallet.GetAccountExtendedPrivKey(ctxb, getAcctReq)
	if err != nil {
		return nil, fmt.Errorf("unable to get xpriv: %v", err)
	}

	// We successfully opened the wallet and pass the instance back to
	// avoid it needing to be unlocked again.
	walletUnlockMsg := &WalletUnlockMsg{
		Passphrase: password,
		Conn:       conn,
	}

	// Before we return the unlock payload, we'll check if we can extract
	// any channel backups to pass up to the higher level sub-system.
	chansToRestore := extractChanBackups(chanBackups)
	if chansToRestore != nil {
		walletUnlockMsg.ChanBackups = *chansToRestore
	}

	// At this point we were able to open the existing wallet with the
	// provided password. We send the password over the UnlockMsgs channel,
	// such that it can be used by lnd to open the wallet.
	u.UnlockMsgs <- walletUnlockMsg

	return &lnrpc.UnlockWalletResponse{}, nil
}

// UnlockWallet sends the password provided by the incoming UnlockWalletRequest
// over the UnlockMsgs channel in case it successfully decrypts an existing
// wallet found in the chain's wallet database directory.
func (u *UnlockerService) UnlockWallet(ctx context.Context,
	in *lnrpc.UnlockWalletRequest) (*lnrpc.UnlockWalletResponse, error) {

	password := in.WalletPassword
	if u.dcrwHost != "" && u.dcrwCert != "" {
		return u.unlockRemoteWallet(password, in.ChannelBackups)
	}
	gapLimit := wallet.DefaultGapLimit
	if int(in.RecoveryWindow) > gapLimit {
		gapLimit = int(in.RecoveryWindow)
	}

	netDir := dcrwallet.NetworkDir(u.chainDir, u.netParams)
	loader := walletloader.NewLoader(u.netParams, netDir,
		&walletloader.StakeOptions{}, gapLimit, false,
		txrules.DefaultRelayFeePerKb, wallet.DefaultAccountGapLimit,
		false)

	// Check if wallet already exists.
	walletExists, err := loader.WalletExists()
	if err != nil {
		return nil, err
	}

	if !walletExists {
		// Cannot unlock a wallet that does not exist!
		return nil, fmt.Errorf("wallet not found")
	}

	// Try opening the existing wallet with the provided password.
	unlockedWallet, err := loader.OpenExistingWallet(ctx, password)
	if err != nil {
		// Could not open wallet, most likely this means that provided
		// password was incorrect.
		return nil, err
	}

	// We successfully opened the wallet and pass the instance back to
	// avoid it needing to be unlocked again.
	walletUnlockMsg := &WalletUnlockMsg{
		Passphrase:     password,
		RecoveryWindow: uint32(gapLimit),
		Wallet:         unlockedWallet,
		Loader:         loader,
	}

	// Before we return the unlock payload, we'll check if we can extract
	// any channel backups to pass up to the higher level sub-system.
	chansToRestore := extractChanBackups(in.ChannelBackups)
	if chansToRestore != nil {
		walletUnlockMsg.ChanBackups = *chansToRestore
	}

	// At this point we was able to open the existing wallet with the
	// provided password. We send the password over the UnlockMsgs
	// channel, such that it can be used by lnd to open the wallet.
	u.UnlockMsgs <- walletUnlockMsg

	return &lnrpc.UnlockWalletResponse{}, nil
}

// ChangePassword changes the password of the wallet and sends the new password
// across the UnlockPasswords channel to automatically unlock the wallet if
// successful.
func (u *UnlockerService) ChangePassword(ctx context.Context,
	in *lnrpc.ChangePasswordRequest) (*lnrpc.ChangePasswordResponse, error) {

	netDir := dcrwallet.NetworkDir(u.chainDir, u.netParams)
	loader := walletloader.NewLoader(u.netParams, netDir,
		&walletloader.StakeOptions{}, wallet.DefaultGapLimit, false,
		txrules.DefaultRelayFeePerKb, wallet.DefaultAccountGapLimit,
		false)

	// First, we'll make sure the wallet exists for the specific chain and
	// network.
	walletExists, err := loader.WalletExists()
	if err != nil {
		return nil, err
	}

	if !walletExists {
		return nil, errors.New("wallet not found")
	}

	publicPw := in.CurrentPassword
	privatePw := in.CurrentPassword

	// If the current password is blank, we'll assume the user is coming
	// from a --noseedbackup state, so we'll use the default passwords.
	if len(in.CurrentPassword) == 0 {
		publicPw = lnwallet.DefaultPublicPassphrase
		privatePw = lnwallet.DefaultPrivatePassphrase
	}

	// Make sure the new password meets our constraints.
	if err := ValidatePassword(in.NewPassword); err != nil {
		return nil, err
	}

	// Load the existing wallet in order to proceed with the password change.
	w, err := loader.OpenExistingWallet(ctx, publicPw)
	if err != nil {
		return nil, err
	}
	// Unload the wallet to allow lnd to open it later on.
	defer loader.UnloadWallet()

	// Since the macaroon database is also encrypted with the wallet's
	// password, we'll remove all of the macaroon files so that they're
	// re-generated at startup using the new password. We'll make sure to do
	// this after unlocking the wallet to ensure macaroon files don't get
	// deleted with incorrect password attempts.
	for _, file := range u.macaroonFiles {
		err := os.Remove(file)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}

	// Attempt to change both the public and private passphrases for the
	// wallet. This will be done atomically in order to prevent one
	// passphrase change from being successful and not the other.
	//
	// TODO(decred) This is not an atomic operation. Discuss whether we want to
	// actually use the public pssword.
	err = w.ChangePrivatePassphrase(ctx, privatePw, in.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("unable to change wallet private passphrase: "+
			"%v", err)
	}
	err = w.ChangePublicPassphrase(ctx, publicPw, in.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("unable to change wallet public passphrase: "+
			"%v", err)
	}

	// Finally, send the new password across the UnlockPasswords channel to
	// automatically unlock the wallet.
	u.UnlockMsgs <- &WalletUnlockMsg{Passphrase: in.NewPassword}

	return &lnrpc.ChangePasswordResponse{}, nil
}

// ValidatePassword assures the password meets all of our constraints.
func ValidatePassword(password []byte) error {
	// Passwords should have a length of at least 8 characters.
	if len(password) < 8 {
		return errors.New("password must have at least 8 characters")
	}

	return nil
}
