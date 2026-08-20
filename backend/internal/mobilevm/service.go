package mobilevm

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gpuhub/cloud/internal/shared/rail"
)

var ErrUnknownType = errors.New("unknown transfer type")

// Service runs cross-rail transfer jobs inside mobileVM.
type Service struct {
	mu   sync.Mutex
	jobs map[string]*Job
}

func New() *Service {
	return &Service{jobs: make(map[string]*Job)}
}

// Start creates and runs a transfer job (simulated).
func (s *Service) Start(id string, req TransferRequest) (*Job, error) {
	if req.Type == "" {
		return nil, errors.New("transfer type required")
	}
	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	job := &Job{
		ID:        id,
		Type:      req.Type,
		Status:    "running",
		FromRef:   req.FromRef,
		ToRef:     req.ToRef,
		Carrier:   req.Carrier,
		Phone:     req.Phone,
		Amount:    req.Amount,
		CreatedAt: time.Now(),
	}

	switch req.Type {
	case TransferNumberToNumber:
		job.Message = fmt.Sprintf("Moved %d units from %s to %s via MM pool", req.Amount, req.FromRef, req.ToRef)
	case TransferFintechToBank:
		job.Message = fmt.Sprintf("Settled fintech %s → bank %s for %d units", req.FromRef, req.ToRef, req.Amount)
	case TransferBankToMobileMoney:
		ussd, err := rail.BuildUSSD(req.Carrier, req.Phone, req.Amount)
		if err != nil {
			return nil, err
		}
		job.USSD = ussd
		job.Message = fmt.Sprintf("Bank→MM dial-up queued: %s", ussd)
	case TransferCarrierDialup:
		ussd, err := rail.BuildUSSD(req.Carrier, req.Phone, req.Amount)
		if err != nil {
			return nil, err
		}
		job.USSD = ussd
		job.Message = fmt.Sprintf("Special carrier dial-up: %s", ussd)
	default:
		return nil, ErrUnknownType
	}

	job.Status = "completed"
	job.FinishedAt = time.Now()

	s.mu.Lock()
	s.jobs[id] = job
	s.mu.Unlock()
	return job, nil
}

func (s *Service) Get(id string) (*Job, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok {
		return nil, false
	}
	cp := *j
	return &cp, true
}

func (s *Service) List() []Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		out = append(out, *j)
	}
	return out
}
