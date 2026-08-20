package payments

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gpuhub/cloud/internal/paymgr"
)

var ErrNoTopup = errors.New("topup not found")

// Orchestrator routes top-up requests through paymgr and settles into wallet.
type Orchestrator struct {
	mu        sync.Mutex
	topups    map[string]*TopupRequest
	registry  *registry
	methods   *MethodStore
	walletURL string
	paymgr    *PaymgrClient
	client    *http.Client
}

// NewOrchestrator builds an orchestrator. paymgrURL delegates routing decisions.
func NewOrchestrator(walletURL, paymgrURL string) *Orchestrator {
	r := newRegistry()
	r.add(MethodMobileMoney, NewMobileMoney())
	r.add(MethodBank, NewBank("bank"))
	r.add(MethodFintech, NewBank("fintech"))
	return &Orchestrator{
		topups:    make(map[string]*TopupRequest),
		registry:  r,
		methods:   newMethodStore(),
		walletURL: walletURL,
		paymgr: &PaymgrClient{
			BaseURL: paymgrURL,
			Client:  &http.Client{Timeout: 10 * time.Second},
		},
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (o *Orchestrator) Catalog() PaymentCatalog { return DefaultCatalog() }

func (o *Orchestrator) SystemConfig() SystemPaymentConfig { return DefaultSystemConfig() }

func (o *Orchestrator) ListMethods(userID string) []SavedMethod {
	return o.methods.List(userID)
}

func (o *Orchestrator) AddMethod(m SavedMethod) SavedMethod {
	return o.methods.Add(m)
}

func (o *Orchestrator) DeleteMethod(userID, methodID string) bool {
	return o.methods.Delete(userID, methodID)
}

// StartTopup resolves the saved method, asks paymgr where to route, stores pending.
func (o *Orchestrator) StartTopup(req TopupRequest) (PaymentIntent, error) {
	if err := o.resolveTopup(&req); err != nil {
		return PaymentIntent{}, err
	}

	provider, err := o.registry.get(req.Method)
	if err != nil {
		return PaymentIntent{}, err
	}

	var intent PaymentIntent

	if o.paymgr.BaseURL != "" {
		res, err := o.paymgr.StartTopup(paymgr.TopupInput{
			ID:              req.ID,
			UserID:          req.UserID,
			Method:          string(req.Method),
			PaymentMethodID: req.PaymentMethodID,
			Subunits:        req.Subunits,
			Currency:        req.Currency,
			Carrier:         req.Carrier,
			Phone:           req.Phone,
			RailProvider:    req.RailProvider,
			AccountRef:      req.AccountRef,
		})
		if err != nil {
			return PaymentIntent{}, err
		}
		intent = routeToIntent(res, provider.Name())
		intent.Carrier = req.Carrier
		intent.Phone = req.Phone
		intent.UserPhone = req.Phone
		intent.AmountUnits = amountUnits(req.Subunits)
	} else {
		intent, err = provider.Create(req)
		if err != nil {
			return PaymentIntent{}, err
		}
	}

	req.Status = "pending"
	req.Adapter = intent.Provider
	req.Reference = intent.ChargeID
	req.CreatedAt = time.Now()
	o.mu.Lock()
	o.topups[req.ID] = &req
	o.mu.Unlock()
	return intent, nil
}

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
