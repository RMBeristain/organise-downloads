# Organise Downloads

Organise `~/Downloads` folders.

For example, move 'file.exe' to 'exe_folder' subdir to keep things tidy.

Can add exceptions.

For options run with `-help`

## Testing

This project uses vanilla tests.

### Run and view tests on command line

```zsh
go test -v -cover ./...
```

### Run tests with HTML output

```zsh
# Create the coverage dir if not exists
mkdir -pv coverage

# run tests and open in browser
go test -coverprofile=coverage/coverage.out ./... && go tool cover -html=coverage/coverage.out
```
