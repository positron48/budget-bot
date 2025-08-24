// Package grpc contains gRPC client facades used by the bot.
package grpc

import (
    "context"
    "math"

    pb "budget-bot/internal/pb/budget/v1"
    "google.golang.org/grpc/metadata"
    "go.uber.org/zap"
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
type TenantGRPCClient struct{ 
    client pb.TenantServiceClient 
    logger *zap.Logger
}

// NewGRPCTenantClient constructs a TenantGRPCClient.
func NewGRPCTenantClient(c pb.TenantServiceClient, logger *zap.Logger) *TenantGRPCClient { 
    return &TenantGRPCClient{client: c, logger: logger} 
}

// ListTenants returns a list of tenants for current user.
func (g *TenantGRPCClient) ListTenants(ctx context.Context, accessToken string) ([]*Tenant, error) {
    g.logger.Debug("ListTenants request", 
        zap.String("accessToken", accessToken[:int(math.Min(float64(len(accessToken)), 10))] + "..."))
    
    if accessToken != "" { ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) }
    
    req := &pb.ListMyTenantsRequest{}
    g.logger.Debug("ListTenants gRPC request")
    
    res, err := g.client.ListMyTenants(ctx, req)
    if err != nil { 
        g.logger.Error("ListTenants gRPC call failed", zap.Error(err))
        return nil, err 
    }
    
    g.logger.Debug("ListTenants gRPC response", 
        zap.Int("membershipsCount", len(res.Memberships)))
    
    var out []*Tenant
    for _, m := range res.Memberships {
        out = append(out, &Tenant{ID: m.Tenant.Id, Name: m.Tenant.Name})
    }
    
    g.logger.Debug("ListTenants processed", 
        zap.Int("tenantsReturned", len(out)))
    
    return out, nil
}


