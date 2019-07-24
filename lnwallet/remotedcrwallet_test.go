package lnwallet_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sync/atomic"
	"testing"
	"time"

	"github.com/decred/dcrd/rpcclient/v5"
	pb "github.com/decred/dcrwallet/rpc/walletrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultGRPCPort = 29555
)

var (
	activeNodes int32
)

func consumeSyncMsgs(syncStream pb.WalletLoaderService_RpcSyncClient, onSyncedChan chan struct{}) {
	for {
		msg, err := syncStream.Recv()
		if err != nil {
			// All errors are final here.
			return
		}
		if msg.Synced {
			onSyncedChan <- struct{}{}
		}
	}
}

// newTestRemoteDcrwallet creates a new dcrwallet process that can be used by a
// remotedcrwallet instance to perform the interface tests. This currently only
// supports running the wallet in rpc sync mode.
//
// This function returns the grpc conn and a cleanup function to close the
// wallet.
func newTestRemoteDcrwallet(t *testing.T, nodeName, dataDir string,
	hdSeed, privatePass []byte,
	dcrd rpcclient.ConnConfig) (*grpc.ClientConn, func()) {

	// Save the dcrd ca file in the wallet dir.
	cafile := path.Join(dataDir, "ca.cert")
	ioutil.WriteFile(cafile, dcrd.Certificates, 0644)

	// Setup the args to run the underlying dcrwallet.
	id := atomic.AddInt32(&activeNodes, 1)
	port := defaultGRPCPort + id
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	certFile := path.Join(dataDir, "rpc.cert")
	args := []string{
		"--noinitialload",
		"--debuglevel=debug",
		"--rpcconnect=" + dcrd.Host,
		"--username=" + dcrd.User,
		"--password=" + dcrd.Pass,
		"--cafile=" + cafile,
		"--simnet",
		"--nolegacyrpc",
		"--grpclisten=" + addr,
		"--appdata=" + dataDir,
	}

	logFilePath := path.Join(fmt.Sprintf("output-remotedcrw-%.2d-%s.log",
		id, nodeName))
	logFile, err := os.Create(logFilePath)
	if err != nil {
		t.Fatalf("Unable to create %s dcrwallet log file: %v",
			nodeName, err)
	}

	const dcrwalletExe = "dcrwallet-dcrlnd"

	// Run dcrwallet.
	cmd := exec.Command(dcrwalletExe, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	err = cmd.Start()
	if err != nil {
		t.Fatalf("Unable to start %s dcrwallet: %v", nodeName, err)
	}
	time.Sleep(time.Millisecond * 250)
	t.Logf("Started %s dcrwallet at %s", nodeName, dataDir)

	// Open the GRPC conn to the wallet.
	creds, err := credentials.NewClientTLSFromFile(certFile, "localhost")
	if err != nil {
		t.Fatalf("Unable to create credentials: %v", err)

	}
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		t.Fatalf("Unable to dial grpc: %v", err)
	}

	loader := pb.NewWalletLoaderServiceClient(conn)

	// Create the wallet.
	ctxb := context.Background()
	reqCreate := &pb.CreateWalletRequest{
		Seed:              hdSeed,
		PublicPassphrase:  privatePass,
		PrivatePassphrase: privatePass,
	}
	ctx, cancel := context.WithTimeout(ctxb, time.Second*5)
	defer cancel()
	_, err = loader.CreateWallet(ctx, reqCreate)
	if err != nil {
		t.Fatalf("unable to create wallet: %v", err)
	}

	// Run the rpc syncer.
	req := &pb.RpcSyncRequest{
		NetworkAddress:    dcrd.Host,
		Username:          dcrd.User,
		Password:          []byte(dcrd.Pass),
		Certificate:       dcrd.Certificates,
		DiscoverAccounts:  true,
		PrivatePassphrase: privatePass,
	}
	ctxSync, cancelSync := context.WithCancel(context.Background())
	syncStream, err := loader.RpcSync(ctxSync, req)
	if err != nil {
		cancelSync()
		t.Fatalf("error running rpc sync: %v", err)
	}

	// Wait for the wallet to sync. Remote wallets are assumed synced
	// before an ln wallet is started for them.
	onSyncedChan := make(chan struct{})
	go consumeSyncMsgs(syncStream, onSyncedChan)
	select {
	case <-onSyncedChan:
		// Sync done.
	case <-time.After(time.Second * 15):
		cancelSync()
		t.Fatalf("timeout waiting for initial sync to complete")
	}

	cleanup := func() {
		t.Logf("Going to cleanup %s", nodeName)
		cancelSync()
		if cmd.ProcessState != nil {
			return
		}

		err := cmd.Process.Signal(os.Interrupt)
		if err != nil {
			t.Errorf("Error sending SIGINT to %s dcrwallet: %v",
				nodeName, err)
			return
		}

		// Wait for dcrwallet to exit or force kill it after a timeout.
		// For this, we run the wait on a goroutine and signal once it
		// has returned.
		errChan := make(chan error)
		go func() {
			errChan <- cmd.Wait()
		}()

		select {
		case err := <-errChan:
			if err != nil {
				t.Errorf("%s dcrwallet exited with an error: %v",
					nodeName, err)
			}

		case <-time.After(time.Second * 15):
			t.Errorf("%s dcrwallet timed out after SIGINT", nodeName)
			err := cmd.Process.Kill()
			if err != nil {
				t.Errorf("Error killing %s dcrwallet: %v",
					nodeName, err)
			}
		}
	}

	return conn, cleanup
}
