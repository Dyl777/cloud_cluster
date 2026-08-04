package httpx

import (
	"net/http"
)

// Server wraps an *http.ServeMux and exposes ListenAndServe. Use NewServer,
// then Method-specific handlers, then Run.
type Server struct {
	mux  *http.ServeMux
	addr string
}

// NewServer creates a server on the given address.
func NewServer(addr string) *Server {
	return &Server{mux: http.NewServeMux(), addr: addr}
}

// Mux returns the underlying handler for registration.
func (s *Server) Mux() *http.ServeMux { return s.mux }

// Run starts serving until the process is interrupted.
func (s *Server) Run() error {
	return http.ListenAndServe(s.addr, s.mux)
}
