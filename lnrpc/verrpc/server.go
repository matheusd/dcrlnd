package verrpc

import (
	"context"
	"runtime"

	"github.com/decred/dcrlnd/build"
	"google.golang.org/grpc"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

const subServerName = "VersionRPC"

var macPermissions = map[string][]bakery.Op{
	"/verrpc.Versioner/GetVersion": {{
		Entity: "info",
		Action: "read",
	}},
}

// Server is an rpc server that supports querying for information about the
// running binary.
type Server struct{}

// Start launches any helper goroutines required for the rpcServer to function.
//
// NOTE: This is part of the lnrpc.SubServer interface.
func (s *Server) Start() error {
	return nil
}

// Stop signals any active goroutines for a graceful closure.
//
// NOTE: This is part of the lnrpc.SubServer interface.
func (s *Server) Stop() error {
	return nil
}

// Name returns a unique string representation of the sub-server. This can be
// used to identify the sub-server and also de-duplicate them.
//
// NOTE: This is part of the lnrpc.SubServer interface.
func (s *Server) Name() string {
	return subServerName
}

// RegisterWithRootServer will be called by the root gRPC server to direct a
// sub RPC server to register itself with the main gRPC root server. Until this
// is called, each sub-server won't be able to have requests routed towards it.
//
// NOTE: This is part of the lnrpc.SubServer interface.
func (s *Server) RegisterWithRootServer(grpcServer *grpc.Server) error {
	RegisterVersionerServer(grpcServer, s)

	log.Debugf("Versioner RPC server successfully registered with root " +
		"gRPC server")

	return nil
}

// GetVersion returns information about the compiled binary.
func (s *Server) GetVersion(_ context.Context,
	_ *VersionRequest) (*Version, error) {

	major, minor, patch := build.MajorMinorPatch()
	return &Version{
		Commit:        build.Commit,
		CommitHash:    build.Commit,
		Version:       build.Version(),
		AppMajor:      uint32(major),
		AppMinor:      uint32(minor),
		AppPatch:      uint32(patch),
		AppPreRelease: build.PreRelease,
		BuildTags:     nil,
		GoVersion:     runtime.Version(),
	}, nil
}
