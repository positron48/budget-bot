// Package bot contains the core Telegram bot business logic.
package bot

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
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

var currencyTokenRe = regexp.MustCompile(`(?i)₽|\$|€|£|¥|rub|usd|eur|gbp|jpy|евро|доллар(?:ами|ов|а|ы|у|ом|е)?|бакс(?:ами|ов|а|ы|у|ом|е)?|рубл(?:ями|ей|я|ь|ю|ём|ем|ях)|руб\.?`)

// ParseCurrency returns a normalized ISO code. Different currencies in one
// message are ambiguous and return the deliberately invalid CONFLICT code.
func (cp *CurrencyParser) ParseCurrency(text string) (code string, matched string, cleaned string) {
	indices := currencyTokenRe.FindAllStringIndex(text, -1)
	cleaned = text
	valid := make([][]int, 0, len(indices))
	for _, loc := range indices {
		token := text[loc[0]:loc[1]]
		c := cp.symbolToCode[token]
		if c == "" {
			// Unicode letter boundaries, allowing currency next to a numeric amount.
			before, _ := utf8.DecodeLastRuneInString(text[:loc[0]])
			after, _ := utf8.DecodeRuneInString(text[loc[1]:])
			if unicode.IsLetter(before) || unicode.IsLetter(after) || before == '_' || after == '_' {
				continue
			}
			lower := strings.ToLower(token)
			switch {
			case lower == "евро":
				c = "EUR"
			case strings.HasPrefix(lower, "доллар"), strings.HasPrefix(lower, "бакс"):
				c = "USD"
			case strings.HasPrefix(lower, "руб"):
				c = "RUB"
			default:
				c = strings.ToUpper(token)
			}
		}
		if code != "" && code != c {
			return "CONFLICT", "", text
		}
		code, matched = c, token
		valid = append(valid, loc)
	}
	for i := len(valid) - 1; i >= 0; i-- {
		loc := valid[i]
		cleaned = cleaned[:loc[0]] + " " + cleaned[loc[1]:]
	}
	return code, matched, strings.TrimSpace(cleaned)
}

// ValidateCurrency checks if code looks like an ISO 4217 currency.
func (cp *CurrencyParser) ValidateCurrency(code string) bool {
	if code == "" {
		return false
	}
	_, ok := cp.codeToSymbol[strings.ToUpper(code)]
	return ok
}
