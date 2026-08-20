package payments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gpuhub/cloud/internal/paymgr"
)

// PaymgrClient delegates routing to the backend payment manager.
type PaymgrClient struct {
	BaseURL string
	Client  *http.Client
}

func (c *PaymgrClient) StartTopup(in paymgr.TopupInput) (*paymgr.RouteResult, error) {
	if c.BaseURL == "" {
		return nil, fmt.Errorf("paymgr not configured")
	}
	body, _ := json.Marshal(in)
	resp, err := c.http().Post(c.BaseURL+"/paymgr/topup", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("paymgr: %s", string(b))
	}
	var res paymgr.RouteResult
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *PaymgrClient) http() *http.Client {
	if c.Client != nil {
		return c.Client
	}
	return http.DefaultClient
}

func routeToIntent(res *paymgr.RouteResult, adapter string) PaymentIntent {
	d := res.Decision
	return PaymentIntent{
		ChargeID:          res.TopupID,
		Provider:          adapter,
		Raw:               res.Message,
		USSDCode:          d.USSDCode,
		UserPhone:         d.UserPhone,
		RoutePath:         string(d.Path),
		NodeID:            d.NodeID,
		VMJobID:           d.VMJobID,
		CommandID:         d.CommandID,
		TransferType:      d.TransferType,
		RouteReason:       d.Reason,
		SystemDestination: d.SystemDestination,
	}
}
