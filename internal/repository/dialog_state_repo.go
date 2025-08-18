package repository

import (
	"context"
	"database/sql"
	"encoding/json"
)

type DialogState string

const (
	StateIdle                   DialogState = "idle"
	StateWaitingForEmail        DialogState = "waiting_for_email"
	StateWaitingForPassword     DialogState = "waiting_for_password"
	StateWaitingForRegisterEmail    DialogState = "waiting_for_register_email"
	StateWaitingForRegisterPassword DialogState = "waiting_for_register_password"
	StateWaitingForRegisterName     DialogState = "waiting_for_register_name"
)

type DialogStateRecord struct {
	TelegramID int64
	State      DialogState
	DraftID    *string
	Context    map[string]any
}

type DialogStateRepository interface {
	SetState(ctx context.Context, telegramID int64, state DialogState, context map[string]any, draftID *string) error
	GetState(ctx context.Context, telegramID int64) (*DialogStateRecord, error)
	ClearState(ctx context.Context, telegramID int64) error
}

type SQLiteDialogStateRepository struct {
	db *sql.DB
}

func NewSQLiteDialogStateRepository(db *sql.DB) *SQLiteDialogStateRepository {
	return &SQLiteDialogStateRepository{db: db}
}

func (r *SQLiteDialogStateRepository) SetState(ctx context.Context, telegramID int64, state DialogState, ctxMap map[string]any, draftID *string) error {
	var ctxJSON *string
	if ctxMap != nil {
		b, _ := json.Marshal(ctxMap)
		s := string(b)
		ctxJSON = &s
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO dialog_states (telegram_id, state, draft_id, context)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(telegram_id) DO UPDATE SET
			state = excluded.state,
			draft_id = excluded.draft_id,
			context = excluded.context,
			updated_at = CURRENT_TIMESTAMP
	`, telegramID, string(state), draftID, ctxJSON)
	return err
}

func (r *SQLiteDialogStateRepository) GetState(ctx context.Context, telegramID int64) (*DialogStateRecord, error) {
	row := r.db.QueryRowContext(ctx, `SELECT state, draft_id, context FROM dialog_states WHERE telegram_id = ?`, telegramID)
	var state string
	var draftID *string
	var ctxStr *string
	if err := row.Scan(&state, &draftID, &ctxStr); err != nil {
		return nil, err
	}
	var ctxMap map[string]any
	if ctxStr != nil && *ctxStr != "" {
		_ = json.Unmarshal([]byte(*ctxStr), &ctxMap)
	}
	return &DialogStateRecord{TelegramID: telegramID, State: DialogState(state), DraftID: draftID, Context: ctxMap}, nil
}

func (r *SQLiteDialogStateRepository) ClearState(ctx context.Context, telegramID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM dialog_states WHERE telegram_id = ?`, telegramID)
	return err
}


