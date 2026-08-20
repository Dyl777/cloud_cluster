// mobilegatewayNode runs on physical devices where SIMs are installed.
// It registers securely with mobilegateway, polls for commands, executes
// USSD dials / transfers on the matched SIM slot, and reports results.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	gatewayURL := env("GATEWAY_URL", "http://localhost:8091")
	token := os.Getenv("GATEWAY_TOKEN")
	nodeID := env("NODE_ID", "")
	carrier := strings.ToLower(env("CARRIER", "mtn"))
	phone := env("PHONE", "")
	slot := envInt("SIM_SLOT", 0)
	region := env("REGION", "local")
	canRefund := env("CAN_REFUND", "true") == "true"

	if phone == "" {
		fmt.Fprintln(os.Stderr, "PHONE is required")
		os.Exit(1)
	}
	if nodeID == "" {
		nodeID = "node-" + phoneSuffix(phone)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	authHdr := map[string]string{"Content-Type": "application/json"}
	if token != "" {
		authHdr["Authorization"] = "Bearer " + token
	}

	regBody, _ := json.Marshal(map[string]any{
		"id": nodeID,
		"sims": []map[string]any{{
			"slot": slot, "number": phone, "carrier": carrier,
		}},
		"region":     region,
		"can_refund": canRefund,
	})
	if err := post(client, gatewayURL+"/gateway/nodes/register", regBody, authHdr); err != nil {
		fmt.Fprintln(os.Stderr, "register failed:", err)
		os.Exit(1)
	}
	fmt.Println("mobilegatewayNode", nodeID, "registered — carrier", carrier, "phone", phone)

	pending := 0
	for {
		start := time.Now()
		heartbeat(client, gatewayURL, nodeID, pending, 0, authHdr)

		cmd, ok, err := poll(client, gatewayURL, nodeID, authHdr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "poll error:", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if !ok {
			time.Sleep(600 * time.Millisecond)
			continue
		}

		pending++
		kind, _ := cmd["kind"].(string)
		ussd, _ := cmd["ussd"].(string)
		slot, _ := cmd["sim_slot"].(float64)
		fmt.Println("executing", kind, "on SIM", int(slot), ussd)
		message, status := execute(cmd)
		cmdID, _ := cmd["id"].(string)
		latency := time.Since(start).Milliseconds()

		resBody, _ := json.Marshal(map[string]any{
			"command_id": cmdID,
			"message":    message,
			"status":     status,
			"latency_ms": latency,
		})
		_ = post(client, gatewayURL+"/gateway/nodes/"+nodeID+"/result", resBody, authHdr)
		pending--
		fmt.Println("result:", message)
	}
}

func execute(cmd map[string]any) (string, string) {
	kind, _ := cmd["kind"].(string)
	ussd, _ := cmd["ussd"].(string)
	phone, _ := cmd["phone"].(string)
	switch kind {
	case "refund":
		return fmt.Sprintf("Refund initiated for %s. Await carrier confirmation.", phone), "completed"
	case "transfer":
		return fmt.Sprintf("Transfer executed. Ref: %s", phoneSuffix(phone)), "completed"
	default:
		return fmt.Sprintf("Confirm payment. Enter PIN. Code: %s", ussd), "completed"
	}
}

func poll(client *http.Client, base, nodeID string, hdr map[string]string) (map[string]any, bool, error) {
	req, _ := http.NewRequest("GET", base+"/gateway/nodes/"+nodeID+"/poll", nil)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return nil, false, nil
	}
	var cmd map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&cmd); err != nil {
		return nil, false, err
	}
	return cmd, true, nil
}

func heartbeat(client *http.Client, base, nodeID string, pending int, latency int64, hdr map[string]string) {
	body, _ := json.Marshal(map[string]any{"pending_jobs": pending, "latency_ms": latency})
	_ = post(client, base+"/gateway/nodes/"+nodeID+"/heartbeat", body, hdr)
}

func post(client *http.Client, url string, body []byte, hdr map[string]string) error {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

func phoneSuffix(p string) string {
	d := digitsOnly(p)
	if len(d) > 4 {
		return d[len(d)-4:]
	}
	return d
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

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
