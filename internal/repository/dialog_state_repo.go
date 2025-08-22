// Package repository contains persistence layer implementations.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
)

// DialogState is a finite state of a user in a dialog.
type DialogState string

const (
	// StateIdle means no active dialog
	StateIdle DialogState = "idle"
	// StateWaitingForEmail when user is entering email
	StateWaitingForEmail DialogState = "waiting_for_email"
	// StateWaitingForPassword when user is entering password
	StateWaitingForPassword DialogState = "waiting_for_password"
	// StateWaitingForRegisterEmail when registering email
	StateWaitingForRegisterEmail DialogState = "waiting_for_register_email"
	// StateWaitingForRegisterPassword when registering password
	StateWaitingForRegisterPassword DialogState = "waiting_for_register_password"
	// StateWaitingForRegisterName when registering name
	StateWaitingForRegisterName DialogState = "waiting_for_register_name"
	// StateConfirmingTransaction when user confirms parsed transaction
	StateConfirmingTransaction DialogState = "confirming_transaction"
	// StateWaitingForCategory when user chooses a category
	StateWaitingForCategory DialogState = "waiting_for_category"
	// OAuth States
	StateWaitingForOAuthEmail DialogState = "waiting_for_oauth_email"
	StateWaitingForOAuthCode DialogState = "waiting_for_oauth_code"
)

// DialogStateRecord is a persisted dialog state.
type DialogStateRecord struct {
	TelegramID int64
	State      DialogState
	DraftID    *string
	Context    map[string]any
}

// DialogStateRepository defines dialog state operations.
type DialogStateRepository interface {
	SetState(ctx context.Context, telegramID int64, state DialogState, context map[string]any, draftID *string) error
	GetState(ctx context.Context, telegramID int64) (*DialogStateRecord, error)
	ClearState(ctx context.Context, telegramID int64) error
}

// SQLiteDialogStateRepository stores dialog state in SQLite.
type SQLiteDialogStateRepository struct {
	db *sql.DB
}

// NewSQLiteDialogStateRepository constructs a repository.
func NewSQLiteDialogStateRepository(db *sql.DB) *SQLiteDialogStateRepository {
	return &SQLiteDialogStateRepository{db: db}
}

// SetState upserts a dialog state.
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

// GetState returns a dialog state by telegram id.
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

// ClearState deletes a dialog state by telegram id.
func (r *SQLiteDialogStateRepository) ClearState(ctx context.Context, telegramID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM dialog_states WHERE telegram_id = ?`, telegramID)
	return err
}


