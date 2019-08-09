package wtclientrpc

import (
	"github.com/decred/dcrlnd/build"
	"github.com/decred/slog"
)

// Subsystem defines the logging code for this subsystem.
const Subsystem = "WTCL"

// log is a logger that is initialized with no output filters.  This means the
// package will not perform any logging by default until the caller requests
// it.
var log slog.Logger

// The default amount of logging is none.
func init() {
	UseLogger(build.NewSubLogger(Subsystem, nil))
}

// DisableLog disables all library log output.  Logging output is disabled by
// by default until UseLogger is called.
func DisableLog() {
	UseLogger(slog.Disabled)
}

// UseLogger uses a specified Logger to output package logging info.  This
// should be used in preference to SetLogWriter if the caller is also using
// slog.
func UseLogger(logger slog.Logger) {
	log = logger
}

// logClosure is used to provide a closure over expensive logging operations so
// don't have to be performed when the logging level doesn't warrant it.
type logClosure func() string

// String invokes the underlying function and returns the result.
func (c logClosure) String() string {
	return c()
}

// newLogClosure returns a new closure over a function that returns a string
// which itself provides a Stringer interface so that it can be used with the
// logging system.
func newLogClosure(c func() string) logClosure {
	return logClosure(c)
}