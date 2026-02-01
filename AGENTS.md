TUI Screen Capture CLI

## Build & Test

```
golangci-lint run ./... --fix
go test -v -race -timeout=30s ./...
go build -o tui-capture ./cmd
```

Before committing:

```
bunx prettier --write "**/*.md"
```

## TDD Methodology

Write a failing test first, implement minimal code to pass, refactor. Never write implementation before the test exists.

## Coding Conventions

### Error Handling

Always wrap errors with context:

```go
return fmt.Errorf("read config: %w", err)
```

### Interfaces

Define interfaces where consumed, not where implemented:

```go
// In app/app.go (consumer)
type Decomposer interface {
    Decompose(ctx context.Context, prd, progress string) error
}
```

### Naming

- Short, conventional names: `Runner`, `Tracker`, `Writer`
- Avoid stuttering: `progress.Tracker` not `progress.ProgressTracker`
- Options named `With*`: `WithModel`, `WithStdout`

### Testing

- Table-driven tests with testify
- Test file alongside implementation: `foo.go` → `foo_test.go`
- Mock via interfaces, inject in constructors

## Dependencies

Prefer stdlib. Add external dependencies only when they provide significant value over hand-rolled solutions. Don't reimplement what stdlib or well-established packages already provide.
