# Go Standards Catalog - Vibe Canonical Reference

**Version:** 1.0.0
**Last Updated:** 2026-01-21
**Purpose:** Canonical Go standards for vibe skill validation

---

## Table of Contents

1. [Error Handling Patterns](#error-handling-patterns)
2. [Interface Design](#interface-design)
3. [Concurrency Patterns](#concurrency-patterns)
4. [Security Practices](#security-practices)
5. [Package Organization](#package-organization)
6. [Testing Patterns](#testing-patterns)
7. [Anti-Patterns Avoided](#anti-patterns-avoided)
8. [Compliance Assessment](#compliance-assessment)

---

## Error Handling Patterns

### Custom Error Types

```go
type AppError struct {
    Code     string
    Message  string
    Cause    error
    Metadata map[string]any
}

func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Cause
}
```

### Error Wrapping with %w

```go
// CORRECT
resp, err := client.Do(req)
if err != nil {
    return nil, fmt.Errorf("sending request: %w", err)
}

// INCORRECT - Breaks error chains
if err != nil {
    return nil, fmt.Errorf("sending request: %v", err)
}
```

### Intentional Error Ignores

```go
// CORRECT - Documented
defer func() {
    _ = conn.Close() // nolint:errcheck - best effort cleanup
}()

// INCORRECT - Silent ignore
defer func() {
    _ = conn.Close()
}()
```

---

## Interface Design

### Accept Interfaces, Return Structs

```go
// Define interface
type Agent interface {
    Initialize(ctx context.Context) error
    Invoke(ctx context.Context, req *Request) (*Response, error)
}

// Functions accept interface (flexible)
func ProcessAgent(ctx context.Context, agent Agent) error {
    if err := agent.Initialize(ctx); err != nil {
        return fmt.Errorf("initialization failed: %w", err)
    }
    return nil
}

// Constructors return struct (concrete)
func NewRegistry() *Registry {
    return &Registry{
        agents: make(map[string]Agent),
        mu:     sync.RWMutex{},
    }
}
```

### Small, Focused Interfaces

```go
// Good - Composable
type Initializer interface {
    Initialize(ctx context.Context) error
}

type Invoker interface {
    Invoke(ctx context.Context, req *Request) (*Response, error)
}

type Agent interface {
    Initializer
    Invoker
}
```

---

## Concurrency Patterns

### Context Propagation (Required)

```go
// HTTP Requests
req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)

// Database Operations
rows, err := db.QueryContext(ctx, query)

// Custom Functions
func (c *Client) Invoke(ctx context.Context, req *Request) (*Response, error)
```

### Thread-Safe Data Structures

```go
type Registry struct {
    items map[string]Item
    mu    sync.RWMutex
}

func (r *Registry) Get(key string) (Item, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    // ...
}

func (r *Registry) Set(key string, item Item) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    // ...
}
```

---

## Security Practices

### Constant-Time Comparison

```go
import "crypto/subtle"

// CORRECT - Timing attack resistant
if subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
    return ErrUnauthorized
}

// INCORRECT - Vulnerable
if token == expectedToken { ... }
```

### TLS Configuration

```go
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS13,
}
```

---

## Package Organization

### Layered Architecture

```
project/
├── cmd/                    # Binaries
│   ├── server/
│   └── cli/
├── internal/              # Private packages
│   ├── domain/
│   ├── handlers/
│   └── repository/
├── pkg/                   # Public packages
│   ├── api/
│   └── client/
└── tests/
    ├── e2e/
    └── integration/
```

### Import Grouping

```go
import (
    // Standard library
    "context"
    "fmt"

    // External dependencies
    "github.com/external/package"

    // Internal packages
    "myproject.com/internal/domain"
)
```

---

## Testing Patterns

### Table-Driven Tests

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid", "user@example.com", false},
        {"missing @", "userexample.com", true},
        {"empty", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail(%q) error = %v, wantErr %v",
                    tt.email, err, tt.wantErr)
            }
        })
    }
}
```

### Test Helpers with t.Helper()

```go
func setupTestServer(t *testing.T) *httptest.Server {
    t.Helper()
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Mock responses
    }))
    t.Cleanup(func() { server.Close() })
    return server
}
```

---

## Anti-Patterns Avoided

### No Naked Returns
```go
// BAD
func bad() (err error) {
    err = doSomething()
    return
}

// GOOD
func good() error {
    err := doSomething()
    return err
}
```

### No Pointer to Interface
```go
// BAD
func bad(agent *Agent) // Interface is already a reference

// GOOD
func good(agent Agent)
```

### No Goroutine Leaks
```go
// BAD
go func() {
    for { work() }
}()

// GOOD
go func() {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            work()
        }
    }
}()
```

---

## Compliance Assessment

**Use letter grades + evidence, NOT numeric scores.**

### Grading Scale

| Grade | Finding Threshold | Description |
|-------|------------------|-------------|
| A+ | 0-2 minor findings | Exemplary |
| A | <5 HIGH findings | Excellent |
| A- | 5-15 HIGH findings | Very Good |
| B+ | 15-25 HIGH findings | Good |
| B | 25-40 HIGH findings | Satisfactory |
| C | 60+ HIGH findings | Significant Issues |
| D | 1+ CRITICAL findings | Not production-ready |
| F | Multiple CRITICAL | Critical Issues |

---

## Additional Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)
- [golangci-lint Linters](https://golangci-lint.run/usage/linters/)
