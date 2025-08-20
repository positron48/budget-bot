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
)

type fakeFxServer struct{ pb.UnimplementedFxServiceServer; sawAuth string }

func (s *fakeFxServer) GetRate(ctx context.Context, _ *pb.GetRateRequest) (*pb.GetRateResponse, error) {
    if md, ok := metadata.FromIncomingContext(ctx); ok {
        vals := md.Get("authorization")
        if len(vals) > 0 { s.sawAuth = vals[0] }
    }
    return &pb.GetRateResponse{Rate: &pb.FxRate{RateDecimal: "2.50"}}, nil
}

func startFxServer(t *testing.T) (*grpc.Server, string, *fakeFxServer) {
    t.Helper()
    lis, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatal(err) }
    s := grpc.NewServer()
    impl := &fakeFxServer{}
    pb.RegisterFxServiceServer(s, impl)
    go func(){ _ = s.Serve(lis) }()
    return s, lis.Addr().String(), impl
}

func TestFxGRPCClient_GetRate(t *testing.T) {
    srv, addr, impl := startFxServer(t)
    defer srv.Stop()
    conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil { t.Fatal(err) }
    defer func(){ _ = conn.Close() }()
    c := NewGRPCFxClient(pb.NewFxServiceClient(conn))
    r, err := c.GetRate(context.Background(), "USD", "RUB", time.Now(), "tok")
    if err != nil { t.Fatalf("get rate: %v", err) }
    if r < 2.49 || r > 2.51 { t.Fatalf("unexpected rate: %v", r) }
    if impl.sawAuth != "Bearer tok" { t.Fatalf("auth not set: %q", impl.sawAuth) }
}


