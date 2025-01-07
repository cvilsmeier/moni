/*
Moni is a command line tool for interacting with the
Monibot REST API, see https://monibot.io for details.
It supports a number of commands.
To get a list of supported commands, run

	$ moni help
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cvilsmeier/monibot-go"
	"github.com/cvilsmeier/monibot-go/histogram"
)

// Version is the moni tool version
const Version = "v0.3.0"

// config flag definitions
const (
	urlEnvKey  = "MONIBOT_URL"
	urlFlag    = "url"
	defaultUrl = "https://monibot.io"

	apiKeyEnvKey  = "MONIBOT_API_KEY"
	apiKeyFlag    = "apiKey"
	defaultApiKey = ""

	verboseEnvKey     = "MONIBOT_VERBOSE"
	verboseFlag       = "v"
	defaultVerbose    = false
	defaultVerboseStr = "false"

	trialsEnvKey     = "MONIBOT_TRIALS"
	trialsFlag       = "trials"
	defaultTrials    = 12
	defaultTrialsStr = "12"

	delayEnvKey  = "MONIBOT_DELAY"
	delayFlag    = "delay"
	defaultDelay = 5 * time.Second
)

// min/max values
const (
	maxMachineTextSize   = 200 * 1024
	minHeartbeatInterval = 5 * time.Minute
	minSampleInterval    = 5 * time.Minute
)

func printUsage(w io.Writer) {
	fprt(w, "moni - a command line tool for https://monibot.io")
	fprt(w, "")
	fprt(w, "usage")
	fprt(w, "")
	fprt(w, "    moni [flags] command")
	fprt(w, "")
	fprt(w, "flags")
	fprt(w, "")
	fprt(w, "    -%s", urlFlag)
	fprt(w, "        Monibot URL, default is %q.", defaultUrl)
	fprt(w, "        You can set this also via environment variable %s.", urlEnvKey)
	fprt(w, "")
	fprt(w, "    -%s", apiKeyFlag)
	fprt(w, "        Monibot API Key, default is %q.", defaultApiKey)
	fprt(w, "        You can set this also via environment variable %s", apiKeyEnvKey)
	fprt(w, "        (recommended). You can find your API Key in your profile on")
	fprt(w, "        https://monibot.io.")
	fprt(w, "        Note: For security, we recommend that you specify the API Key")
	fprt(w, "        via %s, and not via -%s flag. The flag will show", apiKeyEnvKey, apiKeyFlag)
	fprt(w, "        up in 'ps aux' outputs and can be eavesdropped.")
	fprt(w, "")
	fprt(w, "    -%s", trialsFlag)
	fprt(w, "        Max. Send trials, default is %v.", defaultTrials)
	fprt(w, "        You can set this also via environment variable %s.", trialsEnvKey)
	fprt(w, "")
	fprt(w, "    -%s", delayFlag)
	fprt(w, "        Delay between trials, default is %v.", fmtDuration(defaultDelay))
	fprt(w, "        You can set this also via environment variable %s.", delayEnvKey)
	fprt(w, "")
	fprt(w, "    -%s", verboseFlag)
	fprt(w, "        Verbose output, default is %v.", defaultVerboseStr)
	fprt(w, "        You can set this also via environment variable %s", verboseEnvKey)
	fprt(w, "        ('true' or 'false').")
	fprt(w, "")
	fprt(w, "commands")
	fprt(w, "")
	fprt(w, "    ping")
	fprt(w, "        Ping the Monibot API. If an error occurs, moni will print")
	fprt(w, "        that error. It it succeeds, moni will print nothing.")
	fprt(w, "")
	fprt(w, "    watchdogs")
	fprt(w, "        List heartbeat watchdogs.")
	fprt(w, "")
	fprt(w, "    heartbeat <watchdogId> [interval]")
	fprt(w, "        Send a heartbeat. If interval is not specified, moni sends")
	fprt(w, "        one heartbeat and exits. If interval is specified, moni")
	fprt(w, "        will stay in the background and send heartbeats in that")
	fprt(w, "        interval. Min. interval is %s.", fmtDuration(minHeartbeatInterval))
	fprt(w, "")
	fprt(w, "    machines")
	fprt(w, "        List machines.")
	fprt(w, "")
	fprt(w, "    sample <machineId> <interval>")
	fprt(w, "        Send resource usage (load/cpu/mem/disk) samples for machine.")
	fprt(w, "        Moni consults various files (/proc/loadavg, /proc/cpuinfo,")
	fprt(w, "        etc.) and commands (/usr/bin/free, /usr/bin/df, etc.) to")
	fprt(w, "        calculate resource usage. Therefore it currently supports")
	fprt(w, "        linux only. Moni will stay in background and keep sampling")
	fprt(w, "        in specified interval. Min. interval is %s.", fmtDuration(minSampleInterval))
	fprt(w, "")
	fprt(w, "    text <machineId> <filename>")
	fprt(w, "        Send filename as text for machine.")
	fprt(w, "        Filename can contain arbitrary text, e.g. arbitrary command")
	fprt(w, "        outputs. It's used for information only, no logic is")
	fprt(w, "        associated with texts. Moni will send the file as text and")
	fprt(w, "        then exit. If an error occurs, moni will print an error")
	fprt(w, "        message. Otherwise moni will print nothing.")
	fprt(w, "        Max. filesize is %d bytes.", maxMachineTextSize)
	fprt(w, "")
	fprt(w, "    metrics")
	fprt(w, "        List metrics.")
	fprt(w, "")
	fprt(w, "    inc <metricId> <value>")
	fprt(w, "        Increment a counter metric.")
	fprt(w, "        Value must be a non-negative 64-bit integer value.")
	fprt(w, "")
	fprt(w, "    set <metricId> <value>")
	fprt(w, "        Set a gauge metric value.")
	fprt(w, "        Value must be a non-negative 64-bit integer value.")
	fprt(w, "")
	fprt(w, "    values <metricId> <values>")
	fprt(w, "        Send histogram metric values.")
	fprt(w, "        Values is a comma-separated list of 'value:count' pairs.")
	fprt(w, "        Each value is a non-negative 64-bit integer value, each")
	fprt(w, "        count is an integer value greater or equal to 1.")
	fprt(w, "        If count is 1, the ':count' part is optional, so")
	fprt(w, "        values '13:1,14:1' and '13,14' are sematically equal.")
	fprt(w, "        A specific value may occur multiple times, its counts will")
	fprt(w, "        then be added together, so values '13:2,13:2' and '13:4'")
	fprt(w, "        are sematically equal.")
	fprt(w, "")
	fprt(w, "    config")
	fprt(w, "        Show config values.")
	fprt(w, "")
	fprt(w, "    version")
	fprt(w, "        Show moni program version.")
	fprt(w, "")
	fprt(w, "    sdk-version")
	fprt(w, "        Show the monibot-go SDK version moni was built with.")
	fprt(w, "")
	fprt(w, "    help")
	fprt(w, "        Show this help page.")
	fprt(w, "")
	fprt(w, "Exit Codes")
	fprt(w, "    0 ok")
	fprt(w, "    1 error")
	fprt(w, "    2 wrong user input")
}

func main() {
	log.SetOutput(os.Stdout)
	// flags
	// -url https://monibot.io
	url := os.Getenv(urlEnvKey)
	if url == "" {
		url = defaultUrl
	}
	flag.StringVar(&url, urlFlag, url, "")
	// -apiKey 007
	apiKey := os.Getenv(apiKeyEnvKey)
	if apiKey == "" {
		apiKey = defaultApiKey
	}
	flag.StringVar(&apiKey, apiKeyFlag, apiKey, "")
	// -trials 12
	trialsStr := os.Getenv(trialsEnvKey)
	if trialsStr == "" {
		trialsStr = defaultTrialsStr
	}
	trials, err := strconv.Atoi(trialsStr)
	if err != nil {
		fatal(2, "cannot parse trials %q: %s", trialsStr, err)
	}
	flag.IntVar(&trials, trialsFlag, trials, "")
	// -delay 5s
	delayStr := os.Getenv(delayEnvKey)
	if delayStr == "" {
		delayStr = fmtDuration(defaultDelay)
	}
	delay, err := time.ParseDuration(delayStr)
	if err != nil {
		fatal(2, "cannot parse delay %q: %s", delayStr, err)
	}
	flag.DurationVar(&delay, delayFlag, delay, "")
	// -v
	verboseStr := os.Getenv(verboseEnvKey)
	if verboseStr == "" {
		verboseStr = strconv.FormatBool(defaultVerbose)
	}
	verbose := verboseStr == "true"
	flag.BoolVar(&verbose, verboseFlag, verbose, "")
	// parse flags
	flag.Usage = func() { printUsage(os.Stdout) }
	flag.Parse()
	// execute non-API commands
	command := flag.Arg(0)
	switch command {
	case "", "help":
		printUsage(os.Stdout)
		os.Exit(0)
	case "config":
		prt("url             %v", url)
		prt("apiKey          %v", apiKey)
		prt("trials          %v", trials)
		prt("delay           %v", fmtDuration(delay))
		prt("verbose         %v", verbose)
		os.Exit(0)
	case "version":
		prt("moni %s", Version)
		os.Exit(0)
	case "sdk-version":
		prt("monibot-go %s", monibot.Version)
		os.Exit(0)
	}
	// validate flags
	if url == "" {
		fatal(2, "empty url")
	}
	if apiKey == "" {
		fatal(2, "empty apiKey")
	}
	const minTrials = 1
	const maxTrials = 100
	if trials < minTrials {
		fatal(2, "invalid trials %v, must be >= %v", trials, minTrials)
	} else if trials > maxTrials {
		fatal(2, "invalid trials %v, must be <= %v", trials, maxTrials)
	}
	const minDelay = 1 * time.Second
	const maxDelay = 24 * time.Hour
	if delay < minDelay {
		fatal(2, "invalid delay %s, must be >= %s", fmtDuration(delay), fmtDuration(minDelay))
	} else if delay > maxDelay {
		fatal(2, "invalid delay %s, must be <= %s", fmtDuration(delay), fmtDuration(maxDelay))
	}
	// init monibot Api
	var options monibot.ApiOptions
	if verbose {
		options.Logger = &ApiLogger{}
	}
	options.MonibotUrl = url
	options.Trials = trials
	options.Delay = delay
	api := monibot.NewApiWithOptions(apiKey, options)
	// execute API commands
	switch command {
	case "ping":
		// moni ping
		err := api.GetPing()
		if err != nil {
			fatal(1, "%s", err)
		}
	case "watchdogs":
		// moni watchdogs
		watchdogs, err := api.GetWatchdogs()
		if err != nil {
			fatal(1, "%s", err)
		}
		printWatchdogs(watchdogs)
	case "heartbeat":
		// moni heartbeat <watchdogId> [interval]
		watchdogId := flag.Arg(1)
		if watchdogId == "" {
			fatal(2, "empty watchdogId")
		}
		var interval time.Duration
		intervalStr := flag.Arg(2)
		if intervalStr != "" {
			interval, err = time.ParseDuration(intervalStr)
			if err != nil {
				fatal(2, "cannot parse interval %q: %s", intervalStr, err)
			}
			if interval < minHeartbeatInterval {
				log.Printf("WARNING: interval %s is below min, force-changing it to %s", fmtDuration(interval), fmtDuration(minHeartbeatInterval))
				interval = minHeartbeatInterval
			}
			log.Printf("INFO: will send heartbeats in background every %s", fmtDuration(interval))
		}
		err := api.PostWatchdogHeartbeat(watchdogId)
		if err != nil {
			fatal(1, "cannot send heartbeat: %s", err)
		}
		if interval > 0 {
			// enter heartbeat loop
			for {
				// sleep
				time.Sleep(interval)
				// send
				err := api.PostWatchdogHeartbeat(watchdogId)
				if err != nil {
					prt("WARNING: cannot send heartbeat: %s", err)
				}
			}
		}
	case "machines":
		// moni machines
		machines, err := api.GetMachines()
		if err != nil {
			fatal(1, "%s", err)
		}
		printMachines(machines)
	case "sample":
		// moni sample <machineId> <interval>
		machineId := flag.Arg(1)
		if machineId == "" {
			fatal(2, "empty machineId")
		}
		intervalStr := flag.Arg(2)
		if intervalStr == "" {
			fatal(2, "empty interval")
		}
		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			fatal(2, "cannot parse interval %q: %s", intervalStr, err)
		}
		if interval < minSampleInterval {
			log.Printf("WARNING: interval %s is below min, force-changing it to %s", fmtDuration(interval), fmtDuration(minSampleInterval))
			interval = minSampleInterval
		}
		sampler := NewSampler(Platform{})
		// we must warm up the sampler first
		_, err = sampler.Sample()
		if err != nil {
			fatal(1, "cannot sample: %s", err)
		}
		// entering sampling loop
		log.Printf("INFO: will send samples in background every %s", fmtDuration(interval))
		for {
			// sleep
			time.Sleep(interval)
			// sample
			sample, err := sampler.Sample()
			if err != nil {
				prt("WARNING: cannot sample: %s", err)
			}
			err = api.PostMachineSample(machineId, sample)
			if err != nil {
				prt("WARNING: cannot POST sample: %s", err)
			}
		}
	case "text":
		// moni text <machineId> <filename>
		machineId := flag.Arg(1)
		if machineId == "" {
			fatal(2, "empty machineId")
		}
		filename := flag.Arg(2)
		if filename == "" {
			fatal(2, "empty filename")
		}
		filedata, err := os.ReadFile(filename)
		if err != nil {
			fatal(2, "cannot read %s: %s", filename, err)
		}
		if len(filedata) > maxMachineTextSize {
			fatal(2, "file %s too big: %d bytes (max is %d)", filename, len(filedata), maxMachineTextSize)
		}
		err = api.PostMachineText(machineId, string(filedata))
		if err != nil {
			fatal(1, "%s", err)
		}
	case "metrics":
		// moni metrics
		metrics, err := api.GetMetrics()
		if err != nil {
			fatal(1, "%s", err)
		}
		printMetrics(metrics)
	case "inc":
		// moni inc <metricId> <value>
		metricId := flag.Arg(1)
		if metricId == "" {
			fatal(2, "empty metricId")
		}
		valueStr := flag.Arg(2)
		if valueStr == "" {
			fatal(2, "empty value")
		}
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			fatal(2, "cannot parse value %q: %s", valueStr, err)
		}
		err = api.PostMetricInc(metricId, value)
		if err != nil {
			fatal(1, "%s", err)
		}
	case "set":
		// moni set <metricId> <value>
		metricId := flag.Arg(1)
		if metricId == "" {
			fatal(2, "empty metricId")
		}
		valueStr := flag.Arg(2)
		if valueStr == "" {
			fatal(2, "empty value")
		}
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			fatal(2, "cannot parse value %q: %s", valueStr, err)
		}
		err = api.PostMetricSet(metricId, value)
		if err != nil {
			fatal(1, "%s", err)
		}
	case "values":
		// moni values <metricId> <values>
		metricId := flag.Arg(1)
		if metricId == "" {
			fatal(2, "empty metricId")
		}
		valuesStr := flag.Arg(2)
		if valuesStr == "" {
			fatal(2, "empty values")
		}
		values, err := histogram.ParseValues(valuesStr)
		if err != nil {
			fatal(2, "cannot parse values: %s", err)
		}
		err = api.PostMetricValues(metricId, values)
		if err != nil {
			fatal(1, "%s", err)
		}
	default:
		fatal(2, "unknown command %q, run 'moni help'", command)
	}
}

// fatal prints a message to stdout and exits with exitCode.
func fatal(exitCode int, f string, a ...any) {
	fmt.Printf(f+"\n", a...)
	os.Exit(exitCode)
}

func fmtDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = strings.TrimSuffix(s, "0s")
	}
	if strings.HasSuffix(s, "h0m") {
		s = strings.TrimSuffix(s, "0m")
	}
	return s
}

// printWatchdogs prints watchdogs.
func printWatchdogs(watchdogs []monibot.Watchdog) {
	prt("%-35s | %-25s | %s", "Id", "Name", "IntervalMillis")
	for _, watchdog := range watchdogs {
		prt("%-35s | %-25s | %d", watchdog.Id, watchdog.Name, watchdog.IntervalMillis)
	}
}

// printMachines prints machines.
func printMachines(machines []monibot.Machine) {
	prt("%-35s | %s", "Id", "Name")
	for _, machine := range machines {
		prt("%-35s | %s", machine.Id, machine.Name)
	}
}

// printMetrics prints metrics.
func printMetrics(metrics []monibot.Metric) {
	prt("%-35s | %-25s | %s", "Id", "Name", "Type")
	for _, metric := range metrics {
		typeSuffix := ""
		switch metric.Type {
		case 0:
			typeSuffix = " (Counter)"
		case 1:
			typeSuffix = " (Gauge)"
		case 2:
			typeSuffix = " (Histogram)"
		}
		prt("%-35s | %-25s | %d%s", metric.Id, metric.Name, metric.Type, typeSuffix)
	}
}

// prt prints a line to stdout.
func prt(f string, a ...any) {
	fprt(os.Stdout, f, a...)
}

// fprt prints a line to a io.Writer.
func fprt(w io.Writer, f string, a ...any) {
	fmt.Fprintf(w, f+"\n", a...)
}

// ApiLogger logs monibot debug messages
type ApiLogger struct{}

func (l *ApiLogger) Debug(format string, args ...any) {
	log.Printf("[MONIBOT] "+format, args...)
}
