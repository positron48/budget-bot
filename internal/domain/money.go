package domain

// Money stores value in minor units (e.g., cents) with an associated ISO currency code.
type Money struct {
	AmountMinor  int64
	CurrencyCode string
}

// NewMoney constructs Money from minor units and ISO currency code.
func NewMoney(minor int64, code string) *Money {
	return &Money{AmountMinor: minor, CurrencyCode: code}
}


