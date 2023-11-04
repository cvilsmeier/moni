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
	"log"
	"os"
	"strconv"
	"time"

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

	delayEnvKey     = "MONIBOT_DELAY"
	delayFlag       = "delay"
	defaultDelay    = 5 * time.Second
	defaultDelayStr = "5s"
)

func usage() {
	print("Moni - A command line tool for https://monibot.io")
	print("")
	print("Usage")
	print("")
	print("    moni [flags] command")
	print("")
	print("Flags")
	print("")
	print("    -%s", urlFlag)
	print("        Monibot URL, default is %q.", defaultUrl)
	print("        You can set this also via environment variable %s.", urlEnvKey)
	print("")
	print("    -%s", apiKeyFlag)
	print("        Monibot API Key, default is %q.", defaultApiKey)
	print("        You can set this also via environment variable %s (recommended).", apiKeyEnvKey)
	print("        You can find your API Key in your profile on https://monibot.io.")
	print("        Note: For security, we recommend that you specify the API Key")
	print("        via %s, and not via -%s flag. The flag will show up in", apiKeyEnvKey, apiKeyFlag)
	print("        'ps aux' outputs and can so be eavesdropped.")
	print("")
	print("    -%s", trialsFlag)
	print("        Max. Send trials, default is %v.", defaultTrials)
	print("        You can set this also via environment variable %s.", trialsEnvKey)
	print("")
	print("    -%s", delayFlag)
	print("        Delay between trials, default is %v.", defaultDelayStr)
	print("        You can set this also via environment variable %s.", delayEnvKey)
	print("")
	print("    -%s", verboseFlag)
	print("        Verbose output, default is %v.", defaultVerboseStr)
	print("        You can set this also via environment variable %s ('true' or 'false').", verboseEnvKey)
	print("")
	print("Commands")
	print("")
	print("    ping")
	print("        Ping the Monibot API.")
	print("")
	print("    watchdogs")
	print("        List watchdogs.")
	print("")
	print("    watchdog <watchdogId>")
	print("        Get watchdog by id.")
	print("")
	print("    heartbeat <watchdogId> [interval]")
	print("        Send a heartbeat.")
	print("        If interval is specified, moni will keep sending heartbeats")
	print("        in the background. Min. interval is 1m. If interval is left")
	print("        out, moni will send one heartbeat and then exit.")
	print("")
	print("    machines")
	print("        List machines.")
	print("")
	print("    machine <machineId>")
	print("        Get machine by id.")
	print("")
	print("    sample <machineId> [interval]")
	print("        Send resource usage (load/cpu/mem/disk) samples for machine.")
	print("        Moni consults various files (/proc/loadavg, /proc/cpuinfo)")
	print("        and commands (free, df) to calculate resource usage.")
	print("        It currently supports linux only.")
	print("        If interval is specified, moni will keep sampling in")
	print("        the background. Min. interval is 1m. If interval is left")
	print("        out, moni will send one sample and then exit.")
	print("")
	print("    metrics")
	print("        List metrics.")
	print("")
	print("    metric <metricId>")
	print("        Get and print metric info.")
	print("")
	print("    inc <metricId> <value>")
	print("        Increment a Counter metric.")
	print("        Value must be a non-negative 64-bit integer value.")
	print("")
	print("    set <metricId> <value>")
	print("        Set a Gauge metric.")
	print("        Value must be a non-negative 64-bit integer value.")
	print("")
	print("    config")
	print("        Show config values.")
	print("")
	print("    version")
	print("        Show moni program version.")
	print("")
	print("    sdk-version")
	print("        Show the monibot-go SDK version moni was built with.")
	print("")
	print("    help")
	print("        Show this help page.")
	print("")
	print("Exit Codes")
	print("    0 ok")
	print("    1 error")
	print("    2 wrong user input")
	print("")
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
		print("url      %v", url)
		print("apiKey   %v", apiKey)
		print("trials   %v", trials)
		print("delay    %v", delay)
		print("verbose  %v", verbose)
		os.Exit(0)
	case "version":
		print("moni %s", Version)
		os.Exit(0)
	case "sdk-version":
		print("monibot-go %s", monibot.Version)
		os.Exit(0)
	}
	// validate flags
	if url == "" {
		fatal(2, "empty url")
	}
	if apiKey == "" {
		fatal(2, "empty apiKey")
	}
	if trials < 0 {
		fatal(2, "invalid trials: %d", trials)
	}
	if delay < 0 {
		fatal(2, "invalid delay: %s", delay)
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
			if interval < 1*time.Minute {
				fatal(2, "invalid interval %s: must be >= 1m", interval)
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
					print("WARNING: cannot POST heartbeat: %s", err)
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
		// moni sample <machineId> [interval]
		machineId := flag.Arg(1)
		if machineId == "" {
			fatal(2, "empty machineId")
		}
		var interval time.Duration
		intervalStr := flag.Arg(2)
		if intervalStr != "" {
			interval, err = time.ParseDuration(intervalStr)
			if err != nil {
				fatal(2, "cannot parse interval %q: %s", intervalStr, err)
			}
			if interval < 1*time.Minute {
				fatal(2, "invalid interval %s: must be >= 1m", interval)
			}
		}
		sampler := newSampler()
		if interval == 0 {
			logger.Debug("fetching sample")
			sample, err := sampler.sample()
			if err != nil {
				fatal(1, "cannot sample: %s", err)
			}
			err = api.PostMachineSample(machineId, sample)
			if err != nil {
				fatal(1, "%s", err)
			}
		} else {
			logger.Debug("will send samples in background")
			for {
				// sample
				sample, err := sampler.sample()
				if err != nil {
					print("WARNING: cannot sample: %s", err)
				}
				err = api.PostMachineSample(machineId, sample)
				if err != nil {
					print("WARNING: cannot POST sample: %s", err)
				}
				// sleep
				time.Sleep(interval)
			}
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
	default:
		fatal(2, "unknown command %q, run 'moni help'", command)
	}
}
