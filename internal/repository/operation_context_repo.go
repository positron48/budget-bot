package repository

import (
	"context"
	"database/sql"
	"time"
)

// OperationContext stores data needed for post-selection callbacks.
type OperationContext struct {
	OpID                  string
	TelegramID            int64
	TenantID              string
	TransactionID         *string
	DescriptionOriginal   string
	CategoryIDSelected    *string
	CategoryNameSelected  *string
	SelectionSource       string
	TxType                string
	AmountMinor           int64
	Currency              string
	OccurredAt            *time.Time
	CategoryListMessageID *int
	ConfirmationMessageID *int
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// OperationContextRepository persists operation contexts.
type OperationContextRepository interface {
	Create(ctx context.Context, op *OperationContext) error
	Get(ctx context.Context, opID string) (*OperationContext, error)
	UpdateSelection(ctx context.Context, opID, categoryID, categoryName, source string) error
	SetTransactionID(ctx context.Context, opID, transactionID string) error
	SetCategoryListMessageID(ctx context.Context, opID string, messageID int) error
	SetConfirmationMessageID(ctx context.Context, opID string, messageID int) error
	Delete(ctx context.Context, opID string) error
}

// SQLiteOperationContextRepository is a SQLite-backed repository.
type SQLiteOperationContextRepository struct{ db *sql.DB }

func NewSQLiteOperationContextRepository(db *sql.DB) *SQLiteOperationContextRepository {
	return &SQLiteOperationContextRepository{db: db}
}

func (r *SQLiteOperationContextRepository) Create(ctx context.Context, op *OperationContext) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO operation_contexts (
		op_id, telegram_id, tenant_id, transaction_id, description_original,
		category_id_selected, category_name_selected, selection_source,
		tx_type, amount_minor, currency, occurred_at,
		category_list_message_id, confirmation_message_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		op.OpID, op.TelegramID, op.TenantID, op.TransactionID, op.DescriptionOriginal,
		op.CategoryIDSelected, op.CategoryNameSelected, op.SelectionSource,
		op.TxType, op.AmountMinor, op.Currency, op.OccurredAt,
		op.CategoryListMessageID, op.ConfirmationMessageID,
	)
	return err
}

func (r *SQLiteOperationContextRepository) Get(ctx context.Context, opID string) (*OperationContext, error) {
	row := r.db.QueryRowContext(ctx, `SELECT
		op_id, telegram_id, tenant_id, transaction_id, description_original,
		category_id_selected, category_name_selected, selection_source,
		tx_type, amount_minor, currency, occurred_at,
		category_list_message_id, confirmation_message_id, created_at, updated_at
	FROM operation_contexts WHERE op_id = ?`, opID)
	var op OperationContext
	if err := row.Scan(
		&op.OpID, &op.TelegramID, &op.TenantID, &op.TransactionID, &op.DescriptionOriginal,
		&op.CategoryIDSelected, &op.CategoryNameSelected, &op.SelectionSource,
		&op.TxType, &op.AmountMinor, &op.Currency, &op.OccurredAt,
		&op.CategoryListMessageID, &op.ConfirmationMessageID, &op.CreatedAt, &op.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &op, nil
}

func (r *SQLiteOperationContextRepository) UpdateSelection(ctx context.Context, opID, categoryID, categoryName, source string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE operation_contexts SET category_id_selected = ?, category_name_selected = ?, selection_source = ?, updated_at = CURRENT_TIMESTAMP WHERE op_id = ?`, categoryID, categoryName, source, opID)
	return err
}

func (r *SQLiteOperationContextRepository) SetTransactionID(ctx context.Context, opID, transactionID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE operation_contexts SET transaction_id = ?, updated_at = CURRENT_TIMESTAMP WHERE op_id = ?`, transactionID, opID)
	return err
}

func (r *SQLiteOperationContextRepository) SetCategoryListMessageID(ctx context.Context, opID string, messageID int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE operation_contexts SET category_list_message_id = ?, updated_at = CURRENT_TIMESTAMP WHERE op_id = ?`, messageID, opID)
	return err
}

func (r *SQLiteOperationContextRepository) SetConfirmationMessageID(ctx context.Context, opID string, messageID int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE operation_contexts SET confirmation_message_id = ?, updated_at = CURRENT_TIMESTAMP WHERE op_id = ?`, messageID, opID)
	return err
}

func (r *SQLiteOperationContextRepository) Delete(ctx context.Context, opID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM operation_contexts WHERE op_id = ?`, opID)
	return err
}
