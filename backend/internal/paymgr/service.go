package paymgr

import (
	"errors"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gpuhub/cloud/internal/mobilevm"
	"github.com/gpuhub/cloud/internal/shared/payconfig"
	"github.com/gpuhub/cloud/internal/shared/rail"
)

var ErrNoTopup = errors.New("topup not found")

// Service is the backend payment manager — decides where top-ups move.
type Service struct {
	mu      sync.Mutex
	topups  map[string]*TopupInput
	sys     payconfig.SystemPaymentConfig
	gateway *GatewayClient
	vm      *VMClient
}

func New(gatewayURL, vmURL string) *Service {
	return &Service{
		topups: make(map[string]*TopupInput),
		sys:    payconfig.DefaultSystemConfig(),
		gateway: &GatewayClient{
			BaseURL: gatewayURL,
			Client:  &http.Client{Timeout: 8 * time.Second},
		},
		vm: &VMClient{
			BaseURL: vmURL,
			Client:  &http.Client{Timeout: 8 * time.Second},
		},
	}
}

// SystemConfig returns the platform collection configuration (system side).
func (s *Service) SystemConfig() payconfig.SystemPaymentConfig {
	return s.sys
}

// LiveNodes scans mobilegateway for connected nodes (refund / load-balance view).
func (s *Service) LiveNodes() ([]liveNode, error) {
	return s.gateway.LiveNodes()
}

func (s *Service) resolveDestination(in TopupInput) (payconfig.Destination, error) {
	method := in.Method
	if in.TransferHint == "bank_to_mobilemoney" {
		method = "mobile_money"
	}
	return s.sys.ResolveDestination(method, in.Carrier, in.RailProvider)
}

// Route decides the execution path without side effects.
func (s *Service) Route(in TopupInput) RouteDecision {
	units := amountUnits(in.Subunits)
	dest, _ := s.resolveDestination(in)

	d := RouteDecision{
		UserPhone:         in.Phone,
		SystemDestination: dest,
	}

	switch in.Method {
	case "bank":
		if in.TransferHint == "bank_to_mobilemoney" || in.RailProvider == "orange_bank" {
			d.Target = "mobile_money"
			d.Path = PathMobileVM
			d.TransferType = string(mobilevm.TransferBankToMobileMoney)
			d.Reason = "bank→mobilemoney cross-rail via mobileVM"
			return d
		}
		d.Target = "bank"
		d.Path = PathBank
		d.Reason = "direct bank rail → " + dest.Label
		return d

	case "fintech":
		if in.TransferHint == "fintech_to_bank" {
			d.Target = "bank"
			d.Path = PathMobileVM
			d.TransferType = string(mobilevm.TransferFintechToBank)
			d.Reason = "fintech→bank settlement via mobileVM"
			return d
		}
		d.Target = "fintech"
		d.Path = PathFintech
		d.Reason = "direct fintech rail → " + dest.Label
		return d

	case "mobile_money":
		collection := dest.Number
		if collection == "" {
			collection = in.Phone
		}

		if in.TransferHint == "number_to_number" {
			d.Target = "mobile_money"
			d.Path = PathMobileVM
			d.TransferType = string(mobilevm.TransferNumberToNumber)
			d.Reason = "number→number pool move via mobileVM"
			return d
		}
		if isSpecialCarrier(in.Carrier) {
			d.Target = "mobile_money"
			d.Path = PathMobileVM
			d.TransferType = string(mobilevm.TransferCarrierDialup)
			d.Reason = "special carrier dial-up via mobileVM"
			return d
		}

		nodes, _ := s.gateway.LiveNodes()
		if node := pickBestNode(nodes, in.Carrier, in.Phone); node != nil {
			ussd, _ := rail.BuildUSSD(in.Carrier, collection, units)
			d.Target = "mobile_money"
			d.Path = PathMobileGateway
			d.NodeID = node.ID
			d.USSDCode = ussd
			d.Reason = "live gateway node — load-balanced → " + dest.Label
			return d
		}

		ussd, _ := rail.BuildUSSD(in.Carrier, collection, units)
		d.Target = "mobile_money"
		d.Path = PathDirect
		d.USSDCode = ussd
		d.Reason = "user SIM USSD → platform " + dest.Number
		return d
	}

	d.Target = in.Method
	d.Path = PathDirect
	d.Reason = "fallback"
	return d
}

func pickBestNode(nodes []liveNode, carrier, userPhone string) *liveNode {
	var cands []liveNode
	for _, n := range nodes {
		if !n.Connected {
			continue
		}
		for _, sim := range n.SIMs {
			if userPhone != "" && rail.PhoneSuffixMatch(sim.Number, userPhone) {
				cands = append(cands, n)
				break
			}
			if carrier != "" && sim.Carrier == carrier {
				cands = append(cands, n)
				break
			}
		}
	}
	if len(cands) == 0 {
		return nil
	}
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].PendingJobs != cands[j].PendingJobs {
			return cands[i].PendingJobs < cands[j].PendingJobs
		}
		return cands[i].LatencyMs < cands[j].LatencyMs
	})
	return &cands[0]
}

// StartTopup routes and kicks off execution.
func (s *Service) StartTopup(in TopupInput) (*RouteResult, error) {
	dest, err := s.resolveDestination(in)
	if err != nil {
		return nil, err
	}

	decision := s.Route(in)
	decision.SystemDestination = dest
	decision.UserPhone = in.Phone

	units := amountUnits(in.Subunits)
	collection := dest.Number
	if collection == "" {
		collection = in.Phone
	}

	var message string

	switch decision.Path {
	case PathMobileGateway:
		ussd, nodeID, slot, cmdID, err := s.gateway.DispatchUSSD(
			in.ID, in.Carrier, in.Phone, collection, units,
		)
		if err != nil {
			decision.Path = PathDirect
			decision.Reason = "gateway dispatch failed — fallback to direct USSD"
			decision.USSDCode = ussd
		} else {
			decision.USSDCode = ussd
			decision.NodeID = nodeID
			decision.SIMSlot = slot
			decision.CommandID = cmdID
			message = "dispatched to gateway node " + nodeID + " → " + dest.Label
		}

	case PathMobileVM:
		job, err := s.vm.StartTransfer(mobilevm.TransferRequest{
			Type:    mobilevm.TransferType(decision.TransferType),
			FromRef: in.AccountRef,
			ToRef:   collection,
			Carrier: in.Carrier,
			Phone:   in.Phone,
			Amount:  units,
			TopupID: in.ID,
		})
		if err != nil {
			return nil, err
		}
		decision.VMJobID = job.ID
		decision.USSDCode = job.USSD
		message = job.Message + " → " + dest.Label

	case PathDirect:
		if decision.USSDCode == "" {
			ussd, err := rail.BuildUSSD(in.Carrier, collection, units)
			if err != nil {
				return nil, err
			}
			decision.USSDCode = ussd
		}
		message = "pay from " + in.Phone + " via USSD → platform " + dest.Number

	default:
		message = "routed to " + string(decision.Path) + " → " + dest.Label
	}

	s.mu.Lock()
	cp := in
	s.topups[in.ID] = &cp
	s.mu.Unlock()

	return &RouteResult{
		TopupID:  in.ID,
		Decision: decision,
		Message:  message,
		Status:   "pending",
	}, nil
}

// Refund routes a refund to a gateway node that supports it.
func (s *Service) Refund(topupID string) error {
	s.mu.Lock()
	in, ok := s.topups[topupID]
	s.mu.Unlock()
	if !ok {
		return ErrNoTopup
	}
	nodes, err := s.gateway.LiveNodes()
	if err != nil {
		return err
	}
	for _, n := range nodes {
		if n.CanRefund {
			return s.gateway.Refund(n.ID, topupID, in.Phone)
		}
	}
	return errors.New("no refund-capable node online")
}

func (s *Service) Get(topupID string) (*TopupInput, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.topups[topupID]
	return t, ok
}
