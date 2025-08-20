package bot

import "testing"

func TestMessageParser_CompactCurrencyAndDateSlash(t *testing.T) {
    p := NewMessageParser()
    // amount without space before symbol (e.g., 100₽)
    r, _ := p.ParseMessage("100₽ кофе")
    if !r.IsValid || r.Amount == nil || r.Currency != "RUB" { t.Fatalf("compact currency failed: %+v", r) }
    // DD/MM/YYYY variant
    r2, _ := p.ParseMessage("12/02/2025 200 такси")
    if !r2.IsValid || r2.OccurredAt == nil || r2.Amount == nil { t.Fatalf("dd/mm/yyyy failed: %+v", r2) }
}


