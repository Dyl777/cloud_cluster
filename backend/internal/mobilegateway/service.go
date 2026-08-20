package mobilegateway

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gpuhub/cloud/internal/shared/rail"
)

var (
	ErrNoNode       = errors.New("no live gateway node available")
	ErrNodeOffline  = errors.New("gateway node offline")
	ErrCmdNotFound  = errors.New("command not found")
	ErrUnauthorized = errors.New("unauthorized")
)

// Service tracks physical gateway nodes and dispatches work.
type Service struct {
	mu       sync.Mutex
	token    string
	nodes    map[string]*Node
	pending  map[string][]Command
	results  map[string]CommandResult
	latency  map[string][]int64 // rolling latency samples per node
}

// New returns a gateway service. token secures node registration.
func New(token string) *Service {
	return &Service{
		token:   token,
		nodes:   make(map[string]*Node),
		pending: make(map[string][]Command),
		results: make(map[string]CommandResult),
		latency: make(map[string][]int64),
	}
}

func (s *Service) Authorize(provided string) error {
	if s.token == "" {
		return nil
	}
	if provided != s.token {
		return ErrUnauthorized
	}
	return nil
}

// RegisterNode upserts a node and its SIM inventory.
func (s *Service) RegisterNode(n Node) Node {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.nodes[n.ID]
	if !ok {
		existing = &Node{ID: n.ID}
		s.nodes[n.ID] = existing
	}
	existing.SIMs = n.SIMs
	existing.Region = n.Region
	existing.CanRefund = n.CanRefund
	existing.Connected = true
	existing.LastSeen = time.Now()
	return *existing
}

// Heartbeat updates node liveness and reported metrics.
func (s *Service) Heartbeat(id string, pending int, latencyMs int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.nodes[id]
	if !ok {
		return
	}
	n.LastSeen = time.Now()
	n.Connected = true
	n.PendingJobs = pending
	if latencyMs > 0 {
		n.LatencyMs = latencyMs
		s.latency[id] = append(s.latency[id], latencyMs)
		if len(s.latency[id]) > 20 {
			s.latency[id] = s.latency[id][len(s.latency[id])-20:]
		}
	}
}

// LiveNodes returns connected nodes seen within the last 30s.
func (s *Service) LiveNodes() []Node {
	s.mu.Lock()
	defer s.mu.Unlock()
	cutoff := time.Now().Add(-30 * time.Second)
	out := make([]Node, 0)
	for _, n := range s.nodes {
		if n.Connected && n.LastSeen.After(cutoff) {
			out = append(out, *n)
		}
	}
	return out
}

// AllNodes returns every registered node (for admin scan).
func (s *Service) AllNodes() []Node {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Node, 0, len(s.nodes))
	for _, n := range s.nodes {
		out = append(out, *n)
	}
	return out
}

// PickNode selects the best live node for carrier/phone (load + latency).
func (s *Service) PickNode(carrier, phone string) (*Node, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cutoff := time.Now().Add(-30 * time.Second)
	type candidate struct {
		node *Node
		slot int
	}
	var cands []candidate

	for _, n := range s.nodes {
		if !n.Connected || n.LastSeen.Before(cutoff) {
			continue
		}
		slot := matchSIM(n.SIMs, carrier, phone)
		if slot >= 0 {
			cands = append(cands, candidate{n, slot})
		}
	}
	if len(cands) == 0 {
		return nil, 0, ErrNoNode
	}

	sort.Slice(cands, func(i, j int) bool {
		a, b := cands[i].node, cands[j].node
		if a.PendingJobs != b.PendingJobs {
			return a.PendingJobs < b.PendingJobs
		}
		return a.LatencyMs < b.LatencyMs
	})

	chosen := cands[0]
	chosen.node.PendingJobs++
	return chosen.node, chosen.slot, nil
}

func matchSIM(sims []SIMSlot, carrier, phone string) int {
	carrier = strings.ToLower(strings.TrimSpace(carrier))
	for _, sim := range sims {
		if phone != "" && rail.PhoneSuffixMatch(sim.Number, phone) {
			return sim.Slot
		}
	}
	for _, sim := range sims {
		if carrier != "" && strings.ToLower(sim.Carrier) == carrier {
			return sim.Slot
		}
	}
	if len(sims) == 1 {
		return sims[0].Slot
	}
	return -1
}

// Dispatch queues a command on the chosen node.
func (s *Service) Dispatch(nodeID string, cmd Command) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.nodes[nodeID]
	if !ok || !n.Connected {
		return ErrNodeOffline
	}
	s.pending[nodeID] = append(s.pending[nodeID], cmd)
	return nil
}

// DispatchBest picks a node and queues the command.
func (s *Service) DispatchBest(carrier, phone string, cmd Command) (string, int, error) {
	node, slot, err := s.PickNode(carrier, phone)
	if err != nil {
		return "", 0, err
	}
	cmd.SIMSlot = slot
	if err := s.Dispatch(node.ID, cmd); err != nil {
		return "", 0, err
	}
	return node.ID, slot, nil
}

// Poll returns the next command for a node.
func (s *Service) Poll(nodeID string) (Command, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	q := s.pending[nodeID]
	if len(q) == 0 {
		return Command{}, false
	}
	cmd := q[0]
	s.pending[nodeID] = q[1:]
	return cmd, true
}

// ReportResult stores the outcome and decrements pending count.
func (s *Service) ReportResult(nodeID string, res CommandResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results[res.CommandID] = res
	if n, ok := s.nodes[nodeID]; ok {
		if n.PendingJobs > 0 {
			n.PendingJobs--
		}
		if res.LatencyMs > 0 {
			n.LatencyMs = res.LatencyMs
		}
		n.LastSeen = time.Now()
	}
}

// GetResult returns a stored command result.
func (s *Service) GetResult(cmdID string) (CommandResult, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.results[cmdID]
	return r, ok
}

// BuildAndDispatchUSSD constructs USSD and sends to best node.
func (s *Service) BuildAndDispatchUSSD(cmdID, topupID, carrier, phone string, amount int64) (string, string, int, error) {
	ussd, err := rail.BuildUSSD(carrier, phone, amount)
	if err != nil {
		return "", "", 0, err
	}
	cmd := Command{
		ID:      cmdID,
		Kind:    "ussd_dial",
		USSD:    ussd,
		Carrier: carrier,
		Phone:   phone,
		Amount:  amount,
		TopupID: topupID,
	}
	nodeID, slot, err := s.DispatchBest(carrier, phone, cmd)
	if err != nil {
		return ussd, "", 0, err
	}
	return ussd, nodeID, slot, nil
}
