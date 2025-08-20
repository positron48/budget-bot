package bot

import "testing"

func TestMessageParser_FractionsAndDates(t *testing.T) {
    p := NewMessageParser()
    // two decimals with dot
    r1, _ := p.ParseMessage("12.34 такси")
    if !r1.IsValid || r1.Amount == nil || r1.Amount.AmountMinor != 1234 { t.Fatalf("12.34 -> 1234") }
    // three decimals -> truncate
    r2, _ := p.ParseMessage("12.345 такси")
    if !r2.IsValid || r2.Amount == nil || r2.Amount.AmountMinor != 1234 { t.Fatalf("12.345 -> 1234") }
    // comma as decimal separator
    r3, _ := p.ParseMessage("12,50 продукты")
    if !r3.IsValid || r3.Amount == nil || r3.Amount.AmountMinor != 1250 { t.Fatalf("12,50 -> 1250") }
}


