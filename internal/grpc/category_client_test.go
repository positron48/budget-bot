package grpc

import (
    "context"
    "net"
    "testing"

    pb "budget-bot/internal/pb/budget/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
)

type fakeCategoryServer struct{ pb.UnimplementedCategoryServiceServer; sawAuth string }

func (s *fakeCategoryServer) ListCategories(ctx context.Context, r *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
    if md, ok := metadata.FromIncomingContext(ctx); ok {
        vals := md.Get("authorization")
        if len(vals) > 0 { s.sawAuth = vals[0] }
    }
    return &pb.ListCategoriesResponse{Categories: []*pb.Category{
        {Id: "c1", Code: "food", Translations: []*pb.CategoryTranslation{{Locale: "ru", Name: "Питание"}}},
        {Id: "c2", Code: "transport"},
    }}, nil
}

func startCategoryServer(t *testing.T) (*grpc.Server, string, *fakeCategoryServer) {
    t.Helper()
    lis, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatal(err) }
    s := grpc.NewServer()
    impl := &fakeCategoryServer{}
    pb.RegisterCategoryServiceServer(s, impl)
    go func() { _ = s.Serve(lis) }()
    return s, lis.Addr().String(), impl
}

func TestGRPCCategoryClient_ListCategories(t *testing.T) {
    srv, addr, impl := startCategoryServer(t)
    defer srv.Stop()
    conn, err := grpc.Dial(addr, grpc.WithInsecure())
    if err != nil { t.Fatal(err) }
    defer conn.Close()

    c := NewGRPCCategoryClient(pb.NewCategoryServiceClient(conn))
    got, err := c.ListCategories(context.Background(), "tenant", "tok", "ru")
    if err != nil { t.Fatalf("err: %v", err) }
    if len(got) != 2 { t.Fatalf("want 2, got %d", len(got)) }
    if got[0].Name != "Питание" || got[1].Name != "transport" { t.Fatalf("unexpected names: %+v", got) }
    if impl.sawAuth != "Bearer tok" { t.Fatalf("auth metadata not set: %q", impl.sawAuth) }
}


