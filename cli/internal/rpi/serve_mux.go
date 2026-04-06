package rpi

import (
	"net/http"
	"sync"
	"time"
)

// ServeMuxRoot provides thread-safe mutable root path for the HTTP mux.
// In orchestration mode the root switches from the original repo to the
// worktree once the engine has created it; handlers read the current value
// on every request via Get().
type ServeMuxRoot struct {
	mu   sync.RWMutex
	path string
}

// NewServeMuxRoot creates a ServeMuxRoot initialized with path.
func NewServeMuxRoot(path string) *ServeMuxRoot {
	return &ServeMuxRoot{path: path}
}

// Get returns the current root path.
func (r *ServeMuxRoot) Get() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.path
}

// Set updates the current root path.
func (r *ServeMuxRoot) Set(p string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.path = p
}

// NewDashboardServer creates an http.Server with standard timeouts for the RPI dashboard.
func NewDashboardServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    8192,
	}
}
