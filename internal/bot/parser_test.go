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
	
	// Test parsing "01.08" - should be August 1st in local timezone
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
	
	// Get current year to make test more flexible
	currentYear := time.Now().Year()
	
	// The time should be August 1st, 00:00 in local timezone, converted to UTC
	// Since the parser uses time.Now().Location(), we need to account for that
	localTime := time.Date(currentYear, 8, 1, 0, 0, 0, 0, time.Now().Location())
	expectedUTC := localTime.UTC()
	
	if !parsed.OccurredAt.Equal(expectedUTC) {
		t.Errorf("Expected time %v, got %v", expectedUTC, parsed.OccurredAt)
	}
	
	// Verify that the UTC time represents August 1st in local timezone
	localTimeFromUTC := parsed.OccurredAt.In(time.Now().Location())
	if localTimeFromUTC.Day() != 1 || localTimeFromUTC.Month() != time.August {
		t.Errorf("Time in local timezone should be August 1st, got %v", localTimeFromUTC)
	}
	
	t.Logf("Successfully parsed date: %v (UTC) = %v (Local)", parsed.OccurredAt, localTimeFromUTC)
}


