# Organise Downloads

![Build Status](https://github.com/RMBeristain/organise-downloads/actions/workflows/go.yaml/badge.svg)

`organise-downloads` is a CLI tool to organise your `~/Downloads` folder by moving files into subdirectories based on their extensions.

If you're like me, you keep tons of downloaded files there and you tell yourself you'll "sort them out later". Your file finder can show them sorted by type, but it's still a huge list.

`organise-downloads` is here to help you sort them into folders, so you can decide faster what to keepd and what to bin (or you can keep them around for ages in their subfolders ^_^').

For example, it moves `file.exe` to an `exe_folder` subdirectory to keep things tidy, and creates the appropriate folder if it didn't exist.
```bash
# Example
/home/raider/Downloads
├── gz_files
│   └── go1.25.5.linux-amd64.tar.gz
├── json_files
│   └── someFile.json
├── log_files
│   └── organise-downloads.log
└── toml_files
    └── sampleOrganiseDownloads.toml
```

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Development](#development)

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

#### Build for windows

I run this program on windows, but I compile it from linux. You can do that with:
```bash
GOOS=windows GOARCH=amd64 go build -o organise-downloads.exe
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

### Run as a service

#### Run as a service on Linux

To run your `organise-downloads` program automatically every 20 minutes on Linux, you can use **systemd timers**.

Unlike a traditional cron job, systemd timers provide better logging through `journalctl` and ensure that if the computer is suspended, the task can trigger immediately upon waking.

#### 1. Create the Service File

The service file tells Linux *what* to run. Create a new configuration file in your user directory:

```bash
mkdir -p ~/.config/systemd/user/
vi ~/.config/systemd/user/organise-downloads.service

```

Paste the following content (replacing `/path/to/` with the actual path to your compiled Go binary):

```ini
[Unit]
Description=Organize Downloads Folder by Extension

[Service]
Type=oneshot
ExecStart=/path/to/organise-downloads

[Install]
WantedBy=default.target
```

#### 2. Create the Timer File

The timer file tells Linux *when* to run the service. The names must match (e.g., `name.service` and `name.timer`).

```bash
vi ~/.config/systemd/user/organise-downloads.timer
```

Paste the following:

```ini
[Unit]
Description=Run organise-downloads every 20 minutes

[Timer]
# Run 20 minutes after the timer is first activated
OnActiveSec=1min
# Run every 20 minutes thereafter
OnUnitActiveSec=20min
Unit=organise-downloads.service

[Install]
WantedBy=timers.target
```

#### 3. Enable and Start the Timer

Run the following commands on your terminal to register the new configuration and start the schedule:

```bash
# Reload the systemd manager configuration for the user
systemctl --user daemon-reload

# Enable the timer so it starts on boot
systemctl --user enable organise-downloads.timer

# Start the timer immediately
systemctl --user start organise-downloads.timer
```

#### Run as a service on macOS

To run your `organise-downloads` program automatically every 20 minutes on macOS, you can use **launchd**.

#### 1. Create the .plist File

Create a new file at `~/Library/LaunchAgents/com.user.organise-downloads.plist` and add the following content.

*Note: Replace `/path/to/organise-downloads` with the actual path to your compiled Go binary.*

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.user.organise-downloads</string>
    <key>ProgramArguments</key>
    <array>
        <string>/path/to/organise-downloads</string>
    </array>
    <key>StartInterval</key>
    <integer>1200</integer> <!-- 20 minutes in seconds -->
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
```

#### 2. Load the .plist File

To load the service and have it run at the specified interval, run the following command in your terminal:

```bash
launchctl load ~/Library/LaunchAgents/com.user.organise-downloads.plist
```

The program will now run every 20 minutes.

#### Unloading the service

If you want to stop the service from running, use the following command:

```bash
launchctl unload ~/Library/LaunchAgents/com.user.organise-downloads.plist
```

### Troubleshooting

`organise-downloads` creates its own logs at `~/Downloads/log_files/organise-downloads.log` by default. If a file didn't get moved as expected you can look in there for a possible cause.

`organise-downloads` won't overrwrite files with the same name. For example, if you have these files in your Downloads folder, "sampleOrganiseDownloads.toml" won't be moved because it already exists:
```bash
/home/raider/Downloads
├── log_files
│   └── organise-downloads.log
├── sampleOrganiseDownloads.toml  <--
└── toml_files
    └── sampleOrganiseDownloads.toml  <--
```

I this case you'd see this message in the log:
```json
{"level":"info","fileName":"sampleOrganiseDownloads.toml","dstFilePath":"/home/raider/Downloads/toml_files/sampleOrganiseDownloads.toml","time":"2026-01-07T10:01:26+11:00","caller":"/SourceCode/Golang/Github/organise-downloads/internal/org/org.go:77","message":"skipped"}
```

Yes, the logs are quite verbose ;) You may want to trim the file every few months; this version doesn't do it automatically.

#### macOS

To check the status of the service and view logs, you can use the following commands:

To see if the service is loaded:
```bash
launchctl list | grep com.user.organise-downloads
```

To view logs, you can check the system log for entries from your program:
```bash
log show --predicate 'process == "organise-downloads"' --last 1h
```

#### Linux

#### Check Status

To see if the timer is active and when it is scheduled to run next:

```bash
systemctl --user list-timers
```

#### View Logs

If the files aren't moving as expected, you can check the output of your Go program:

```bash
journalctl --user -u organise-downloads.service

```

#### Manual Run

If you want to trigger the organization immediately without waiting for the 20-minute interval:

```bash
systemctl --user start organise-downloads.service
```

## Development

### Testing

This project uses the standard Go testing framework.
`organise-downloads` is tested against `ubuntu26.04` and `macOS 26`. It probably does run on older systems (I used to run it on ubuntu 20) but that's not supported.

#### Run and view tests on command line

```bash
go test -v -cover ./...
```

### Run tests with HTML output

```bash
# Create the coverage dir if not exists
mkdir -pv coverage

# run tests and open in browser
go test -coverprofile=coverage/coverage.out ./... && go tool cover -html=coverage/coverage.out
```
