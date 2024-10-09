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

func NewState() State {
	return State{m: make(map[string]any), useMutex: true}
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
	request *http.Request

	// The raw response writer
	response *http.ResponseWriter

	// The name of the procedure
	procedureName string

	// The type of the procedure
	procedureType ProcedureType

	// User-defined state - this can be used to store any data that needs to be shared across procedures
	// For example, database connections, etc.
	//
	// NOTE: this is shared across all a functions in a single request
	State State
}

func NewContext(req *http.Request, res *http.ResponseWriter) *Context {
	return &Context{
		request:  req,
		response: res,
		State:    NewState(),
	}
}

// Request returns the underlying request
func (c *Context) Request() *http.Request {
	return c.request
}

// Response returns the underlying response writer
func (c *Context) Response() http.ResponseWriter {
	return *c.response
}

// Cookie returns the cookie with the specified key from the request and a boolean indicating whether the cookie exists
func (c *Context) Cookie(key string) (*http.Cookie, bool) {
	cookie, err := c.request.Cookie(key)
	if err != nil {
		return nil, false
	}
	return cookie, true
}

// SetCookie sets a cookie in the response
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(*c.response, cookie)
}

// ProcedureName returns the name of the procedure
func (c *Context) ProcedureName() string {
	return c.procedureName
}

// ProcedureType returns the type of the procedure
func (c *Context) ProcedureType() ProcedureType {
	return c.procedureType
}

// SetProcedureName sets the name of the procedure
func (c *Context) SetProcedureName(name string) {
	c.procedureName = name
}

// SetProcedureType sets the type of the procedure
func (c *Context) SetProcedureType(procedureType ProcedureType) {
	c.procedureType = procedureType
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
	return c.request.Header.Get(key)
}

// SetHeader sets the value of the specified header
func (c *Context) SetHeader(key, value string) {
	c.Response().Header().Set(key, value)
}

// Query returns the value of the specified query parameter
func (c *Context) Query(key string) string {
	return c.request.URL.Query().Get(key)
}

// GetBody returns the body of the request as a byte slice
func (c *Context) GetBody() []byte {
	body, _ := io.ReadAll(c.request.Body)
	return body
}
