package grpc

import (
    "context"
    "net"
    "testing"
    "time"

    pb "budget-bot/internal/pb/budget/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    status "google.golang.org/grpc/status"
)

type errTxServer struct{ pb.UnimplementedTransactionServiceServer }

func (s *errTxServer) CreateTransaction(_ context.Context, _ *pb.CreateTransactionRequest) (*pb.CreateTransactionResponse, error) {
    return nil, status.Error(13, "backend error")
}

func TestGRPCTransactionClient_Create_Error(t *testing.T) {
    lis, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatal(err) }
    srv := grpc.NewServer()
    pb.RegisterTransactionServiceServer(srv, &errTxServer{})
    go func(){ _ = srv.Serve(lis) }()
    defer srv.Stop()

    conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil { t.Fatal(err) }
    defer func(){ _ = conn.Close() }()
    c := NewGRPCTransactionClient(pb.NewTransactionServiceClient(conn))
    _, e := c.CreateTransaction(context.Background(), &CreateTransactionRequest{Description: "x", AmountMinor: 1, Currency: "RUB", Type: "expense", OccurredAt: time.Now()}, "tok")
    if e == nil { t.Fatalf("expected error") }
}


