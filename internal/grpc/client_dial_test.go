package grpc

import (
    "testing"
)

func TestDial_Insecure(_ *testing.T) {
    // Dialing a localhost port with insecure flag should proceed to dial attempt and likely error,
    // but the function path is exercised regardless of outcome.
    _, _ = Dial(DialOptions{Address: "127.0.0.1:1", Insecure: true})
}


