# Testing Requirements

## New features require tests

All new features need tests before or alongside implementation:
- Business logic: table-driven unit tests covering happy path + edge cases
- HTTP/gRPC handlers: integration tests with real DB, not mocks
- Auth paths: explicit tests for both success and failure cases
- Money/financial calculations: exhaustive edge case coverage

## Go test patterns

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    WantType
        wantErr bool
    }{
        {name: "happy path", ...},
        {name: "edge case", ...},
        {name: "error case", wantErr: true, ...},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

## Integration tests

- Use real database — no mocked repositories for integration tests
- Use the project's `testutil` package patterns
- Clean up test state after each test

## Coverage targets

- Business logic packages: 80%+
- Auth packages: 90%+
- Infrastructure adapters: prefer integration tests over unit tests
