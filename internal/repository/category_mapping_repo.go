package repository

import (
	"context"
	"database/sql"
)

type CategoryMapping struct {
	ID         string
	TenantID   string
	Keyword    string
	CategoryID string
	Priority   int
}

type CategoryMappingRepository interface {
	AddMapping(ctx context.Context, m *CategoryMapping) error
	RemoveMapping(ctx context.Context, tenantID string, keyword string) error
	FindMapping(ctx context.Context, tenantID string, keyword string) (*CategoryMapping, error)
	ListMappings(ctx context.Context, tenantID string) ([]*CategoryMapping, error)
}

type SQLiteCategoryMappingRepository struct {
	db *sql.DB
}

func NewSQLiteCategoryMappingRepository(db *sql.DB) *SQLiteCategoryMappingRepository {
	return &SQLiteCategoryMappingRepository{db: db}
}

func (r *SQLiteCategoryMappingRepository) AddMapping(ctx context.Context, m *CategoryMapping) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO category_mappings (id, tenant_id, keyword, category_id, priority)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(tenant_id, keyword) DO UPDATE SET
			category_id = excluded.category_id,
			priority = excluded.priority
	`, m.ID, m.TenantID, m.Keyword, m.CategoryID, m.Priority)
	return err
}

func (r *SQLiteCategoryMappingRepository) RemoveMapping(ctx context.Context, tenantID string, keyword string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM category_mappings WHERE tenant_id = ? AND keyword = ?`, tenantID, keyword)
	return err
}

func (r *SQLiteCategoryMappingRepository) FindMapping(ctx context.Context, tenantID string, keyword string) (*CategoryMapping, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, tenant_id, keyword, category_id, priority FROM category_mappings WHERE tenant_id = ? AND keyword = ?`, tenantID, keyword)
	var m CategoryMapping
	if err := row.Scan(&m.ID, &m.TenantID, &m.Keyword, &m.CategoryID, &m.Priority); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *SQLiteCategoryMappingRepository) ListMappings(ctx context.Context, tenantID string) ([]*CategoryMapping, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, tenant_id, keyword, category_id, priority FROM category_mappings WHERE tenant_id = ? ORDER BY priority DESC, keyword ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*CategoryMapping
	for rows.Next() {
		var m CategoryMapping
		if err := rows.Scan(&m.ID, &m.TenantID, &m.Keyword, &m.CategoryID, &m.Priority); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, rows.Err()
}


