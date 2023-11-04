# moni

A command line tool for https://monibot.io


## Install

Needs Go 1.20 or higher, download from https://go.dev/

```
$ go install github.com/cvilsmeier/moni
```

## Usage

Moni supports many commands, list them all with:

```
$ moni help

moni v0.0.4

Monibot command line tool, see https://monibot.io.

Usage

    moni [flags] command

Flags

    -url
        Monibot URL, default is "https://monibot.io".
        You can set this also via environment variable MONIBOT_URL.

    -apiKey
        Monibot API Key, default is "".
        You can set this also via environment variable MONIBOT_API_KEY (recommended).
        You can find your API Key in your profile on https://monibot.io.
        Note: Keep your API Key secret at all times.

    -trials
        Max. Send trials, default is 12.
        You can set this also via environment variable MONIBOT_TRIALS.

    -delay
        Delay between trials, default is 5s.
        You can set this also via environment variable MONIBOT_DELAY.

    -v
        Verbose output, default is false.
        You can set this also via environment variable MONIBOT_VERBOSE ('true' or 'false').

Commands

    ping
        Ping the Monibot API.

    watchdogs
        List watchdogs.

    watchdog <watchdogId>
        Get watchdog by id.

    heartbeat <watchdogId> [interval]
        Send a heartbeat.
        If interval is specified, moni will keep sending heartbeats
        in the background. Min. interval is 1m. If interval is left
        out, moni will send one heartbeat and then exit.

    machines
        List machines.

    machine <machineId>
        Get machine by id.

    sample <machineId> [interval]
        Send resource usage (load/cpu/mem/disk) samples for machine.
        Moni consults various files (/proc/loadavg, /proc/cpuinfo)
        and commands (free, df) to calculate resource usage.
        It currently supports linux only.
        If interval is specified, moni will keep sampling in
        the background. Min. interval is 1m. If interval is left
        out, moni will send one sample and then exit.

    metrics
        List metrics.

    metric <metricId>
        Get and print metric info.

    inc <metricId> <value>
        Increment a Counter metric.
        Value must be a non-negative 64-bit integer value.

    set <metricId> <value>
        Set a Gauge metric.
        Value must be a non-negative 64-bit integer value.

    config
        Show config values.

    version
        Show program version.

    help
        Show this help page.

Exit Codes
    0 ok
    1 error
    2 wrong user input

```


## Changelog

### v0.0.0

- first version
