package grpc

import (
    "context"
    "net"
    "testing"
    "time"

    pb "budget-bot/internal/pb/budget/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/protobuf/types/known/timestamppb"
    "go.uber.org/zap"
)

type fakeTxListServer struct{ pb.UnimplementedTransactionServiceServer }

func (s *fakeTxListServer) ListTransactions(_ context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
    // honor page size boundaries
    n := int(req.GetPage().GetPageSize())
    if n <= 0 { n = 10 }
    if n > 100 { n = 100 }
    out := make([]*pb.Transaction, 0, n)
    for i := 0; i < n; i++ {
        out = append(out, &pb.Transaction{Type: pb.TransactionType_TRANSACTION_TYPE_EXPENSE, Amount: &pb.Money{CurrencyCode: "RUB", MinorUnits: 100}, Comment: "x", OccurredAt: timestamppb.New(time.Now())})
    }
    return &pb.ListTransactionsResponse{Transactions: out}, nil
}

func startTxListServer(t *testing.T) (*grpc.Server, string) {
    t.Helper()
    lis, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatal(err) }
    s := grpc.NewServer()
    pb.RegisterTransactionServiceServer(s, &fakeTxListServer{})
    go func(){ _ = s.Serve(lis) }()
    return s, lis.Addr().String()
}

func TestGRPCTransactionClient_ListRecent_And_Export(t *testing.T) {
    srv, addr := startTxListServer(t)
    defer srv.Stop()
    conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil { t.Fatal(err) }
    defer func(){ _ = conn.Close() }()
    c := NewGRPCTransactionClient(pb.NewTransactionServiceClient(conn), zap.NewNop())
    ctx := context.Background()

    // ListRecent with <=0 uses default 10
    items, err := c.ListRecent(ctx, "tenant", 0, "tok")
    if err != nil || len(items) != 10 { t.Fatalf("recent len: %v n=%d", err, len(items)) }

    // ListForExport with limit 3 should return exactly 3
    from := time.Now().AddDate(0, 0, -7)
    to := time.Now()
    items2, err := c.ListForExport(ctx, "tenant", from, to, 3, "tok")
    if err != nil || len(items2) != 3 { t.Fatalf("export len: %v n=%d", err, len(items2)) }
}


