package simbridge

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var ErrNoSIM = errors.New("no matching SIM found")

// SIMSlot is a SIM or modem line on the local device.
type SIMSlot struct {
	Slot    int    `json:"slot"`
	Number  string `json:"number"`
	Carrier string `json:"carrier,omitempty"`
}

// DialResult is returned after entering a USSD code on the matched SIM.
type DialResult struct {
	USSD    string `json:"ussd"`
	Phone   string `json:"phone"`
	Carrier string `json:"carrier"`
	SIMSlot int    `json:"sim_slot"`
	Message string `json:"message"`
}

// Service manages local SIM inventory and USSD dialing.
type Service struct {
	mu   sync.Mutex
	sims []SIMSlot
}

// New returns a simbridge service, optionally seeded with SIMs from env.
func New(sims []SIMSlot) *Service {
	return &Service{sims: sims}
}

// SIMs returns the registered SIM inventory.
func (s *Service) SIMs() []SIMSlot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]SIMSlot(nil), s.sims...)
}

// RegisterSIMs replaces the local SIM list (phone or USB modem discovery).
func (s *Service) RegisterSIMs(sims []SIMSlot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sims = append([]SIMSlot(nil), sims...)
}

// Dial builds USSD if needed, picks the matching SIM, and enters the code.
func (s *Service) Dial(carrier, phone string, amount int64, ussd string) (DialResult, error) {
	if ussd == "" {
		var err error
		ussd, err = BuildUSSD(carrier, phone, amount)
		if err != nil {
			return DialResult{}, err
		}
	}

	s.mu.Lock()
	slot, sim, ok := pickSIM(s.sims, phone, carrier)
	s.mu.Unlock()
	if !ok {
		return DialResult{}, ErrNoSIM
	}

	message := enterUSSD(ussd, slot, sim)
	return DialResult{
		USSD:    ussd,
		Phone:   phone,
		Carrier: carrier,
		SIMSlot: slot,
		Message: message,
	}, nil
}

func pickSIM(sims []SIMSlot, phone, carrier string) (int, SIMSlot, bool) {
	if len(sims) == 0 {
		return 0, SIMSlot{}, false
	}
	phoneDigits := digitsOnly(phone)
	carrier = strings.ToLower(strings.TrimSpace(carrier))

	for _, sim := range sims {
		if phoneDigits != "" && suffixMatch(digitsOnly(sim.Number), phoneDigits) {
			return sim.Slot, sim, true
		}
	}
	for _, sim := range sims {
		if carrier != "" && strings.ToLower(sim.Carrier) == carrier {
			return sim.Slot, sim, true
		}
	}
	if len(sims) == 1 {
		return sims[0].Slot, sims[0], true
	}
	return 0, SIMSlot{}, false
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

func suffixMatch(a, b string) bool {
	if a == b {
		return true
	}
	if len(a) >= len(b) {
		return strings.HasSuffix(a, b)
	}
	return strings.HasSuffix(b, a)
}

// enterUSSD sends the code on the given SIM slot. Production builds use
// TelephonyManager (Android) or AT commands on a USB modem; simulated here.
func enterUSSD(code string, slot int, sim SIMSlot) string {
	return fmt.Sprintf(
		"Confirm payment via %s on SIM %d (%s). Enter PIN to approve. Code: %s",
		strings.ToUpper(sim.Carrier), slot, maskNumber(sim.Number), code,
	)
}

func maskNumber(n string) string {
	d := digitsOnly(n)
	if len(d) <= 4 {
		return d
	}
	return "•••" + d[len(d)-4:]
}
