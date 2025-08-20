package grpc

import (
    "context"
    "net"
    "testing"

    pb "budget-bot/internal/pb/budget/v1"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// fake auth server
type fakeAuthServer struct{ pb.UnimplementedAuthServiceServer }

func (f *fakeAuthServer) Register(_ context.Context, _ *pb.RegisterRequest) (*pb.RegisterResponse, error) {
    return &pb.RegisterResponse{
        User:   &pb.User{Id: "user-1"},
        Tenant: &pb.Tenant{Id: "tenant-1"},
        Tokens: &pb.TokenPair{AccessToken: "a1", RefreshToken: "r1"},
    }, nil
}

func (f *fakeAuthServer) Login(_ context.Context, _ *pb.LoginRequest) (*pb.LoginResponse, error) {
    return &pb.LoginResponse{
        Tokens: &pb.TokenPair{AccessToken: "a2", RefreshToken: "r2"},
        Memberships: []*pb.TenantMembership{{Tenant: &pb.Tenant{Id: "tenant-1"}, IsDefault: true}},
    }, nil
}

func (f *fakeAuthServer) RefreshToken(_ context.Context, _ *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
    return &pb.RefreshTokenResponse{Tokens: &pb.TokenPair{AccessToken: "a3", RefreshToken: "r3"}}, nil
}

func startAuthTestServer(t *testing.T) (*grpc.Server, string) {
    t.Helper()
    lis, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatal(err) }
    s := grpc.NewServer()
    pb.RegisterAuthServiceServer(s, &fakeAuthServer{})
    go func() { _ = s.Serve(lis) }()
    return s, lis.Addr().String()
}

func TestAuthClient_Flow(t *testing.T) {
    srv, addr := startAuthTestServer(t)
    defer srv.Stop()

    conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil { t.Fatal(err) }
    defer func(){ _ = conn.Close() }()

    log, _ := zap.NewDevelopment()
    c := NewAuthClient(pb.NewAuthServiceClient(conn), log)

    ctx := context.Background()
    // Register
    _, _, at, rt, _, _, err := c.Register(ctx, "e@ex", "p", "n")
    if err != nil || at == "" || rt == "" { t.Fatalf("register failed: %v", err) }
    // Login
    _, tenantID, at2, rt2, _, _, err := c.Login(ctx, "e@ex", "p")
    if err != nil || at2 == "" || rt2 == "" || tenantID == "" { t.Fatalf("login failed: %v", err) }
    // Refresh
    at3, rt3, _, _, err := c.RefreshToken(ctx, "r2")
    if err != nil || at3 == "" || rt3 == "" { t.Fatalf("refresh failed: %v", err) }
}


