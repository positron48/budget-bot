package bot

import "testing"

func TestMulticurrencyMessages(t *testing.T) {
	for _, tc := range []struct {
		text, currency string
		minor          int64
	}{
		{"20 евро продукты", "EUR", 2000}, {"20€ продукты", "EUR", 2000}, {"€20 продукты", "EUR", 2000},
		{"$20 такси", "USD", 2000}, {"20 usd такси", "USD", 2000}, {"20USD такси", "USD", 2000},
		{"20 долларов такси", "USD", 2000}, {"1 доллар кофе", "USD", 100}, {"2 доллара кофе", "USD", 200},
		{"12.05 EUR кофе", "EUR", 1205}, {"12,05евро кофе", "EUR", 1205}, {"+12.05$ возврат", "USD", 1205},
		{"100 рублей кофе", "RUB", 10000}, {"100руб. кофе", "RUB", 10000}, {"20 кофе", "", 2000},
		{"20 eurovision", "", 2000}, {"20 недолларов", "", 2000}, {"12.08 20 евро такси", "EUR", 2000},
	} {
		t.Run(tc.text, func(t *testing.T) {
			parsed, err := NewMessageParser().ParseMessage(tc.text)
			if err != nil || !parsed.IsValid || parsed.Currency != tc.currency || parsed.Amount.AmountMinor != tc.minor {
				t.Fatalf("parsed=%+v err=%v", parsed, err)
			}
		})
	}
}
func TestConflictingCurrencies(t *testing.T) {
	for _, text := range []string{"20 EUR USD кофе", "$20 евро кофе", "20 рублей $"} {
		parsed, _ := NewMessageParser().ParseMessage(text)
		if parsed.IsValid || len(parsed.Errors) == 0 {
			t.Fatalf("accepted ambiguous currency: %+v", parsed)
		}
	}
}
