package grpc

import (
	"crypto/tls"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type DialOptions struct {
	Address  string
	Insecure bool
}

func Dial(opts DialOptions) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if opts.Insecure {
		creds = insecure.NewCredentials()
	} else {
		creds = credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
	}
	return grpc.Dial(
		opts.Address,
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
		grpc.WithReturnConnectionError(),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
		grpc.WithTimeout(5*time.Second),
	)
}


