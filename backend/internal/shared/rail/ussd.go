package rail

import (
	"errors"
	"strings"
)

var ErrUnknownCarrier = errors.New("unknown carrier")

// Carrier is a mobile-money operator with a USSD template.
type Carrier struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	USSDTemplate string `json:"ussd_template"`
	SpecialDial  bool   `json:"special_dial,omitempty"` // needs mobileVM dial-up path
}

// Carriers supported (Orange, MTN).
var Carriers = []Carrier{
	{ID: "mtn", Name: "MTN Mobile Money", USSDTemplate: "*126*16*{phone}*{amount}#"},
	{ID: "orange", Name: "Orange Money", USSDTemplate: "*144*16*{phone}*{amount}#"},
	{ID: "orange_bank", Name: "Orange Bank→MM", USSDTemplate: "*144*99*{phone}*{amount}#", SpecialDial: true},
}

func carrierByID(id string) (Carrier, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	for _, c := range Carriers {
		if c.ID == id {
			return c, true
		}
	}
	return Carrier{}, false
}

// BuildUSSD fills a carrier template. phone is the collection/recipient number
// embedded in the dial code (platform destination for top-ups).
func BuildUSSD(carrierID, phone string, amount int64) (string, error) {
	c, ok := carrierByID(carrierID)
	if !ok {
		return "", ErrUnknownCarrier
	}
	digits := digitsOnly(phone)
	if digits == "" {
		return "", errors.New("phone required")
	}
	if amount <= 0 {
		return "", errors.New("amount must be positive")
	}
	code := c.USSDTemplate
	code = strings.ReplaceAll(code, "{phone}", digits)
	code = strings.ReplaceAll(code, "{amount}", formatInt(amount))
	return code, nil
}

func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func PhoneSuffixMatch(a, b string) bool {
	a, b = digitsOnly(a), digitsOnly(b)
	if a == b {
		return true
	}
	if len(a) >= len(b) {
		return strings.HasSuffix(a, b)
	}
	return strings.HasSuffix(b, a)
}
