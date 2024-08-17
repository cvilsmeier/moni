# moni - A command line tool for Monibot

[![GoDoc Reference](https://godoc.org/github.com/cvilsmeier/moni?status.svg)](http://godoc.org/github.com/cvilsmeier/moni)
[![Build Status](https://github.com/cvilsmeier/moni/actions/workflows/go-linux.yml/badge.svg)](https://github.com/cvilsmeier/moni/actions/workflows/go-linux.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A command line tool for https://monibot.io - Easy Server and Application Monitoring.

Moni is a command-line tool to interact with the Monibot REST API. It is used to

- Query watchdogs from Monibot
- Send watchdog heartbeats to Monibot
- Query machines from Monibot
- Send machine resource usage samples (CPU/Memory/Disk/Network) to Monibot
- Send machine texts (command outputs) to Monibot
- Query metrics from Monibot
- Send metric values to Monibot

It runs on Linux/amd64 and is written in Go.

## Download

Download a pre-built linux/amd64 binary from here:
https://github.com/cvilsmeier/moni/releases/latest


## Install

If you do not want to download a pre-built binary, you
can install moni from the command line. You need
Go 1.20 or higher, see https://go.dev/

```
$ go install github.com/cvilsmeier/moni@latest
```


## Build

If you do not want to download a pre-built binary, you
can build moni from the command line. You need
Go 1.20 or higher. For installing Go, see https://go.dev/

```
$ git clone https://github.com/cvilsmeier/moni
$ cd moni/
$ CGO_ENABLED=0 go build
$ ./moni help
```


## Usage

```
$ moni help
Moni - A command line tool for https://monibot.io

Usage

    moni [flags] command

Flags

    -url
        Monibot URL, default is "https://monibot.io".
        You can set this also via environment variable MONIBOT_URL.

    -apiKey
        Monibot API Key, default is "".
        You can set this also via environment variable MONIBOT_API_KEY
        (recommended). You can find your API Key in your profile on
        https://monibot.io.
        Note: For security, we recommend that you specify the API Key
        via MONIBOT_API_KEY, and not via -apiKey flag. The flag will show
        up in 'ps aux' outputs and can be eavesdropped.

    -trials
        Max. Send trials, default is 12.
        You can set this also via environment variable MONIBOT_TRIALS.

    -delay
        Delay between trials, default is 5s.
        You can set this also via environment variable MONIBOT_DELAY.

    -v
        Verbose output, default is false.
        You can set this also via environment variable MONIBOT_VERBOSE
        ('true' or 'false').

Commands

    ping
        Ping the Monibot API. If an error occurs, moni will print
        that error. It it succeeds, moni will print nothing.

    watchdogs
        List heartbeat watchdogs.

    watchdog <watchdogId>
        Get heartbeat watchdog by id.

    heartbeat <watchdogId> [interval]
        Send a heartbeat. If interval is not specified, moni sends
        one heartbeat and exits. If interval is specified, moni
        will stay in the background and send heartbeats in that
        interval. Min. interval is 5m.

    machines
        List machines.

    machine <machineId>
        Get machine by id.

    sample <machineId> <interval>
        Send resource usage (load/cpu/mem/disk) samples for machine.
        Moni consults various files (/proc/loadavg, /proc/cpuinfo,
        etc.) and commands (/usr/bin/free, /usr/bin/df, etc.) to
        calculate resource usage. Therefore it currently supports
        linux only. Moni will stay in background and keep sampling
        in specified interval. Min. interval is 5m.

    text <machineId> <filename>
        Send filename as text for machine.
        Filename can contain arbitrary text, e.g. arbitrary command
        outputs. It's used for information only, no logic is
        associated with texts. Moni will send the file as text and
        then exit. If an error occurs, moni will print an error
        message. Otherwise moni will print nothing.
        Max. filesize is 204800 bytes.

    metrics
        List metrics.

    metric <metricId>
        Get and print metric info.

    inc <metricId> <value>
        Increment a counter metric.
        Value must be a non-negative 64-bit integer value.

    set <metricId> <value>
        Set a gauge metric value.
        Value must be a non-negative 64-bit integer value.

    values <metricId> <values>
        Send histogram metric values.
        Values is a comma-separated list of 'value:count' pairs.
        Each value is a non-negative 64-bit integer value, each
        count is an integer value greater or equal to 1.
        If count is 1, the ':count' part is optional, so
        values '13:1,14:1' and '13,14' are sematically equal.
        A specific value may occur multiple times, its counts will
        then be added together, so values '13:2,13:2' and '13:4'
        are sematically equal.

    config
        Show config values.

    version
        Show moni program version.

    sdk-version
        Show the monibot-go SDK version moni was built with.

    help
        Show this help page.

Exit Codes
    0 ok
    1 error
    2 wrong user input
```


## Changelog

### v0.2.2

- renamed beat command to heartbeat

### v0.2.1

- update values command docs

### v0.2.0

- add values command for sending histogram metric values

### v0.1.0

- add text command for sending machine text
- breaking change: set min watchdog beat interval 5m
- breaking change: set min machine sample interval 5m

### v0.0.2

- change command names and interval handling

### v0.0.1

- add disk read/write sampling
- add network recv/send sampling

### v0.0.0

- first version
