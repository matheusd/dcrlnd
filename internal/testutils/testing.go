package testutils

// TB is an interface that fulfills both testing.T and testing.B for their
// common methods.
type TB interface {
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Errorf(string, ...interface{})
	Failed() bool
	Helper()
	Logf(string, ...interface{})
}
