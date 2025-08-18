package domain

import "time"

type TransactionType string

const (
	TransactionExpense TransactionType = "expense"
	TransactionIncome  TransactionType = "income"
)

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


