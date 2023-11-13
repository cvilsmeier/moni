/*
Moni is a command line tool for interacting with the
Monibot REST API, see https://monibot.io for details.
It supports a number of commands.
To get a list of supported commands, run

	$ moni
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

	sampleIntervalEnvKey  = "MONIBOT_SAMPLE_INTERVAL"
	sampleIntervalFlag    = "sampleInterval"
	defaultSampleInterval = 5 * time.Minute
)

func usage() {
	internal.Print("Moni - A command line tool for https://monibot.io")
	internal.Print("")
	internal.Print("Usage")
	internal.Print("")
	internal.Print("    moni [flags] command")
	internal.Print("")
	internal.Print("Flags")
	internal.Print("")
	internal.Print("    -%s", urlFlag)
	internal.Print("        Monibot URL, default is %q.", defaultUrl)
	internal.Print("        You can set this also via environment variable %s.", urlEnvKey)
	internal.Print("")
	internal.Print("    -%s", apiKeyFlag)
	internal.Print("        Monibot API Key, default is %q.", defaultApiKey)
	internal.Print("        You can set this also via environment variable %s (recommended).", apiKeyEnvKey)
	internal.Print("        You can find your API Key in your profile on https://monibot.io.")
	internal.Print("        Note: For security, we recommend that you specify the API Key")
	internal.Print("        via %s, and not via -%s flag. The flag will show up in", apiKeyEnvKey, apiKeyFlag)
	internal.Print("        'ps aux' outputs and can be eavesdropped.")
	internal.Print("")
	internal.Print("    -%s", trialsFlag)
	internal.Print("        Max. Send trials, default is %v.", defaultTrials)
	internal.Print("        You can set this also via environment variable %s.", trialsEnvKey)
	internal.Print("")
	internal.Print("    -%s", delayFlag)
	internal.Print("        Delay between trials, default is %v.", fmtDuration(defaultDelay))
	internal.Print("        You can set this also via environment variable %s.", delayEnvKey)
	internal.Print("")
	internal.Print("    -%s", sampleIntervalFlag)
	internal.Print("        Machine sample interval, default is %v.", fmtDuration(defaultSampleInterval))
	internal.Print("        You can set this also via environment variable %s.", sampleIntervalEnvKey)
	internal.Print("        This flag is only relevant for the 'sample' command.")
	internal.Print("")
	internal.Print("    -%s", verboseFlag)
	internal.Print("        Verbose output, default is %v.", defaultVerboseStr)
	internal.Print("        You can set this also via environment variable %s ('true' or 'false').", verboseEnvKey)
	internal.Print("")
	internal.Print("Commands")
	internal.Print("")
	internal.Print("    ping")
	internal.Print("        Ping the Monibot API.")
	internal.Print("")
	internal.Print("    watchdogs")
	internal.Print("        List watchdogs.")
	internal.Print("")
	internal.Print("    watchdog <watchdogId>")
	internal.Print("        Get watchdog by id.")
	internal.Print("")
	internal.Print("    heartbeat <watchdogId>")
	internal.Print("        Send a heartbeat.")
	internal.Print("")
	internal.Print("    machines")
	internal.Print("        List machines.")
	internal.Print("")
	internal.Print("    machine <machineId>")
	internal.Print("        Get machine by id.")
	internal.Print("")
	internal.Print("    sample <machineId>")
	internal.Print("        Send resource usage (load/cpu/mem/disk) samples for machine.")
	internal.Print("        Moni consults various files (/proc/loadavg, /proc/cpuinfo, etc.)")
	internal.Print("        and commands (/usr/bin/free, /usr/bin/df, etc.) to calculate")
	internal.Print("        resource usage. It currently supports linux only.")
	internal.Print("        Moni will stay in background and keep sampling in specified")
	internal.Print("        sample interval, default %s, see flag 'sampleInterval.", fmtDuration(defaultSampleInterval))
	internal.Print("")
	internal.Print("    metrics")
	internal.Print("        List metrics.")
	internal.Print("")
	internal.Print("    metric <metricId>")
	internal.Print("        Get and print metric info.")
	internal.Print("")
	internal.Print("    inc <metricId> <value>")
	internal.Print("        Increment a Counter metric.")
	internal.Print("        Value must be a non-negative 64-bit integer value.")
	internal.Print("")
	internal.Print("    set <metricId> <value>")
	internal.Print("        Set a Gauge metric.")
	internal.Print("        Value must be a non-negative 64-bit integer value.")
	internal.Print("")
	internal.Print("    config")
	internal.Print("        Show config values.")
	internal.Print("")
	internal.Print("    version")
	internal.Print("        Show moni program version.")
	internal.Print("")
	internal.Print("    sdk-version")
	internal.Print("        Show the monibot-go SDK version moni was built with.")
	internal.Print("")
	internal.Print("    help")
	internal.Print("        Show this help page.")
	internal.Print("")
	internal.Print("Exit Codes")
	internal.Print("    0 ok")
	internal.Print("    1 error")
	internal.Print("    2 wrong user input")
	internal.Print("")
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
	// -sampleInterval 5m
	sampleIntervalStr := os.Getenv(sampleIntervalEnvKey)
	if sampleIntervalStr == "" {
		sampleIntervalStr = fmtDuration(defaultSampleInterval)
	}
	sampleInterval, err := time.ParseDuration(sampleIntervalStr)
	if err != nil {
		fatal(2, "cannot parse sampleInterval %q: %s", sampleIntervalStr, err)
	}
	flag.DurationVar(&sampleInterval, sampleIntervalFlag, sampleInterval, "")
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
		internal.Print("url             %v", url)
		internal.Print("apiKey          %v", apiKey)
		internal.Print("trials          %v", trials)
		internal.Print("delay           %v", fmtDuration(delay))
		internal.Print("sampleInterval  %v", fmtDuration(sampleInterval))
		internal.Print("verbose         %v", verbose)
		os.Exit(0)
	case "version":
		internal.Print("moni %s", internal.Version)
		os.Exit(0)
	case "sdk-version":
		internal.Print("monibot-go %s", monibot.Version)
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
	const minSampleInterval = 1 * time.Minute
	const maxSampleInterval = 24 * time.Hour
	if sampleInterval < minSampleInterval {
		fatal(2, "invalid sampleInterval %s, must be >= %s", sampleInterval, minSampleInterval)
	} else if sampleInterval > maxSampleInterval {
		fatal(2, "invalid sampleInterval %s, must be <= %s", sampleInterval, maxSampleInterval)
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
		internal.PrintWatchdogs(watchdogs)
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
		internal.PrintWatchdogs([]monibot.Watchdog{watchdog})
	case "heartbeat":
		// moni heartbeat <watchdogId>
		watchdogId := flag.Arg(1)
		if watchdogId == "" {
			fatal(2, "empty watchdogId")
		}
		err := api.PostWatchdogHeartbeat(watchdogId)
		if err != nil {
			fatal(1, "%s", err)
		}
	case "machines":
		// moni machines
		machines, err := api.GetMachines()
		if err != nil {
			fatal(1, "%s", err)
		}
		internal.PrintMachines(machines)
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
		internal.PrintMachines([]monibot.Machine{machine})
	case "sample":
		// moni sample <machineId>
		machineId := flag.Arg(1)
		if machineId == "" {
			fatal(2, "empty machineId")
		}
		sampler := internal.NewSampler()
		// we must warm up the sampler first
		_, err := sampler.Sample()
		if err != nil {
			internal.Print("WARNING: cannot sample: %s", err)
		}
		// entering sampling loop
		log.Printf("will send samples in background every %s", sampleInterval)
		for {
			// sleep
			time.Sleep(sampleInterval)
			// sample
			sample, err := sampler.Sample()
			if err != nil {
				internal.Print("WARNING: cannot sample: %s", err)
			}
			err = api.PostMachineSample(machineId, sample)
			if err != nil {
				internal.Print("WARNING: cannot POST sample: %s", err)
			}
		}
	case "metrics":
		// moni metrics
		metrics, err := api.GetMetrics()
		if err != nil {
			fatal(1, "%s", err)
		}
		internal.PrintMetrics(metrics)
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
		internal.PrintMetrics([]monibot.Metric{metric})
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

// ApiLogger logs monibot debug messages
type ApiLogger struct{}

func (l *ApiLogger) Debug(format string, args ...any) {
	log.Printf("[MONIBOT] "+format, args...)
}
