// Package grpc contains gRPC client facades used by the bot.
package grpc

import (
    "context"

    pb "budget-bot/internal/pb/budget/v1"
    "google.golang.org/grpc/metadata"
)

// Tenant represents a tenant returned by the backend.
type Tenant struct {
    ID   string
    Name string
}

// TenantClient exposes tenant operations.
type TenantClient interface {
    ListTenants(ctx context.Context, accessToken string) ([]*Tenant, error)
}

// FakeTenantClient is a temporary stub returning two tenants.
// FakeTenantClient is a temporary stub returning static tenants.
type FakeTenantClient struct{}

// ListTenants returns a static list of tenants.
func (f *FakeTenantClient) ListTenants(_ context.Context, _ string) ([]*Tenant, error) {
    return []*Tenant{{ID: "tenant-1", Name: "Личный"}, {ID: "tenant-2", Name: "Семья"}}, nil
}

// TenantGRPCClient calls Tenant service via gRPC.
type TenantGRPCClient struct{ client pb.TenantServiceClient }

// NewGRPCTenantClient constructs a TenantGRPCClient.
func NewGRPCTenantClient(c pb.TenantServiceClient) *TenantGRPCClient { return &TenantGRPCClient{client: c} }

// ListTenants returns a list of tenants for current user.
func (g *TenantGRPCClient) ListTenants(ctx context.Context, accessToken string) ([]*Tenant, error) {
    if accessToken != "" { ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) }
    res, err := g.client.ListMyTenants(ctx, &pb.ListMyTenantsRequest{})
    if err != nil { return nil, err }
    var out []*Tenant
    for _, m := range res.Memberships {
        out = append(out, &Tenant{ID: m.Tenant.Id, Name: m.Tenant.Name})
    }
    return out, nil
}


