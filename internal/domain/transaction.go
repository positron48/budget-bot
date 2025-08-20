// Package domain contains core domain models used across the bot.
package domain

import "time"

// TransactionType enumerates supported transaction types.
type TransactionType string

const (
	// TransactionExpense is a spending transaction
	TransactionExpense TransactionType = "expense"
	// TransactionIncome is an income transaction
	TransactionIncome  TransactionType = "income"
)

// TransactionDraft represents a user-entered transaction before confirmation.
type TransactionDraft struct {
	ID          string
	TelegramID  int64
	Type        TransactionType
	Amount      *Money
	Description string
	CategoryID  string
	OccurredAt  *time.Time
	CreatedAt   time.Time
}


