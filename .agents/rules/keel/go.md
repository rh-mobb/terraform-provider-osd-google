---
description: "Go coding conventions and best practices"
globs: ["**/*.go", "**/go.mod", "**/go.sum"]
alwaysApply: false
---

# Go Standards

Standards for Go development.

## Tooling

- Format with `gofmt` (or `goimports`) — no exceptions
- Lint with `golangci-lint` using a project-level `.golangci.yml` config
- Vet with `go vet ./...` for correctness issues the compiler won't catch
- Run tests with `go test ./...`

## Style

- Follow `gofmt` and `goimports` formatting — no exceptions
- Use `golangci-lint` with a project-level `.golangci.yml` config
- Keep exported names clear and unexported names short
- Use `MixedCaps` (not underscores) for multi-word names
- Package names are lowercase, single-word, and descriptive (`http`, `json`, `auth`)

## Package Design

- Design packages around what they provide, not what they contain
- Avoid package names like `util`, `common`, `helpers` — be specific
- Keep package APIs small — export only what consumers need
- Avoid circular imports — if two packages depend on each other, extract shared types into a third

## Error Handling

- Return errors as the last return value
- Handle every error — never use `_` to discard errors silently
- Wrap errors with context using `fmt.Errorf("operation failed: %w", err)`
- Use sentinel errors (`var ErrNotFound = errors.New(...)`) for expected conditions
- Use `errors.Is()` and `errors.As()` for error checking — never compare error strings

```go
func GetUser(ctx context.Context, id string) (*User, error) {
    user, err := db.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get user %s: %w", id, err)
    }
    if user == nil {
        return nil, ErrNotFound
    }
    return user, nil
}
```

## Concurrency

- Use goroutines and channels for concurrent work — avoid shared mutable state
- Always use `context.Context` as the first parameter for cancellable operations
- Use `sync.WaitGroup` to coordinate goroutine completion
- Protect shared state with `sync.Mutex` when channels aren't appropriate
- Use `errgroup.Group` for concurrent tasks that can fail
- Never start goroutines without a clear shutdown path

## Interfaces

- Define interfaces where they are used, not where they are implemented
- Keep interfaces small — one to three methods
- Use the `-er` suffix for single-method interfaces (`Reader`, `Writer`, `Stringer`)
- Accept interfaces, return concrete types
- Don't use interfaces for mocking alone — only when you need polymorphism

## Structs

- Use struct literals with field names — never positional initialization
- Use pointer receivers when the method modifies state or the struct is large
- Use value receivers for small, immutable structs
- Group struct fields logically; place exported fields first

## Testing

- Use table-driven tests for testing multiple scenarios
- Use `testify/assert` or `testify/require` for assertions
- Use `httptest` for HTTP handler testing
- Name test functions: `TestFunctionName_Scenario`
- Use `t.Helper()` in test helper functions
- Use `t.Parallel()` for tests that can run concurrently

```go
func TestGetUser_NotFound(t *testing.T) {
    t.Parallel()

    svc := NewService(mockDB)
    user, err := svc.GetUser(context.Background(), "nonexistent")

    require.ErrorIs(t, err, ErrNotFound)
    assert.Nil(t, user)
}
```

## Project Layout

- Follow the standard Go project layout:
  - `cmd/` — main applications
  - `internal/` — private application code
  - `pkg/` — public library code (use sparingly)
  - `api/` — API definitions (protobuf, OpenAPI)
- Use `internal/` to prevent external packages from importing implementation details
- Keep `main.go` thin — parse config, wire dependencies, start the server

## Agent Behavior

- After modifying `.go` files, run `gofmt`, `go vet ./...`, and `golangci-lint run` to verify correctness
- Fix all lint and vet findings before presenting changes
- Run `go build ./...` to confirm the project compiles after changes

## .gitignore

Ensure these Go-specific patterns are in the project's `.gitignore`:

```gitignore
/bin/
/vendor/
*.test
coverage.out
```
