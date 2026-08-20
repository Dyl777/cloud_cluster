package paymgr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gpuhub/cloud/internal/mobilevm"
	"github.com/gpuhub/cloud/internal/shared/id"
	"github.com/gpuhub/cloud/internal/shared/rail"
)

// GatewayClient talks to mobilegateway.
type GatewayClient struct {
	BaseURL string
	Client  *http.Client
}

// VMClient talks to mobilevm.
type VMClient struct {
	BaseURL string
	Client  *http.Client
}

type liveNode struct {
	ID          string `json:"id"`
	PendingJobs int    `json:"pending_jobs"`
	LatencyMs   int64  `json:"latency_ms"`
	Connected   bool   `json:"connected"`
	SIMs        []struct {
		Slot    int    `json:"slot"`
		Number  string `json:"number"`
		Carrier string `json:"carrier"`
	} `json:"sims"`
	CanRefund bool `json:"can_refund"`
}

func (g *GatewayClient) LiveNodes() ([]liveNode, error) {
	if g.BaseURL == "" {
		return nil, nil
	}
	resp, err := g.http().Get(g.BaseURL + "/gateway/nodes/live")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var nodes []liveNode
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return nil, err
	}
	return nodes, nil
}

func (g *GatewayClient) DispatchUSSD(topupID, carrier, userPhone, collectionNumber string, amount int64) (ussd, nodeID string, slot int, cmdID string, err error) {
	cmdID = id.New("cmd")
	ussd, err = rail.BuildUSSD(carrier, collectionNumber, amount)
	if err != nil {
		return "", "", 0, "", err
	}
	body, _ := json.Marshal(map[string]any{
		"carrier": carrier,
		"phone":   userPhone,
		"command": map[string]any{
			"id":       cmdID,
			"kind":     "ussd_dial",
			"ussd":     ussd,
			"carrier":  carrier,
			"phone":    userPhone,
			"amount":   amount,
			"topup_id": topupID,
		},
	})
	resp, err := g.http().Post(g.BaseURL+"/gateway/dispatch", "application/json", bytes.NewReader(body))
	if err != nil {
		return ussd, "", 0, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return ussd, "", 0, "", fmt.Errorf("gateway: %s", string(b))
	}
	var out struct {
		CommandID string `json:"command_id"`
		NodeID    string `json:"node_id"`
		SIMSlot   int    `json:"sim_slot"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", "", 0, "", err
	}
	return ussd, out.NodeID, out.SIMSlot, out.CommandID, nil
}

func (g *GatewayClient) Refund(nodeID, topupID, phone string) error {
	body, _ := json.Marshal(map[string]any{
		"node_id": nodeID,
		"phone":   phone,
		"command": map[string]any{
			"id":       id.New("cmd"),
			"kind":     "refund",
			"topup_id": topupID,
			"phone":    phone,
		},
	})
	resp, err := g.http().Post(g.BaseURL+"/gateway/dispatch", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (g *GatewayClient) http() *http.Client {
	if g.Client != nil {
		return g.Client
	}
	return http.DefaultClient
}

func (v *VMClient) StartTransfer(req mobilevm.TransferRequest) (*mobilevm.Job, error) {
	body, _ := json.Marshal(req)
	resp, err := v.http().Post(v.BaseURL+"/mobilevm/transfer", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mobilevm: %s", string(b))
	}
	var job mobilevm.Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (v *VMClient) http() *http.Client {
	if v.Client != nil {
		return v.Client
	}
	return http.DefaultClient
}
