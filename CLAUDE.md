# CLAUDE.md 

## Coding conventions

Always cache `regexp.MustCompile` as global value for better performance.

## Tmp folder

If you want to create temp files or do experimental tasks, always create or output files to the project's tmp folder. They should not be tracked by git.

## Testing

This project uses **`github.com/ysmood/got`** - a minimal testing library with fluent API:

```go
g := got.T(t)
g.E(err)                    // Assert no error
g.Eq(actual, expected)      // Assert equality
g.Neq(actual, expected)     // Assert inequality
g.Req("", url).String()     // Make HTTP request and get body as string
g.Req("", url, client)      // Make request with custom client
```

**Important**: Use `g.E(err)` NOT `if err != nil` in tests. Always chain assertions fluently.

Don't mock db or http calls. Use real requests and a test database.

Use the testcontainers to spin up ephemeral databases.

The test package name should use `_test` suffix.

## Error handling

Standard Go error wrapping with context:

```go
return nil, fmt.Errorf("failed to get proxy: %w", err)
```

Never hide any error.

## Linting

Use `golangci-lint` executable under the repo for linting.

If you want to use `go build` always output the executable to `tmp` dir.
