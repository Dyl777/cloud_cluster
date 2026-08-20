package mobilegateway

import (
	"net/http"
	"strings"

	"github.com/gpuhub/cloud/internal/shared/httpx"
	"github.com/gpuhub/cloud/internal/shared/id"
)

// Handler wires HTTP routes to the gateway Service.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /gateway/nodes", h.listNodes)
	mux.HandleFunc("GET /gateway/nodes/live", h.liveNodes)
	mux.HandleFunc("POST /gateway/nodes/register", h.register)
	mux.HandleFunc("POST /gateway/nodes/{nodeID}/heartbeat", h.heartbeat)
	mux.HandleFunc("GET /gateway/nodes/{nodeID}/poll", h.poll)
	mux.HandleFunc("POST /gateway/nodes/{nodeID}/result", h.reportResult)
	mux.HandleFunc("POST /gateway/dispatch", h.dispatch)
	mux.HandleFunc("GET /gateway/commands/{cmdID}", h.getResult)
}

func (h *Handler) token(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	return strings.TrimPrefix(auth, "Bearer ")
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Authorize(h.token(r)); err != nil {
		httpx.WriteErr(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}
	var n Node
	if err := httpx.Decode(r, &n); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if n.ID == "" {
		n.ID = id.New("node")
	}
	httpx.WriteJSON(w, http.StatusOK, h.svc.RegisterNode(n))
}

func (h *Handler) heartbeat(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Authorize(h.token(r)); err != nil {
		httpx.WriteErr(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}
	nodeID := r.PathValue("nodeID")
	var body struct {
		Pending   int   `json:"pending_jobs"`
		LatencyMs int64 `json:"latency_ms"`
	}
	_ = httpx.Decode(r, &body)
	h.svc.Heartbeat(nodeID, body.Pending, body.LatencyMs)
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) poll(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Authorize(h.token(r)); err != nil {
		httpx.WriteErr(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}
	nodeID := r.PathValue("nodeID")
	cmd, ok := h.svc.Poll(nodeID)
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, cmd)
}

func (h *Handler) reportResult(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Authorize(h.token(r)); err != nil {
		httpx.WriteErr(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}
	nodeID := r.PathValue("nodeID")
	var res CommandResult
	if err := httpx.Decode(r, &res); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	h.svc.ReportResult(nodeID, res)
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "recorded"})
}

func (h *Handler) listNodes(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, h.svc.AllNodes())
}

func (h *Handler) liveNodes(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, h.svc.LiveNodes())
}

func (h *Handler) dispatch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NodeID  string `json:"node_id,omitempty"`
		Carrier string `json:"carrier"`
		Phone   string `json:"phone"`
		Command Command
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if body.Command.ID == "" {
		body.Command.ID = id.New("cmd")
	}
	var nodeID string
	var slot int
	var err error
	if body.NodeID != "" {
		body.Command.SIMSlot = 0
		err = h.svc.Dispatch(body.NodeID, body.Command)
		nodeID = body.NodeID
	} else {
		nodeID, slot, err = h.svc.DispatchBest(body.Carrier, body.Phone, body.Command)
	}
	if err != nil {
		httpx.WriteErr(w, http.StatusServiceUnavailable, "dispatch_failed", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusAccepted, map[string]any{
		"command_id": body.Command.ID,
		"node_id":    nodeID,
		"sim_slot":   slot,
	})
}

func (h *Handler) getResult(w http.ResponseWriter, r *http.Request) {
	cmdID := r.PathValue("cmdID")
	res, ok := h.svc.GetResult(cmdID)
	if !ok {
		httpx.WriteErr(w, http.StatusNotFound, "not_found", "result not ready")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}
