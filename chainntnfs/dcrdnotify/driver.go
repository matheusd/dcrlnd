package dcrdnotify

import (
	"errors"
	"fmt"

	"github.com/decred/dcrd/rpcclient/v2"
	"github.com/decred/dcrlnd/chainntnfs"
)

// createNewNotifier creates a new instance of the ChainNotifier interface
// implemented by DcrdNotifier.
func createNewNotifier(args ...interface{}) (chainntnfs.ChainNotifier, error) {
	const numRequiredArgs = 3
	if len(args) != numRequiredArgs {
		return nil, fmt.Errorf("incorrect number of arguments to "+
			".New(...), expected %d, instead passed %d", numRequiredArgs,
			len(args))
	}

	config, ok := args[0].(*rpcclient.ConnConfig)
	if !ok {
		return nil, errors.New("first argument to dcrdnotifier.New " +
			"is incorrect, expected a *rpcclient.ConnConfig")
	}

	spendHintCache, ok := args[1].(chainntnfs.SpendHintCache)
	if !ok {
		return nil, errors.New("second argument to dcrdnotifier.New " +
			"is incorrect, expected a chainntnfs.SpendHintCache")
	}

	confirmHintCache, ok := args[2].(chainntnfs.ConfirmHintCache)
	if !ok {
		return nil, errors.New("third argument to dcrdnotifier.New " +
			"is incorrect, expected a chainntnfs.ConfirmHintCache")
	}

	return New(config, spendHintCache, confirmHintCache)
}

// init registers a driver for the DcrdNotifier concrete implementation of the
// chainntnfs.ChainNotifier interface.
func init() {
	// Register the driver.
	notifier := &chainntnfs.NotifierDriver{
		NotifierType: notifierType,
		New:          createNewNotifier,
	}

	if err := chainntnfs.RegisterNotifier(notifier); err != nil {
		panic(fmt.Sprintf("failed to register notifier driver '%s': %v",
			notifierType, err))
	}
}
