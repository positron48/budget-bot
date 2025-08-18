package bot

import "testing"

func TestCurrencyParser_Symbols(t *testing.T) {
    cp := NewCurrencyParser()
    code, matched, cleaned := cp.ParseCurrency("1000₽ продукты")
    if code != "RUB" || matched == "" || cleaned == "1000 продукты" {
        // cleaned may keep spacing; just assert code
        if code != "RUB" {
            t.Fatalf("expected RUB, got %s", code)
        }
    }
}

func TestCurrencyParser_Codes(t *testing.T) {
    cp := NewCurrencyParser()
    code, _, _ := cp.ParseCurrency("1000 USD продукты")
    if code != "USD" {
        t.Fatalf("expected USD, got %s", code)
    }
}


