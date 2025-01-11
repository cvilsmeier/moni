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
const Version = "v0.4.0"

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
	fprtf(w, "moni - a command line tool for https://monibot.io")
	fprtf(w, "")
	fprtf(w, "usage")
	fprtf(w, "")
	fprtf(w, "    moni [flags] command")
	fprtf(w, "")
	fprtf(w, "flags")
	fprtf(w, "")
	fprtf(w, "    -%s", urlFlag)
	fprtf(w, "        Monibot URL, default is %q.", defaultUrl)
	fprtf(w, "        You can set this also via environment variable %s.", urlEnvKey)
	fprtf(w, "")
	fprtf(w, "    -%s", apiKeyFlag)
	fprtf(w, "        Monibot API Key, default is %q.", defaultApiKey)
	fprtf(w, "        You can set this also via environment variable %s", apiKeyEnvKey)
	fprtf(w, "        (recommended). You can find your API Key in your profile on")
	fprtf(w, "        https://monibot.io.")
	fprtf(w, "        Note: For security, we recommend that you specify the API Key")
	fprtf(w, "        via %s, and not via -%s flag. The flag will show", apiKeyEnvKey, apiKeyFlag)
	fprtf(w, "        up in 'ps aux' outputs and can be eavesdropped.")
	fprtf(w, "")
	fprtf(w, "    -%s", trialsFlag)
	fprtf(w, "        Max. Send trials, default is %v.", defaultTrials)
	fprtf(w, "        You can set this also via environment variable %s.", trialsEnvKey)
	fprtf(w, "")
	fprtf(w, "    -%s", delayFlag)
	fprtf(w, "        Delay between trials, default is %v.", fmtDuration(defaultDelay))
	fprtf(w, "        You can set this also via environment variable %s.", delayEnvKey)
	fprtf(w, "")
	fprtf(w, "    -%s", verboseFlag)
	fprtf(w, "        Verbose output, default is %v.", defaultVerboseStr)
	fprtf(w, "        You can set this also via environment variable %s", verboseEnvKey)
	fprtf(w, "        ('true' or 'false').")
	fprtf(w, "")
	fprtf(w, "commands")
	fprtf(w, "")
	fprtf(w, "    ping")
	fprtf(w, "        Ping the Monibot API. If an error occurs, moni will print")
	fprtf(w, "        that error. If it succeeds, it will print nothing.")
	fprtf(w, "")
	fprtf(w, "    watchdogs")
	fprtf(w, "        List heartbeat watchdogs.")
	fprtf(w, "")
	fprtf(w, "    watchdog <watchdogId>")
	fprtf(w, "        Get heartbeat watchdog by id.")
	fprtf(w, "")
	fprtf(w, "    heartbeat <watchdogId> [interval]")
	fprtf(w, "        Send a heartbeat. If interval is not specified, moni sends")
	fprtf(w, "        one heartbeat and exits. If interval is specified, moni")
	fprtf(w, "        will stay in the background and send heartbeats in that")
	fprtf(w, "        interval. Minimum interval is %s.", fmtDuration(minHeartbeatInterval))
	fprtf(w, "")
	fprtf(w, "    machines")
	fprtf(w, "        List machines.")
	fprtf(w, "")
	fprtf(w, "    machine <machineId>")
	fprtf(w, "        Get machine by id.")
	fprtf(w, "")
	fprtf(w, "    sample <machineId> <interval>")
	fprtf(w, "        Send resource usage (load/cpu/mem/disk) samples for machine.")
	fprtf(w, "        This command will stay in background and keep sampling")
	fprtf(w, "        in specified interval. Minimum interval is %s.", fmtDuration(minSampleInterval))
	fprtf(w, "")
	fprtf(w, "    text <machineId> <filename>")
	fprtf(w, "        Send filename as text for machine.")
	fprtf(w, "        Filename can contain arbitrary text, e.g. arbitrary command")
	fprtf(w, "        outputs. It's used for information only, no logic is")
	fprtf(w, "        associated with texts. Moni will send the file as text and")
	fprtf(w, "        then exit. If an error occurs, moni will print an error")
	fprtf(w, "        message. Otherwise moni will print nothing.")
	fprtf(w, "        Maximum filesize is %dK.", (maxMachineTextSize / 1024))
	fprtf(w, "")
	fprtf(w, "    metrics")
	fprtf(w, "        List metrics.")
	fprtf(w, "")
	fprtf(w, "    metric <metricId>")
	fprtf(w, "        Get metric by id.")
	fprtf(w, "")
	fprtf(w, "    inc <metricId> <value>")
	fprtf(w, "        Increment a counter metric.")
	fprtf(w, "        Value must be a non-negative 64-bit integer value.")
	fprtf(w, "")
	fprtf(w, "    set <metricId> <value>")
	fprtf(w, "        Set a gauge metric value.")
	fprtf(w, "        Value must be a non-negative 64-bit integer value.")
	fprtf(w, "")
	fprtf(w, "    values <metricId> <values>")
	fprtf(w, "        Send histogram metric values.")
	fprtf(w, "        Values is a comma-separated list of 'value:count' pairs.")
	fprtf(w, "        Each value is a non-negative 64-bit integer value, each")
	fprtf(w, "        count is an integer value greater or equal to 1.")
	fprtf(w, "        If count is 1, the ':count' part is optional, so")
	fprtf(w, "        values '13:1,14:1' and '13,14' are sematically equal.")
	fprtf(w, "        A specific value may occur multiple times, its counts will")
	fprtf(w, "        then be added together, so values '13:2,13:2' and '13:4'")
	fprtf(w, "        are sematically equal.")
	fprtf(w, "")
	fprtf(w, "    config")
	fprtf(w, "        Show config values.")
	fprtf(w, "")
	fprtf(w, "    version")
	fprtf(w, "        Show moni program version.")
	fprtf(w, "")
	fprtf(w, "    sdk-version")
	fprtf(w, "        Show the monibot-go SDK version moni was built with.")
	fprtf(w, "")
	fprtf(w, "    help")
	fprtf(w, "        Show this help page.")
	fprtf(w, "")
	fprtf(w, "Exit Codes")
	fprtf(w, "    0 ok")
	fprtf(w, "    1 error")
	fprtf(w, "    2 wrong user input")
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
	var devMode bool
	flag.BoolVar(&devMode, "dev", devMode, "")
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
		prtf("url             %v", url)
		prtf("apiKey          %v", apiKey)
		prtf("trials          %v", trials)
		prtf("delay           %v", fmtDuration(delay))
		prtf("verbose         %v", verbose)
		if devMode {
			prtf("devMode         %v", devMode)
		}
		os.Exit(0)
	case "version":
		prtf("moni %s", Version)
		os.Exit(0)
	case "sdk-version":
		prtf("monibot-go %s", monibot.Version)
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
	case "watchdog":
		// moni watchdog <watchdogId>
		watchdogId := flag.Arg(1)
		if watchdogId == "" {
			fatal(2, "empty watchdogId")
		}
		watchdog, err := api.GetWatchdog(watchdogId)
		if err != nil {
			fatal(1, "%s", err)
		}
		printWatchdogs([]monibot.Watchdog{watchdog})
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
			if interval < minHeartbeatInterval && !devMode {
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
					prtf("WARNING: cannot send heartbeat: %s", err)
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
	case "machine":
		// moni machine <machineId>
		machineId := flag.Arg(1)
		if machineId == "" {
			fatal(2, "empty machineId")
		}
		machine, err := api.GetMachine(machineId)
		if err != nil {
			fatal(1, "%s", err)
		}
		printMachines([]monibot.Machine{machine})
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
		if interval < minSampleInterval && !devMode {
			log.Printf("WARNING: interval %s is below min, force-changing it to %s", fmtDuration(interval), fmtDuration(minSampleInterval))
			interval = minSampleInterval
		}
		sampler := NewSampler(NewPlatform(verbose))
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
				prtf("WARNING: cannot sample: %s", err)
			}
			err = api.PostMachineSample(machineId, sample)
			if err != nil {
				prtf("WARNING: cannot POST sample: %s", err)
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
	case "metric":
		// moni metric <metricId>
		metricId := flag.Arg(1)
		if metricId == "" {
			fatal(2, "empty metricId")
		}
		metric, err := api.GetMetric(metricId)
		if err != nil {
			fatal(1, "%s", err)
		}
		printMetrics([]monibot.Metric{metric})
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
	prtf(f+"\n", a...)
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
	prtf("%-35s | %-25s | %s", "Id", "Name", "IntervalMillis")
	for _, watchdog := range watchdogs {
		prtf("%-35s | %-25s | %d", watchdog.Id, watchdog.Name, watchdog.IntervalMillis)
	}
}

// printMachines prints machines.
func printMachines(machines []monibot.Machine) {
	prtf("%-35s | %s", "Id", "Name")
	for _, machine := range machines {
		prtf("%-35s | %s", machine.Id, machine.Name)
	}
}

// printMetrics prints metrics.
func printMetrics(metrics []monibot.Metric) {
	prtf("%-35s | %-25s | %s", "Id", "Name", "Type")
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
		prtf("%-35s | %-25s | %d%s", metric.Id, metric.Name, metric.Type, typeSuffix)
	}
}

// prtf prints a line to stdout.
func prtf(f string, a ...any) {
	fprtf(os.Stdout, f, a...)
}

// fprtf prints a line to a io.Writer.
func fprtf(w io.Writer, f string, a ...any) {
	fmt.Fprintf(w, f+"\n", a...)
}

// ApiLogger logs monibot debug messages
type ApiLogger struct{}

func (l *ApiLogger) Debug(format string, args ...any) {
	log.Printf("VERBOSE: (API) "+format, args...)
}
