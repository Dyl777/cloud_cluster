// Package money keeps currency amounts in integer subunits to avoid
// floating-point drift in ledgers. One unit = one micro-currency (USD$ * 1e6).
package money

import (
	"errors"
	"strings"
)

var (
	ErrCurrencyMismatch = errors.New("currency mismatch")
	ErrInsufficient     = errors.New("insufficient balance")
)

// Money is an integer amount in the smallest bookkeeping unit.
type Money struct {
	Subunits int64  `json:"subunits"`
	Currency string `json:"currency"`
}

// Zero returns a zero amount for a currency code.
func Zero(currency string) Money {
	return Money{Currency: strings.ToUpper(currency)}
}

// FromUnits builds Money from whole currency units (e.g. dollars).
func FromUnits(units float64, currency string) Money {
	return Money{
		Subunits: int64(units * 1_000_000),
		Currency: strings.ToUpper(currency),
	}
}

// Add sums two amounts and errors on currency mismatch.
func (m Money) Add(o Money) (Money, error) {
	if m.Currency != o.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{Subunits: m.Subunits + o.Subunits, Currency: m.Currency}, nil
}

// Sub subtracts o from m and errors on mismatch.
func (m Money) Sub(o Money) (Money, error) {
	if m.Currency != o.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{Subunits: m.Subunits - o.Subunits, Currency: m.Currency}, nil
}

// IsNegative reports whether the amount is below zero.
func (m Money) IsNegative() bool { return m.Subunits < 0 }

// Float returns the amount as a decimal currency value (e.g. 25.00).
func (m Money) Float() float64 { return float64(m.Subunits) / 1_000_000 }

// ValidCurrency reports whether the code is a recognized ISO-ish code.
func ValidCurrency(code string) bool {
	c := strings.ToUpper(code)
	if len(c) != 3 {
		return false
	}
	for _, r := range c {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

// Normalize upper-cases a currency code and validates it.
// Returns an error for malformed codes.
func NormalizeCurrency(code string) (string, error) {
	c := strings.ToUpper(code)
	if !ValidCurrency(c) {
		return "", errors.New("invalid currency code: " + code)
	}
	return c, nil
}
