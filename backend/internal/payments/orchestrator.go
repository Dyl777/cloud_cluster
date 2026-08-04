package payments

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"
)

var ErrNoTopup = errors.New("topup not found")

// Orchestrator routes top-up requests to adapters and settles them into
// the wallet service over REST.
type Orchestrator struct {
	mu        sync.Mutex
	topups    map[string]*TopupRequest
	registry  *registry
	walletURL string // e.g. http://wallet:8082
	client    *http.Client
}

// NewOrchestrator builds an orchestrator with all providers registered.
func NewOrchestrator(walletURL string) *Orchestrator {
	r := newRegistry()
	r.add(MethodMobileMoney, NewMobileMoney(&CarrierBridge{UseVMProxy: true}))
	r.add(MethodBank, NewBank("bank"))
	r.add(MethodFintech, NewBank("fintech"))
	return &Orchestrator{
		topups:    make(map[string]*TopupRequest),
		registry:  r,
		walletURL: walletURL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}
}

// StartTopup creates a payment intent on the routed provider.
func (o *Orchestrator) StartTopup(req TopupRequest) (PaymentIntent, error) {
	provider, err := o.registry.get(req.Method)
	if err != nil {
		return PaymentIntent{}, err
	}
	intent, err := provider.Create(req)
	if err != nil {
		return PaymentIntent{}, err
	}
	req.Status = "pending"
	req.Provider = intent.Provider
	req.Reference = intent.ChargeID
	o.mu.Lock()
	o.topups[req.ID] = &req
	o.mu.Unlock()
	return intent, nil
}

// ConfirmTopup resolves the charge and credits the wallet.
func (o *Orchestrator) ConfirmTopup(topupID string) error {
	o.mu.Lock()
	req, ok := o.topups[topupID]
	o.mu.Unlock()
	if !ok {
		return ErrNoTopup
	}
	provider, err := o.registry.get(req.Method)
	if err != nil {
		return err
	}
	confirmed, err := provider.Confirm(req.Reference)
	if err != nil || !confirmed {
		req.Status = "failed"
		return err
	}
	req.Status = "confirmed"
	req.CompletedAt = time.Now()
	return o.creditWallet(req)
}

// Get returns a stored top-up by id.
func (o *Orchestrator) Get(topupID string) (*TopupRequest, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	t, ok := o.topups[topupID]
	return t, ok
}

func (o *Orchestrator) creditWallet(req *TopupRequest) error {
	body, _ := json.Marshal(map[string]any{
		"subunits": req.Subunits,
		"currency": req.Currency,
		"ref":      req.ID,
	})
	url := o.walletURL + "/wallets/" + req.UserID + "/credit"
	resp, err := o.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
