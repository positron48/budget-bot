package grpc

import (
    "context"
    "net"
    "testing"
    "time"

    pb "budget-bot/internal/pb/budget/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"
    "google.golang.org/protobuf/types/known/timestamppb"
)

type fakeTxServer struct{ pb.UnimplementedTransactionServiceServer; sawAuth string }

func (s *fakeTxServer) CreateTransaction(ctx context.Context, _ *pb.CreateTransactionRequest) (*pb.CreateTransactionResponse, error) {
    if md, ok := metadata.FromIncomingContext(ctx); ok {
        vals := md.Get("authorization")
        if len(vals) > 0 { s.sawAuth = vals[0] }
    }
    return &pb.CreateTransactionResponse{Transaction: &pb.Transaction{Id: "tx1"}}, nil
}

func (s *fakeTxServer) ListTransactions(_ context.Context, _ *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
    return &pb.ListTransactionsResponse{Transactions: []*pb.Transaction{
        {Type: pb.TransactionType_TRANSACTION_TYPE_EXPENSE, Amount: &pb.Money{CurrencyCode: "RUB", MinorUnits: 10000}, Comment: "такси", OccurredAt: timestamppb.New(time.Now())},
    }}, nil
}

func startTxServer(t *testing.T) (*grpc.Server, string, *fakeTxServer) {
    t.Helper()
    lis, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatal(err) }
    s := grpc.NewServer()
    impl := &fakeTxServer{}
    pb.RegisterTransactionServiceServer(s, impl)
    go func() { _ = s.Serve(lis) }()
    return s, lis.Addr().String(), impl
}

func TestGRPCTransactionClient_CreateAndList(t *testing.T) {
    srv, addr, impl := startTxServer(t)
    defer srv.Stop()
    conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil { t.Fatal(err) }
    defer func(){ _ = conn.Close() }()
    c := NewGRPCTransactionClient(pb.NewTransactionServiceClient(conn))
    // Create
    _, err = c.CreateTransaction(context.Background(), &CreateTransactionRequest{Description: "такси", AmountMinor: 10000, Currency: "RUB", CategoryID: "cat", Type: "expense", OccurredAt: time.Now()}, "tok")
    if err != nil { t.Fatalf("create: %v", err) }
    if impl.sawAuth != "Bearer tok" { t.Fatalf("auth metadata not set: %q", impl.sawAuth) }
    // ListRecent
    got, err := c.ListRecent(context.Background(), "tenant", 1, "tok")
    if err != nil || len(got) != 1 { t.Fatalf("list: %v, n=%d", err, len(got)) }
}


