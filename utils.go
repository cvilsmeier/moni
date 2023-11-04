package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cvilsmeier/monibot-go"
)

// printWatchdogs prints watchdogs.
func printWatchdogs(watchdogs []monibot.Watchdog) {
	print("%-35s | %-25s | %s", "Id", "Name", "IntervalMillis")
	for _, watchdog := range watchdogs {
		print("%-35s | %-25s | %d", watchdog.Id, watchdog.Name, watchdog.IntervalMillis)
	}
}

// printMachines prints machines.
func printMachines(machines []monibot.Machine) {
	print("%-35s | %s", "Id", "Name")
	for _, machine := range machines {
		print("%-35s | %s", machine.Id, machine.Name)
	}
}

// printMetrics prints metrics.
func printMetrics(metrics []monibot.Metric) {
	print("%-35s | %-25s | %s", "Id", "Name", "Type")
	for _, metric := range metrics {
		print("%-35s | %-25s | %d", metric.Id, metric.Name, metric.Type)
	}
}

// print prints a line to stdout.
func print(f string, a ...any) {
	fmt.Printf(f+"\n", a...)
}

// fatal prints a message to stdout and exits with exitCode.
func fatal(exitCode int, f string, a ...any) {
	fmt.Printf(f+"\n", a...)
	os.Exit(exitCode)
}

// execCommand executes an external binary.
func execCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.WaitDelay = 10 * time.Second
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("cannot run %s: %w", name, err)
	}
	return string(out), err
}

// trimText trims and normalizes a line of text.
func trimText(s string) string {
	s = replace(s, "\t", " ")
	s = replace(s, "\r", "")
	s = replace(s, "\n", "")
	s = replace(s, "  ", " ")
	return strings.TrimSpace(s)
}

// replace replaces strings, even if they occur many times.
func replace(str, old, new string) string {
	var i int
	for strings.Contains(str, old) && i < 100 {
		i++
		str = strings.ReplaceAll(str, old, new)
	}
	return str
}
