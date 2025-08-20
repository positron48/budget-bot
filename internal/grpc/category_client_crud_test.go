package grpc

import (
    "context"
    "net"
    "testing"

    pb "budget-bot/internal/pb/budget/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"
)

type fakeCategoryCRUDServer struct{ pb.UnimplementedCategoryServiceServer; sawAuth string }

func (s *fakeCategoryCRUDServer) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
    if md, ok := metadata.FromIncomingContext(ctx); ok {
        vals := md.Get("authorization")
        if len(vals) > 0 { s.sawAuth = vals[0] }
    }
    return &pb.CreateCategoryResponse{Category: &pb.Category{Id: "cid", Code: req.GetCode()}}, nil
}

func (s *fakeCategoryCRUDServer) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.UpdateCategoryResponse, error) {
    if md, ok := metadata.FromIncomingContext(ctx); ok {
        vals := md.Get("authorization")
        if len(vals) > 0 { s.sawAuth = vals[0] }
    }
    return &pb.UpdateCategoryResponse{Category: &pb.Category{Id: req.GetId()}}, nil
}

func (s *fakeCategoryCRUDServer) DeleteCategory(ctx context.Context, _ *pb.DeleteCategoryRequest) (*pb.DeleteCategoryResponse, error) {
    if md, ok := metadata.FromIncomingContext(ctx); ok {
        vals := md.Get("authorization")
        if len(vals) > 0 { s.sawAuth = vals[0] }
    }
    return &pb.DeleteCategoryResponse{}, nil
}

func startCategoryCRUDServer(t *testing.T) (*grpc.Server, string, *fakeCategoryCRUDServer) {
    t.Helper()
    lis, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatal(err) }
    s := grpc.NewServer()
    impl := &fakeCategoryCRUDServer{}
    pb.RegisterCategoryServiceServer(s, impl)
    go func(){ _ = s.Serve(lis) }()
    return s, lis.Addr().String(), impl
}

func TestGRPCCategoryClient_CRUD(t *testing.T) {
    srv, addr, impl := startCategoryCRUDServer(t)
    defer srv.Stop()
    conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil { t.Fatal(err) }
    defer func(){ _ = conn.Close() }()

    c := NewGRPCCategoryClient(pb.NewCategoryServiceClient(conn))
    // Create
    cat, err := c.CreateCategory(context.Background(), "tok", "code", "Имя", "ru")
    if err != nil || cat == nil || cat.ID == "" { t.Fatalf("create: %v %+v", err, cat) }
    if impl.sawAuth != "Bearer tok" { t.Fatalf("auth not set: %q", impl.sawAuth) }

    // Update
    impl.sawAuth = ""
    upd, err := c.UpdateCategoryName(context.Background(), "tok", cat.ID, "Новое имя", "ru")
    if err != nil || upd.ID != cat.ID { t.Fatalf("update: %v %+v", err, upd) }
    if impl.sawAuth != "Bearer tok" { t.Fatalf("auth not set upd: %q", impl.sawAuth) }

    // Delete
    impl.sawAuth = ""
    if err := c.DeleteCategory(context.Background(), "tok", cat.ID); err != nil { t.Fatalf("delete: %v", err) }
    if impl.sawAuth != "Bearer tok" { t.Fatalf("auth not set del: %q", impl.sawAuth) }
}


