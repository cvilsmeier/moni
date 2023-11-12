package internal

import (
	"fmt"

	"github.com/cvilsmeier/monibot-go"
)

// PrintWatchdogs prints watchdogs.
func PrintWatchdogs(watchdogs []monibot.Watchdog) {
	Print("%-35s | %-25s | %s", "Id", "Name", "IntervalMillis")
	for _, watchdog := range watchdogs {
		Print("%-35s | %-25s | %d", watchdog.Id, watchdog.Name, watchdog.IntervalMillis)
	}
}

// PrintMachines prints machines.
func PrintMachines(machines []monibot.Machine) {
	Print("%-35s | %s", "Id", "Name")
	for _, machine := range machines {
		Print("%-35s | %s", machine.Id, machine.Name)
	}
}

// PrintMetrics prints metrics.
func PrintMetrics(metrics []monibot.Metric) {
	Print("%-35s | %-25s | %s", "Id", "Name", "Type")
	for _, metric := range metrics {
		Print("%-35s | %-25s | %d", metric.Id, metric.Name, metric.Type)
	}
}

// Print prints a line to stdout.
func Print(f string, a ...any) {
	fmt.Printf(f+"\n", a...)
}
