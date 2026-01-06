# Organise Downloads

![Build Status](https://github.com/RMBeristain/organise-downloads/actions/workflows/go.yaml/badge.svg)

A CLI tool to organise your `~/Downloads` folder by moving files into subdirectories based on their extensions.

For example, it moves `file.exe` to an `exe_folder` subdirectory to keep things tidy.

## Features

- **Automatic Sorting**: Organises files into folders based on file type.
- **Exceptions**: Supports adding exceptions for specific files.
- **Cross-Platform**: Supports recent Linux and macOS versions.

## Installation

### Prerequisites

- Go installed.

### Build from Source

To build from source, clone the repository and run `go build` from the root folder:

```bash
go build -o organise-downloads
```

## Usage

Run the program to organise your downloads:

```bash
./organise-downloads
```

To see available options and configure exceptions:

```bash
./organise-downloads -help
```

## Development

### Testing

This project uses the standard Go testing framework.

#### Run and view tests on command line

```bash
go test -v -cover ./...
```

### Run tests with HTML output

```zsh
# Create the coverage dir if not exists
mkdir -pv coverage

# run tests and open in browser
go test -coverprofile=coverage/coverage.out ./... && go tool cover -html=coverage/coverage.out
```
