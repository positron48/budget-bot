// Package domain contains core domain models used across the bot.
package domain

// Stats aggregates totals for a given period.
type Stats struct {
    Period       string
    TotalIncome  int64 // minor units
    TotalExpense int64 // minor units
    Currency     string
}

// CategoryTotal is a per-category total for a period.
type CategoryTotal struct {
    CategoryID string
    Name       string
    SumMinor   int64
    Currency   string
}


