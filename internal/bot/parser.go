package bot

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"budget-bot/internal/domain"
)

type MessageParser struct{ currency *CurrencyParser }

type ParsedTransaction struct {
	Type        domain.TransactionType
	Amount      *domain.Money
	Currency    string
	Description string
	OccurredAt  *time.Time
	IsValid     bool
	Errors      []string
}

func NewMessageParser() *MessageParser { return &MessageParser{currency: NewCurrencyParser()} }

var (
	amountRe   = regexp.MustCompile(`(?i)([+\-]?\d+[\.,]?\d*)`)
	dateDDMMYY = regexp.MustCompile(`\b(\d{1,2})[\./](\d{1,2})(?:[\./](\d{2,4}))?\b`)
)

func (p *MessageParser) ParseMessage(text string) (*ParsedTransaction, error) {
	original := strings.TrimSpace(text)
	result := &ParsedTransaction{IsValid: false}
	if original == "" {
		result.Errors = append(result.Errors, "empty message")
		return result, nil
	}

	// Date parsing: today/yesterday/day before yesterday
	lower := strings.ToLower(original)
	now := time.Now()
	if strings.Contains(lower, "сегодня") {
		d := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		result.OccurredAt = &d
		lower = strings.ReplaceAll(lower, "сегодня", "")
	} else if strings.Contains(lower, "вчера") {
		d := now.AddDate(0, 0, -1)
		d = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
		result.OccurredAt = &d
		lower = strings.ReplaceAll(lower, "вчера", "")
	} else if strings.Contains(lower, "позавчера") {
		d := now.AddDate(0, 0, -2)
		d = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
		result.OccurredAt = &d
		lower = strings.ReplaceAll(lower, "позавчера", "")
	}

	// DD.MM(.YYYY)
	if m := dateDDMMYY.FindStringSubmatch(lower); len(m) > 0 {
		day, _ := strconv.Atoi(m[1])
		mon, _ := strconv.Atoi(m[2])
		year := now.Year()
		if m[3] != "" {
			y, _ := strconv.Atoi(m[3])
			if y < 100 { // YY -> 20YY heuristic
				y = 2000 + y
			}
			year = y
		}
		if day >= 1 && day <= 31 && mon >= 1 && mon <= 12 {
			d := time.Date(year, time.Month(mon), day, 0, 0, 0, 0, now.Location())
			result.OccurredAt = &d
			lower = strings.Replace(lower, m[0], "", 1)
		}
	}

	// Currency (optional)
	if code, _, cleaned := p.currency.ParseCurrency(lower); code != "" {
		result.Currency = code
		lower = cleaned
	}

	// Amount (+/-)
	var sign int64 = 1
	if strings.HasPrefix(strings.TrimSpace(lower), "+") {
		sign = 1
	} else if strings.HasPrefix(strings.TrimSpace(lower), "-") {
		sign = -1
	}
	if m := amountRe.FindStringSubmatch(lower); len(m) > 0 {
		numeric := strings.ReplaceAll(m[1], ",", ".")
		if strings.Contains(numeric, ".") {
			parts := strings.SplitN(numeric, ".", 2)
			whole, _ := strconv.ParseInt(strings.ReplaceAll(parts[0], "+", ""), 10, 64)
			frac := parts[1]
			if len(frac) == 1 {
				frac = frac + "0"
			}
			if len(frac) > 2 {
				frac = frac[:2]
			}
			minor, _ := strconv.ParseInt(frac, 10, 64)
			amountMinor := sign*(whole*100) + sign*minor
			result.Amount = domain.NewMoney(amountMinor, result.Currency)
		} else {
			whole, _ := strconv.ParseInt(strings.ReplaceAll(numeric, "+", ""), 10, 64)
			result.Amount = domain.NewMoney(sign*whole*100, result.Currency)
		}
		lower = strings.Replace(lower, m[0], "", 1)
	}

	// Determine type by sign and default to expense when negative else income
	if sign < 0 {
		result.Type = domain.TransactionExpense
	} else {
		// By convention "+" is income, neutral default to expense unless plus specified.
		if strings.HasPrefix(strings.TrimSpace(text), "+") {
			result.Type = domain.TransactionIncome
		} else {
			result.Type = domain.TransactionExpense
		}
	}

	// Remaining text as description
	desc := strings.TrimSpace(lower)
	result.Description = desc

	// Validate minimal fields
	if result.Amount == nil {
		result.Errors = append(result.Errors, "amount not found")
	}
	result.IsValid = len(result.Errors) == 0
	return result, nil
}


