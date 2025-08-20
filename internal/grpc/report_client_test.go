package grpc

import (
    "context"
    "net"
    "testing"
    "time"

    pb "budget-bot/internal/pb/budget/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
)

type fakeReportServer struct{ pb.UnimplementedReportServiceServer; sawAuth string }

func (s *fakeReportServer) GetMonthlySummary(ctx context.Context, r *pb.GetMonthlySummaryRequest) (*pb.GetMonthlySummaryResponse, error) {
    if md, ok := metadata.FromIncomingContext(ctx); ok {
        vals := md.Get("authorization")
        if len(vals) > 0 { s.sawAuth = vals[0] }
    }
    return &pb.GetMonthlySummaryResponse{
        Items: []*pb.MonthlyCategorySummaryItem{{CategoryId: "c1", CategoryName: "Питание", Type: pb.TransactionType_TRANSACTION_TYPE_EXPENSE, Total: &pb.Money{CurrencyCode: "RUB", MinorUnits: 12345}}},
        TotalIncome:  &pb.Money{CurrencyCode: "RUB", MinorUnits: 0},
        TotalExpense: &pb.Money{CurrencyCode: "RUB", MinorUnits: 12345},
    }, nil
}

func startReportServer(t *testing.T) (*grpc.Server, string, *fakeReportServer) {
    t.Helper()
    lis, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatal(err) }
    s := grpc.NewServer()
    impl := &fakeReportServer{}
    pb.RegisterReportServiceServer(s, impl)
    go func() { _ = s.Serve(lis) }()
    return s, lis.Addr().String(), impl
}

func TestGRPCReportClient_GetStatsAndTop(t *testing.T) {
    srv, addr, impl := startReportServer(t)
    defer srv.Stop()
    conn, err := grpc.Dial(addr, grpc.WithInsecure())
    if err != nil { t.Fatal(err) }
    defer func(){ _ = conn.Close() }()
    c := NewGRPCReportClient(pb.NewReportServiceClient(conn))
    ctx := context.Background()
    from := time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC)
    to := from.AddDate(0, 1, -1)
    st, err := c.GetStats(ctx, "tenant", from, to, "tok")
    if err != nil || st.TotalExpense == 0 { t.Fatalf("get stats: %v %+v", err, st) }
    top, err := c.TopCategories(ctx, "tenant", from, to, 5, "tok")
    if err != nil || len(top) == 0 { t.Fatalf("top: %v n=%d", err, len(top)) }
    if impl.sawAuth != "Bearer tok" { t.Fatalf("auth metadata not set: %q", impl.sawAuth) }
}


