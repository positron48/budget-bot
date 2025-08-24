package grpc

import (
    "context"
    "net"
    "testing"
    "time"

    pb "budget-bot/internal/pb/budget/v1"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    status "google.golang.org/grpc/status"
    "google.golang.org/grpc/codes"
)

type errReportServer struct{ pb.UnimplementedReportServiceServer }

func (s *errReportServer) GetMonthlySummary(_ context.Context, _ *pb.GetMonthlySummaryRequest) (*pb.GetMonthlySummaryResponse, error) {
    return nil, status.Error(codes.Internal, "internal error")
}

func startErrReportServer(t *testing.T) (*grpc.Server, string) {
    t.Helper()
    lis, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatal(err) }
    s := grpc.NewServer()
    pb.RegisterReportServiceServer(s, &errReportServer{})
    go func() { _ = s.Serve(lis) }()
    return s, lis.Addr().String()
}

func TestGRPCReportClient_Error(t *testing.T) {
    srv, addr := startErrReportServer(t)
    defer srv.Stop()
    conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil { t.Fatal(err) }
    defer func(){ _ = conn.Close() }()
    c := NewGRPCReportClient(pb.NewReportServiceClient(conn), zap.NewNop())
    ctx := context.Background()
    from := time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC)
    to := from.AddDate(0, 1, -1)
    _, e := c.GetStats(ctx, "tenant", from, to, "tok")
    if e == nil { t.Fatal("expected error") }
    _, e = c.TopCategories(ctx, "tenant", from, to, 5, "tok")
    if e == nil { t.Fatal("expected error") }
}


