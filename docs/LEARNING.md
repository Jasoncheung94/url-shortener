# Print project structure

`brew install tree`
`tree -L 4`

# 2025/03/23 12:51:58 http: superfluous response.WriteHeader call from github.com/jasoncheung94/url-shortener/common.WriteJSONResponse (json.go:27)

Fix by adding a return because the code continues to execute and attempts to write to response.

# Go vet

`go vet ./...`

Helps catches suspicious constructs not caught by compiler. eg x = append(x)

# Makefile

@ in Makefile: The @ symbol prevents the command from being printed before it runs. This makes the output cleaner.

# Test coverage

github.com/jasoncheung94/url-shortener/cmd/url-shortener/main.go:17.13,25.2 3 0
File: github.com/jasoncheung94/url-shortener/cmd/url-shortener/main.go
Code block: The code starts at line 17, column 13 and ends at line 25, column 2.
Coverage: 3 statements in this code block were covered by the tests, and 0 statements were missed.

## Testing setup

// This test runs before everything else in the package. Required to setup the logger for all tests instead of manual setup per test.
func TestMain(m \*testing.M) {
// Set up test logger
common.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil)) // io.Discard

    // Run all tests
    code := m.Run()

    // Exit with the proper code
    os.Exit(code)

}

# SLOG can be disabled for tests

- slog.New(slog.NewTextHandler(io.Discard, nil))

## Redis Docker

Where is the Data Stored?
Redis stores its persistence data in disk files within the container or the host machine (if mapped via Docker volumes).

For RDB:

Redis stores RDB snapshots in a file called dump.rdb.

Default location: /data/dump.rdb inside the Redis container.

For AOF:

Redis stores AOF logs in a file called appendonly.aof.

Default location: /data/appendonly.aof inside the Redis container.

## Swagger docs

```
go install github.com/swaggo/swag/cmd/swag@latest
go get -u github.com/swaggo/http-swagger
go get -u github.com/alecthomas/template
```

Run from root: `swag init -d ./cmd/url-shortener --pdl 3`
Add import to created docs folder in main.go

## Dump data

    jsonData, _ := json.MarshalIndent(user, "", "  ")
    print .... string(jsonData)

## cyclic import

Very annoying!

Solutions: Abstract common items used in all layers. eg model.ShortURL in it's own package.
Same idea applies to interfaces.

Refactored interface.go of package Repository to be part of shortener package repository.go instead. This works but requires updating all naming of packages to x_test.go aka black box testing and exporting any functions so it can be used.
