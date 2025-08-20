package grpc

import (
    "context"
    "net"
    "testing"

    pb "budget-bot/internal/pb/budget/v1"
    "google.golang.org/grpc"
)

type fakeTenantServer struct{ pb.UnimplementedTenantServiceServer }

func (s *fakeTenantServer) ListMyTenants(ctx context.Context, r *pb.ListMyTenantsRequest) (*pb.ListMyTenantsResponse, error) {
    return &pb.ListMyTenantsResponse{Memberships: []*pb.TenantMembership{{Tenant: &pb.Tenant{Id: "t1", Name: "Личный"}}}}, nil
}

func startTenantServer(t *testing.T) (*grpc.Server, string) {
    t.Helper()
    lis, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatal(err) }
    s := grpc.NewServer()
    pb.RegisterTenantServiceServer(s, &fakeTenantServer{})
    go func() { _ = s.Serve(lis) }()
    return s, lis.Addr().String()
}

func TestGRPCTenantClient_ListTenants(t *testing.T) {
    srv, addr := startTenantServer(t)
    defer srv.Stop()
    conn, err := grpc.Dial(addr, grpc.WithInsecure())
    if err != nil { t.Fatal(err) }
    defer func(){ _ = conn.Close() }()
    c := NewGRPCTenantClient(pb.NewTenantServiceClient(conn))
    list, err := c.ListTenants(context.Background(), "tok")
    if err != nil || len(list) == 0 { t.Fatalf("tenants: %v n=%d", err, len(list)) }
}


