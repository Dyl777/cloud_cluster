package payments

import "errors"

// registry maps a method to its provider.
type registry struct {
	providers map[Method]Provider
}

func newRegistry() *registry {
	return &registry{providers: make(map[Method]Provider)}
}

func (r *registry) add(m Method, p Provider) { r.providers[m] = p }

var ErrMethod = errors.New("unsupported payment method")

func (r *registry) get(m Method) (Provider, error) {
	p, ok := r.providers[m]
	if !ok {
		return nil, ErrMethod
	}
	return p, nil
}
