package grpc

import (
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// DialOptions configures gRPC client dialing.
type DialOptions struct {
	Address  string
	Insecure bool
}

// Dial creates a gRPC client connection using a timeout and TLS/insecure creds.
func Dial(opts DialOptions) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if opts.Insecure {
		creds = insecure.NewCredentials()
	} else {
		creds = credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
	}
	return grpc.NewClient(
		opts.Address,
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
}


