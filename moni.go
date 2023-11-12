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
	urlEnvKey  = "MONI_URL"
	urlFlag    = "url"
	defaultUrl = "https://monibot.io"

	apiKeyEnvKey  = "MONI_API_KEY"
	apiKeyFlag    = "apiKey"
	defaultApiKey = ""

	verboseEnvKey     = "MONI_VERBOSE"
	verboseFlag       = "v"
	defaultVerbose    = false
	defaultVerboseStr = "false"

	trialsEnvKey     = "MONI_TRIALS"
	trialsFlag       = "trials"
	defaultTrials    = 12
	defaultTrialsStr = "12"

	delayEnvKey     = "MONI_DELAY"
	delayFlag       = "delay"
	defaultDelay    = 5 * time.Second
	defaultDelayStr = "5s"

	sampleIntervalEnvKey     = "MONI_SAMPLE_INTERVAL"
	sampleIntervalFlag       = "sampleInterval"
	defaultSampleInterval    = 5 * time.Minute
	defaultSampleIntervalStr = "5m"

	disksEnvKey  = "MONI_DISKS"
	disksFlag    = "disks"
	defaultDisks = ""
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
	internal.Print("        Delay between trials, default is %v.", defaultDelayStr)
	internal.Print("        You can set this also via environment variable %s.", delayEnvKey)
	internal.Print("")
	internal.Print("    -%s", sampleIntervalFlag)
	internal.Print("        Machine sample interval, default is %v.", defaultSampleIntervalStr)
	internal.Print("        You can set this also via environment variable %s.", sampleIntervalEnvKey)
	internal.Print("        This flag is only relevant for the 'sample' command.")
	internal.Print("")
	internal.Print("    -%s", disksFlag)
	internal.Print("        Machine sample disk device names, default is %v.", defaultDisks)
	internal.Print("        You can set this also via environment variable %s.", disksEnvKey)
	internal.Print("        This flag is only relevant for the 'sample' command.")
	internal.Print("        It specifies the device name(s) of the disks (or partitions)")
	internal.Print("        that should be included in machine sampling.")
	internal.Print("        If it's empty then no disk stats are sampled.")
	internal.Print("        This flag can be a single disk name, e.g. 'sda' or multiple")
	internal.Print("        comma-separated disk names, e.g. 'sda,sdb' (spaces must be avoided)")
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
	internal.Print("    heartbeat <watchdogId> [interval]")
	internal.Print("        Send a heartbeat.")
	internal.Print("        If interval is specified, moni will keep sending heartbeats")
	internal.Print("        in the background. Min. interval is 1m. If interval is left")
	internal.Print("        out, moni will send one heartbeat and then exit.")
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
	internal.Print("        sample interval, default %s, see flag 'sampleInterval.", defaultSampleIntervalStr)
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
		delayStr = defaultDelayStr
	}
	delay, err := time.ParseDuration(delayStr)
	if err != nil {
		fatal(2, "cannot parse delay %q: %s", delayStr, err)
	}
	flag.DurationVar(&delay, delayFlag, delay, "")
	// -sampleInterval 5m
	sampleIntervalStr := os.Getenv(sampleIntervalEnvKey)
	if sampleIntervalStr == "" {
		sampleIntervalStr = defaultSampleIntervalStr
	}
	sampleInterval, err := time.ParseDuration(sampleIntervalStr)
	if err != nil {
		fatal(2, "cannot parse sampleInterval %q: %s", sampleIntervalStr, err)
	}
	flag.DurationVar(&sampleInterval, sampleIntervalFlag, sampleInterval, "")
	// -disks sda,sdb
	disks := os.Getenv(disksEnvKey)
	if disks == "" {
		disks = defaultDisks
	}
	flag.StringVar(&disks, disksFlag, disks, "")
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
		internal.Print("delay           %v", delay)
		internal.Print("sampleInterval  %v", sampleInterval)
		internal.Print("disks           %v", disks)
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
	// cvvvvvvvvvvv const minSampleInterval = 1 * time.Minute
	const minSampleInterval = 1 * time.Second
	const maxSampleInterval = 24 * time.Hour
	if sampleInterval < minSampleInterval {
		fatal(2, "invalid sampleInterval %s, must be >= %s", sampleInterval, minSampleInterval)
	} else if sampleInterval > maxSampleInterval {
		fatal(2, "invalid sampleInterval %s, must be <= %s", sampleInterval, maxSampleInterval)
	}
	// init monibot Api
	logger := monibot.NewDiscardLogger()
	if verbose {
		logger = monibot.NewLogLogger(log.Default())
	}
	sender := monibot.NewSenderWithOptions(apiKey, monibot.SenderOptions{
		Logger:     logger,
		MonibotUrl: url,
	})
	retrySender := monibot.NewRetrySenderWithOptions(sender, monibot.RetrySenderOptions{
		Logger: logger,
		Trials: trials,
		Delay:  delay,
	})
	api := monibot.NewApiWithSender(retrySender)
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
			const minInterval = 1 * time.Minute
			if interval < minInterval {
				fatal(2, "invalid interval %s: must be >= %s", interval, minInterval)
			}
		}
		if interval == 0 {
			err := api.PostWatchdogHeartbeat(watchdogId)
			if err != nil {
				fatal(1, "%s", err)
			}
		} else {
			logger.Debug("will send heartbeats in background")
			for {
				// send
				err := api.PostWatchdogHeartbeat(watchdogId)
				if err != nil {
					internal.Print("WARNING: cannot POST heartbeat: %s", err)
				}
				// sleep
				time.Sleep(interval)
			}
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
		var diskDevices []string
		if disks != "" {
			diskDevices = strings.Split(disks, ",")
		}
		sampler := internal.NewSampler(diskDevices)
		// we must warm up the sampler first
		_, err := sampler.Sample()
		if err != nil {
			internal.Print("WARNING: cannot sample: %s", err)
		}
		// entering sampling loop
		logger.Debug("will send samples in background every %s", sampleInterval)
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
