// Package bot contains the core Telegram bot business logic.
package bot

import (
    "regexp"
    "strings"
)

// CurrencyParser validates and normalizes currency codes.
type CurrencyParser struct {
    symbolToCode map[string]string
    codeToSymbol map[string]string
}

// NewCurrencyParser constructs a CurrencyParser.
func NewCurrencyParser() *CurrencyParser {
    s2c := map[string]string{
        "₽": "RUB",
        "$": "USD",
        "€": "EUR",
        "£": "GBP",
        "¥": "JPY",
    }
    c2s := map[string]string{
        "RUB": "₽",
        "USD": "$",
        "EUR": "€",
        "GBP": "£",
        "JPY": "¥",
    }
    return &CurrencyParser{symbolToCode: s2c, codeToSymbol: c2s}
}

var currencyCodeRe = regexp.MustCompile(`\b(RUB|USD|EUR|GBP|JPY)\b`)

// ParseCurrency returns ISO code and the matched token, and the cleaned text without that token.
func (cp *CurrencyParser) ParseCurrency(text string) (code string, matched string, cleaned string) {
    t := text
    // Look for symbols anywhere
    for sym, c := range cp.symbolToCode {
        if strings.Contains(t, sym) {
            cleaned = strings.ReplaceAll(t, sym, "")
            return c, sym, strings.TrimSpace(cleaned)
        }
    }
    // Look for 3-letter codes (case-insensitive)
    upper := strings.ToUpper(t)
    if loc := currencyCodeRe.FindStringIndex(upper); loc != nil {
        matched = upper[loc[0]:loc[1]]
        cleaned = strings.TrimSpace(t[:loc[0]] + t[loc[1]:])
        return matched, matched, cleaned
    }
    return "", "", t
}

// ValidateCurrency checks if code looks like an ISO 4217 currency.
func (cp *CurrencyParser) ValidateCurrency(code string) bool {
    if code == "" {
        return false
    }
    _, ok := cp.codeToSymbol[strings.ToUpper(code)]
    return ok
}


