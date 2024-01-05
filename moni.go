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
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cvilsmeier/moni/internal"
	"github.com/cvilsmeier/monibot-go"
)

// Version is the moni tool version
const Version = "v0.1.0"

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

// variuos other definitions
const (
	maxMachineTextSize = 200 * 1024
	minBeatInterval    = 5 * time.Minute
	minSampleInterval  = 5 * time.Minute
)

func usage() {
	prt("Moni - A command line tool for https://monibot.io")
	prt("")
	prt("Usage")
	prt("")
	prt("    moni [flags] command")
	prt("")
	prt("Flags")
	prt("")
	prt("    -%s", urlFlag)
	prt("        Monibot URL, default is %q.", defaultUrl)
	prt("        You can set this also via environment variable %s.", urlEnvKey)
	prt("")
	prt("    -%s", apiKeyFlag)
	prt("        Monibot API Key, default is %q.", defaultApiKey)
	prt("        You can set this also via environment variable %s", apiKeyEnvKey)
	prt("        (recommended). You can find your API Key in your profile on")
	prt("        https://monibot.io.")
	prt("        Note: For security, we recommend that you specify the API Key")
	prt("        via %s, and not via -%s flag. The flag will show", apiKeyEnvKey, apiKeyFlag)
	prt("        up in 'ps aux' outputs and can be eavesdropped.")
	prt("")
	prt("    -%s", trialsFlag)
	prt("        Max. Send trials, default is %v.", defaultTrials)
	prt("        You can set this also via environment variable %s.", trialsEnvKey)
	prt("")
	prt("    -%s", delayFlag)
	prt("        Delay between trials, default is %v.", fmtDuration(defaultDelay))
	prt("        You can set this also via environment variable %s.", delayEnvKey)
	prt("")
	prt("    -%s", verboseFlag)
	prt("        Verbose output, default is %v.", defaultVerboseStr)
	prt("        You can set this also via environment variable %s", verboseEnvKey)
	prt("        ('true' or 'false').")
	prt("")
	prt("Commands")
	prt("")
	prt("    ping")
	prt("        Ping the Monibot API. If an error occurs, moni will print")
	prt("        that error. It it succeeds, moni will print nothing.")
	prt("")
	prt("    watchdogs")
	prt("        List heartbeat watchdogs.")
	prt("")
	prt("    watchdog <watchdogId>")
	prt("        Get heartbeat watchdog by id.")
	prt("")
	prt("    beat <watchdogId> [interval]")
	prt("        Send a heartbeat. If interval is not specified, moni sends")
	prt("        one heartbeat and exits. If interval is specified, moni")
	prt("        will stay in the background and send heartbeats in that")
	prt("        interval. Min. interval is %s.", fmtDuration(minBeatInterval))
	prt("")
	prt("    machines")
	prt("        List machines.")
	prt("")
	prt("    machine <machineId>")
	prt("        Get machine by id.")
	prt("")
	prt("    sample <machineId> <interval>")
	prt("        Send resource usage (load/cpu/mem/disk) samples for machine.")
	prt("        Moni consults various files (/proc/loadavg, /proc/cpuinfo,")
	prt("        etc.) and commands (/usr/bin/free, /usr/bin/df, etc.) to")
	prt("        calculate resource usage. Therefore it currently supports")
	prt("        linux only. Moni will stay in background and keep sampling")
	prt("        in specified interval. Min. interval is %s.", fmtDuration(minSampleInterval))
	prt("")
	prt("    text <machineId> <filename>")
	prt("        Send filename as text for machine.")
	prt("        Filename can contain arbitrary text, e.g. arbitrary command")
	prt("        outputs. It's used for information only, no logic is")
	prt("        associated with texts. Moni will send the file as text and")
	prt("        then exit. If an error occurs, moni will print an error")
	prt("        message. Otherwise moni will print nothing.")
	prt("        Max. filesize is %d bytes.", maxMachineTextSize)
	prt("")
	prt("    metrics")
	prt("        List metrics.")
	prt("")
	prt("    metric <metricId>")
	prt("        Get and print metric info.")
	prt("")
	prt("    inc <metricId> <value>")
	prt("        Increment a counter metric.")
	prt("        Value must be a non-negative 64-bit integer value.")
	prt("")
	prt("    set <metricId> <value>")
	prt("        Set a gauge metric value.")
	prt("        Value must be a non-negative 64-bit integer value.")
	prt("")
	prt("    values <metricId> <values>")
	prt("        Set histogram metric values.")
	prt("        Values must be a list of non-negative 64-bit integer")
	prt("        values, for example \"0,12,16,16,1,2\".")
	prt("")
	prt("    config")
	prt("        Show config values.")
	prt("")
	prt("    version")
	prt("        Show moni program version.")
	prt("")
	prt("    sdk-version")
	prt("        Show the monibot-go SDK version moni was built with.")
	prt("")
	prt("    help")
	prt("        Show this help page.")
	prt("")
	prt("Exit Codes")
	prt("    0 ok")
	prt("    1 error")
	prt("    2 wrong user input")
	prt("")
	// ---------|---------|---------|---------|---------|---------|---------|
	// end usage
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
	flag.Usage = usage
	flag.Parse()
	// execute non-API commands
	command := flag.Arg(0)
	switch command {
	case "", "help":
		usage()
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
		fatal(2, "invalid trials %s, must be >= %s", trials, minTrials)
	} else if trials > maxTrials {
		fatal(2, "invalid trials %s, must be <= %s", trials, maxTrials)
	}
	const minDelay = 1 * time.Second
	const maxDelay = 24 * time.Hour
	if delay < minDelay {
		fatal(2, "invalid delay %s, must be >= %s", delay, minDelay)
	} else if delay > maxDelay {
		fatal(2, "invalid delay %s, must be <= %s", delay, maxDelay)
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
	case "beat":
		// moni beat <watchdogId> [interval]
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
			if interval < minBeatInterval {
				log.Printf("WARNING: interval " + fmtDuration(interval) + " is below min, force-changing it to " + fmtDuration(minBeatInterval))
				interval = minBeatInterval
			}
			log.Printf("will send heartbeats in background every %s", fmtDuration(interval))
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
		if interval < minSampleInterval {
			log.Printf("WARNING: interval " + fmtDuration(interval) + " is below min, force-changing it to " + fmtDuration(minSampleInterval))
			interval = minSampleInterval
		}
		sampler := internal.NewSampler()
		// we must warm up the sampler first
		_, err = sampler.Sample()
		if err != nil {
			fatal(1, "cannot sample: %s", err)
		}
		// entering sampling loop
		log.Printf("will send samples in background every %s", fmtDuration(interval))
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
		var values []int64
		toks := strings.Split(valuesStr, ",")
		for _, tok := range toks {
			tok = strings.TrimSpace(tok)
			// skip empty
			if len(tok) == 0 {
				continue
			}
			// parse int64 value
			value, err := strconv.ParseInt(tok, 10, 64)
			if err != nil {
				fatal(2, "cannot parse value %q: %s", tok, err)
			}
			values = append(values, value)
		}
		err := api.PostMetricValues(metricId, values)
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
	fmt.Printf(f+"\n", a...)
}

// ApiLogger logs monibot debug messages
type ApiLogger struct{}

func (l *ApiLogger) Debug(format string, args ...any) {
	log.Printf("[MONIBOT] "+format, args...)
}
