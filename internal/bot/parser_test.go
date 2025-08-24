package bot

import (
	"testing"
	"time"
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

func TestMessageParser_NegativeDecimalExpense(t *testing.T) {
    p := NewMessageParser()
    res, err := p.ParseMessage("-123.45 кофе")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !res.IsValid {
        t.Fatalf("expected valid parse, errors: %+v", res.Errors)
    }
    if res.Amount == nil || res.Amount.AmountMinor != 12345 {
        t.Fatalf("expected 12345 minor units, got %+v", res.Amount)
    }
}

func TestMessageParser_EmojiAndCurrencyCodeLowercase(t *testing.T) {
    p := NewMessageParser()
    res, err := p.ParseMessage("вчера +100 usd кофе ☕")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !res.IsValid {
        t.Fatalf("expected valid parse, got errors: %+v", res.Errors)
    }
    if res.Currency != "USD" {
        t.Fatalf("expected currency USD, got %s", res.Currency)
    }
    if res.OccurredAt == nil {
        t.Fatalf("expected occurredAt parsed for 'вчера'")
    }
}

func TestMessageParser_DateWithSlashAndCurrencyCode(t *testing.T) {
    p := NewMessageParser()
    res, err := p.ParseMessage("01/12 200 EUR подарок 🎁")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !res.IsValid {
        t.Fatalf("expected valid parse, got errors: %+v", res.Errors)
    }
    if res.Amount == nil || res.Amount.AmountMinor != 20000 {
        t.Fatalf("expected 20000 minor units, got %+v", res.Amount)
    }
    if res.Currency != "EUR" {
        t.Fatalf("expected EUR, got %s", res.Currency)
    }
    if res.OccurredAt == nil {
        t.Fatalf("expected occurredAt for 01/12")
    }
}

func TestMessageParser_DateParsing_Timezone(t *testing.T) {
	parser := NewMessageParser()
	
	// Test with a specific timezone (Moscow, UTC+3)
	moscowLoc := time.FixedZone("MSK", 3*60*60)
	
	// Test parsing "01.08" - should be August 1st in Moscow timezone
	parsed, err := parser.ParseMessage("01.08 1000 продукты")
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}
	
	if !parsed.IsValid {
		t.Fatalf("Parsed transaction is not valid: %v", parsed.Errors)
	}
	
	if parsed.OccurredAt == nil {
		t.Fatal("OccurredAt is nil")
	}
	
	// The time should be August 1st, 00:00 in Moscow timezone, converted to UTC
	expectedUTC := time.Date(2025, 8, 1, 0, 0, 0, 0, moscowLoc).UTC()
	
	if !parsed.OccurredAt.Equal(expectedUTC) {
		t.Errorf("Expected time %v, got %v", expectedUTC, parsed.OccurredAt)
	}
	
	// Verify that the UTC time represents August 1st in Moscow timezone
	moscowTime := parsed.OccurredAt.In(moscowLoc)
	if moscowTime.Day() != 1 || moscowTime.Month() != time.August {
		t.Errorf("Time in Moscow timezone should be August 1st, got %v", moscowTime)
	}
	
	t.Logf("Successfully parsed date: %v (UTC) = %v (Moscow)", parsed.OccurredAt, moscowTime)
}


