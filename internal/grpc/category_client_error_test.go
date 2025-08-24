package grpc

import (
    "context"
    "net"
    "testing"

    "budget-bot/internal/domain"
    pb "budget-bot/internal/pb/budget/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    status "google.golang.org/grpc/status"
    "go.uber.org/zap"
)

type errCategoryServer struct{ pb.UnimplementedCategoryServiceServer }

func (s *errCategoryServer) ListCategories(_ context.Context, _ *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
    return nil, status.Error(13, "backend error")
}

func TestGRPCCategoryClient_ListCategories_Error(t *testing.T) {
    lis, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatal(err) }
    srv := grpc.NewServer()
    pb.RegisterCategoryServiceServer(srv, &errCategoryServer{})
    go func(){ _ = srv.Serve(lis) }()
    defer srv.Stop()

    conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil { t.Fatal(err) }
    defer func(){ _ = conn.Close() }()
    c := NewGRPCCategoryClient(pb.NewCategoryServiceClient(conn), zap.NewNop())
    	_, e := c.ListCategories(context.Background(), "tenant", "tok", domain.TransactionExpense)
    if e == nil { t.Fatalf("expected error") }
}


