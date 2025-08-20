package grpc

import (
    "testing"

    pb "budget-bot/internal/pb/budget/v1"
)

func TestMapType(t *testing.T) {
    if mapType("income") != pb.TransactionType_TRANSACTION_TYPE_INCOME { t.Fatalf("income") }
    if mapType("expense") != pb.TransactionType_TRANSACTION_TYPE_EXPENSE { t.Fatalf("expense") }
    if mapType("unknown") != pb.TransactionType_TRANSACTION_TYPE_UNSPECIFIED { t.Fatalf("unspec") }
}


