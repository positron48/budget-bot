package bot

import (
	"testing"
)

func TestMessageParser_ParseMessage_SimpleExpense(t *testing.T) {
	p := NewMessageParser()
	res, err := p.ParseMessage("1000 продукты")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsValid {
		t.Fatalf("expected valid parse")
	}
	if res.Amount == nil || res.Amount.AmountMinor != 100000 {
		t.Fatalf("expected 100000 minor units, got %+v", res.Amount)
	}
	if res.Description == "" {
		t.Fatalf("expected non-empty description")
	}
}

func TestMessageParser_ParseMessage_IncomePlus(t *testing.T) {
	p := NewMessageParser()
	res, err := p.ParseMessage("+50000 зарплата")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsValid {
		t.Fatalf("expected valid parse")
	}
	if res.Amount == nil || res.Amount.AmountMinor != 5000000 {
		t.Fatalf("expected 5000000 minor units, got %+v", res.Amount)
	}
}

func TestMessageParser_ParseMessage_WithDate(t *testing.T) {
	p := NewMessageParser()
	res, err := p.ParseMessage("01.12 5000 подарок")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsValid {
		t.Fatalf("expected valid parse")
	}
	if res.OccurredAt == nil {
		t.Fatalf("expected occurredAt parsed")
	}
}


