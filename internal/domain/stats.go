package domain

type Stats struct {
    Period       string
    TotalIncome  int64 // minor units
    TotalExpense int64 // minor units
    Currency     string
}

type CategoryTotal struct {
    CategoryID string
    Name       string
    SumMinor   int64
    Currency   string
}


