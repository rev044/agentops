# Go Kit

> Go development standards and tooling for AgentOps.

## Install

```bash
/plugin install go-kit@agentops
```

Requires: `solo-kit`

## What's Included

### Standards

Comprehensive Go coding standards in `skills/standards/references/go.md`:
- Effective Go patterns
- Error handling (no exceptions)
- Concurrency (goroutines, channels)
- Interface design
- Table-driven tests
- Common anti-patterns to avoid

### Hooks

| Hook | Trigger | What It Does |
|------|---------|--------------|
| `gofmt` | Edit *.go | Auto-format with gofmt |
| `golangci-lint` | Edit *.go | Lint with golangci-lint |
| `error-ignore-check` | Edit *.go | Warn about undocumented `_ =` |

### Patterns

**Error Handling**
```go
func process(data Data) (Result, error) {
    result, err := transform(data)
    if err != nil {
        return Result{}, fmt.Errorf("transform failed: %w", err)
    }
    return result, nil
}
```

**Table-Driven Tests**
```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 1, 2, 3},
        {"negative", -1, -2, -3},
        {"zero", 0, 0, 0},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := Add(tt.a, tt.b); got != tt.expected {
                t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.expected)
            }
        })
    }
}
```

**Interface Design**
```go
// Accept interfaces, return structs
type Reader interface {
    Read(p []byte) (n int, err error)
}

func NewService(r Reader) *Service {
    return &Service{reader: r}
}
```

## Requirements

- Go 1.21+
- Optional: golangci-lint (for hooks)

## License

MIT
