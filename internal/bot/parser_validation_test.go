package bot

import "testing"

func TestMessageParser_EmptyAndNoAmount(t *testing.T) {
    p := NewMessageParser()
    // empty
    res, err := p.ParseMessage("")
    if err != nil || res == nil || res.IsValid { t.Fatalf("expected invalid empty: %+v %v", res, err) }
    // no numeric amount
    res2, err := p.ParseMessage("кофе без суммы")
    if err != nil || res2 == nil || res2.IsValid { t.Fatalf("expected invalid no amount: %+v %v", res2, err) }
}

func TestMessageParser_UnknownCurrencyTokenIgnored(t *testing.T) {
    p := NewMessageParser()
    // Unknown currency token should be treated as part of description; parse remains valid with empty currency
    res, err := p.ParseMessage("100 ABC кофе")
    if err != nil || res == nil { t.Fatalf("unexpected err: %v", err) }
    if !res.IsValid { t.Fatalf("expected valid parse") }
    if res.Currency != "" { t.Fatalf("expected empty currency, got %s", res.Currency) }
}


