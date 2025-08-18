package grpc

import (
    "context"

    pb "budget-bot/internal/pb/budget/v1"
    "google.golang.org/grpc/metadata"
)

type Tenant struct {
    ID   string
    Name string
}

type TenantClient interface {
    ListTenants(ctx context.Context, accessToken string) ([]*Tenant, error)
}

// FakeTenantClient is a temporary stub returning two tenants.
type FakeTenantClient struct{}

func (f *FakeTenantClient) ListTenants(ctx context.Context, accessToken string) ([]*Tenant, error) {
    return []*Tenant{{ID: "tenant-1", Name: "Личный"}, {ID: "tenant-2", Name: "Семья"}}, nil
}

type GRPCTenantClient struct{ client pb.TenantServiceClient }

func NewGRPCTenantClient(c pb.TenantServiceClient) *GRPCTenantClient { return &GRPCTenantClient{client: c} }

func (g *GRPCTenantClient) ListTenants(ctx context.Context, accessToken string) ([]*Tenant, error) {
    if accessToken != "" { ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) }
    res, err := g.client.ListMyTenants(ctx, &pb.ListMyTenantsRequest{})
    if err != nil { return nil, err }
    var out []*Tenant
    for _, m := range res.Memberships {
        out = append(out, &Tenant{ID: m.Tenant.Id, Name: m.Tenant.Name})
    }
    return out, nil
}


