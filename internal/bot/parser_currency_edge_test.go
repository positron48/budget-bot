package bot

import "testing"

func TestMessageParser_NegativeWithCurrencyAndOrder(t *testing.T) {
    p := NewMessageParser()
    r, _ := p.ParseMessage("-123,4 RUB кофе")
    if !r.IsValid || r.Amount == nil || r.Currency != "RUB" { t.Fatalf("neg with currency failed: %+v", r) }
    r2, _ := p.ParseMessage("RUB 200 такси")
    if !r2.IsValid || r2.Currency != "RUB" { t.Fatalf("currency before amount failed: %+v", r2) }
}


