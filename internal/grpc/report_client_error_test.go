package grpc

import (
    "context"
    "net"
    "testing"
    "time"

    pb "budget-bot/internal/pb/budget/v1"
    "google.golang.org/grpc"
    status "google.golang.org/grpc/status"
)

type errReportServer struct{ pb.UnimplementedReportServiceServer }

func (s *errReportServer) GetMonthlySummary(ctx context.Context, r *pb.GetMonthlySummaryRequest) (*pb.GetMonthlySummaryResponse, error) {
    return nil, status.Error(13, "backend error")
}

func TestGRPCReportClient_GetStats_Error(t *testing.T) {
    lis, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatal(err) }
    srv := grpc.NewServer()
    pb.RegisterReportServiceServer(srv, &errReportServer{})
    go func(){ _ = srv.Serve(lis) }()
    defer srv.Stop()

    conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
    if err != nil { t.Fatal(err) }
    defer func(){ _ = conn.Close() }()
    c := NewGRPCReportClient(pb.NewReportServiceClient(conn))
    _, e := c.GetStats(context.Background(), "tenant", time.Now(), time.Now(), "tok")
    if e == nil { t.Fatalf("expected error") }
}


