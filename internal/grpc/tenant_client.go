package grpc

import "context"

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


