package types

import (
	"io"
	"net/http"
	"sync"
)

// A container for user-defined state
type State struct {
	// The state map
	m map[string]any

	useMutex bool
	mu       sync.RWMutex
}

func NewState() *State {
	return &State{m: make(map[string]any), useMutex: true}
}

// Set sets a value in the state container
func (s *State) Set(key string, value any) {
	if s.useMutex {
		s.mu.Lock()
		defer s.mu.Unlock()
	}

	s.m[key] = value
}

// Get gets a value from the state container
func (s *State) Get(key string) any {
	if s.useMutex {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}

	return s.m[key]
}

// UseMutex sets whether to use the mutex lock on the state container
func (s *State) UseMutex(useMutex bool) {
	s.useMutex = useMutex
}

type Context struct {
	// The raw incoming request
	Request *http.Request

	// The raw response writer
	Response http.ResponseWriter

	// The name of the procedure
	ProcedureName string

	// The type of the procedure
	ProcedureType ProcedureType

	// User-defined state - this can be used to store any data that needs to be shared across procedures
	// For example, database connections, etc.
	//
	// NOTE: this is shared across all procedures and requests, the state container has a mutex lock by default to ensure thread safety, you can disable this by calling `DisableStateMutex`
	State State
}

// Set sets a value in the state container
func (c *Context) Set(key string, value any) {
	c.State.Set(key, value)
}

// Get gets a value from the state container
func (c *Context) Get(key string) any {
	return c.State.Get(key)
}

// EnableStateMutex enables the mutex lock on the state container
func (c *Context) EnableStateMutex() { c.State.UseMutex(true) }

// DisableStateMutex disables the mutex lock on the state container
func (c *Context) DisableStateMutex() { c.State.UseMutex(false) }

// Header returns the value of the specified header
func (c *Context) Header(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader sets the value of the specified header
func (c *Context) SetHeader(key, value string) {
	c.Response.Header().Set(key, value)
}

// Query returns the value of the specified query parameter
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// GetBody returns the body of the request as a byte slice
func (c *Context) GetBody() []byte {
	body, _ := io.ReadAll(c.Request.Body)
	return body
}
