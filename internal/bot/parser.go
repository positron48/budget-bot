package bot

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"budget-bot/internal/domain"
)

type MessageParser struct{ currency *CurrencyParser }

type ValidationError struct {
	Field   string
	Message string
}

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

	// Amount (+/-) and type inference from the matched token
	if m := amountRe.FindStringSubmatch(lower); len(m) > 0 {
		token := strings.TrimSpace(m[1])
		var sign int64 = 1
		if strings.HasPrefix(token, "-") {
			sign = -1
		} else if strings.HasPrefix(token, "+") {
			sign = 1
		}
		numeric := strings.ReplaceAll(token, ",", ".")
		numericAbs := strings.TrimLeft(numeric, "+-")
		if strings.Contains(numericAbs, ".") {
			parts := strings.SplitN(numericAbs, ".", 2)
			wholeAbs, _ := strconv.ParseInt(parts[0], 10, 64)
			frac := parts[1]
			if len(frac) == 1 {
				frac = frac + "0"
			}
			if len(frac) > 2 {
				frac = frac[:2]
			}
			minorAbs, _ := strconv.ParseInt(frac, 10, 64)
			amountMinor := (wholeAbs*100)+minorAbs
			result.Amount = domain.NewMoney(amountMinor, result.Currency)
		} else {
			wholeAbs, _ := strconv.ParseInt(numericAbs, 10, 64)
			result.Amount = domain.NewMoney(wholeAbs*100, result.Currency)
		}
		lower = strings.Replace(lower, m[0], "", 1)

		// Determine type by sign; default to expense
		if sign < 0 {
			result.Type = domain.TransactionExpense
		} else if strings.HasPrefix(token, "+") {
			result.Type = domain.TransactionIncome
		} else {
			result.Type = domain.TransactionExpense
		}
	}

	// Remaining text as description
	desc := strings.TrimSpace(lower)
	result.Description = desc

	// Validate
	verrs := p.Validate(result)
	if len(verrs) > 0 {
		for _, e := range verrs {
			result.Errors = append(result.Errors, e.Field+": "+e.Message)
		}
		result.IsValid = false
	} else {
		result.IsValid = true
	}
	return result, nil
}

// Validate performs basic validation of the parsed transaction.
func (p *MessageParser) Validate(parsed *ParsedTransaction) []ValidationError {
	var errs []ValidationError
	if parsed == nil {
		return []ValidationError{{Field: "_", Message: "nil parsed transaction"}}
	}
	if parsed.Amount == nil || parsed.Amount.AmountMinor == 0 {
		errs = append(errs, ValidationError{Field: "amount", Message: "not found"})
	}
	if parsed.Currency != "" && !p.currency.ValidateCurrency(parsed.Currency) {
		errs = append(errs, ValidationError{Field: "currency", Message: "invalid"})
	}
	return errs
}


