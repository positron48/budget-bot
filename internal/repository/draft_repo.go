package repository

import (
    "context"
    "database/sql"
    "time"
)

type TransactionDraft struct {
    ID          string
    TelegramID  int64
    Type        string
    AmountMinor int64
    Currency    string
    Description string
    CategoryID  string
    OccurredAt  *time.Time
    CreatedAt   time.Time
}

type DraftRepository interface {
    Create(ctx context.Context, d *TransactionDraft) error
    Get(ctx context.Context, id string) (*TransactionDraft, error)
    Delete(ctx context.Context, id string) error
}

type SQLiteDraftRepository struct { db *sql.DB }

func NewSQLiteDraftRepository(db *sql.DB) *SQLiteDraftRepository { return &SQLiteDraftRepository{db: db} }

func (r *SQLiteDraftRepository) Create(ctx context.Context, d *TransactionDraft) error {
    _, err := r.db.ExecContext(ctx, `INSERT INTO transaction_drafts (id, telegram_id, type, amount_minor, currency, description, category_id, occurred_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, d.ID, d.TelegramID, d.Type, d.AmountMinor, d.Currency, d.Description, d.CategoryID, d.OccurredAt)
    return err
}

func (r *SQLiteDraftRepository) Get(ctx context.Context, id string) (*TransactionDraft, error) {
    row := r.db.QueryRowContext(ctx, `SELECT id, telegram_id, type, amount_minor, currency, description, category_id, occurred_at, created_at FROM transaction_drafts WHERE id = ?`, id)
    var d TransactionDraft
    if err := row.Scan(&d.ID, &d.TelegramID, &d.Type, &d.AmountMinor, &d.Currency, &d.Description, &d.CategoryID, &d.OccurredAt, &d.CreatedAt); err != nil {
        return nil, err
    }
    return &d, nil
}

func (r *SQLiteDraftRepository) Delete(ctx context.Context, id string) error {
    _, err := r.db.ExecContext(ctx, `DELETE FROM transaction_drafts WHERE id = ?`, id)
    return err
}


