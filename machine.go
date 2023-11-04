package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cvilsmeier/monibot-go"
)

type sampler struct {
	lastStat cpuStat
}

func newSampler() *sampler {
	return &sampler{}
}

// sample calculates a MachineSample for the current resource usage.
func (s *sampler) sample() (monibot.MachineSample, error) {
	var sample monibot.MachineSample
	// load loadavg
	loadAvg, err := loadLoadAvg()
	if err != nil {
		return sample, fmt.Errorf("cannot loadLoadAvg: %w", err)
	}
	sample.Load1 = loadAvg[0]
	sample.Load5 = loadAvg[1]
	sample.Load15 = loadAvg[2]
	// cpu
	cpuPercent, stat, err := loadCpuPercent(s.lastStat)
	if err != nil {
		return sample, fmt.Errorf("cannot loadCpuPercent: %w", err)
	}
	s.lastStat = stat // remember stat for next time
	sample.CpuPercent = cpuPercent
	// mem
	memPercent, err := loadMemPercent()
	if err != nil {
		return sample, fmt.Errorf("cannot loadMemPercent: %w", err)
	}
	sample.MemPercent = memPercent
	// disk
	diskPercent, err := loadDiskPercent()
	if err != nil {
		return sample, fmt.Errorf("cannot loadDiskPercent: %w", err)
	}
	sample.DiskPercent = diskPercent
	// tstamp
	sample.Tstamp = time.Now().UnixMilli()
	return sample, nil
}

// loadLoadAvg loads loadavg from /proc/loadavg
func loadLoadAvg() ([3]float64, error) {
	filename := "/proc/loadavg"
	data, err := os.ReadFile(filename)
	if err != nil {
		return [3]float64{}, fmt.Errorf("cannot read %s: %w", filename, err)
	}
	loadavg, err := parseLoadAvg(string(data))
	if err != nil {
		return [3]float64{}, fmt.Errorf("cannot parseLoadAvg %w", err)
	}
	return loadavg, nil
}

func parseLoadAvg(text string) ([3]float64, error) {
	// cv@cv:~$ cat /proc/loadavg
	// 0.54 0.56 0.55 1/1006 176235
	loadavg := [3]float64{0, 0, 0}
	toks := strings.Split(text, " ")
	if len(toks) < 3 {
		return loadavg, fmt.Errorf("len(toks) < 3 in %q", text)
	}
	for i := 0; i < 3; i++ {
		load, err := strconv.ParseFloat(toks[i], 64)
		if err != nil {
			return loadavg, fmt.Errorf("toks[%d]=%q: cannot ParseFloat: %w", i, toks[i], err)
		}
		loadavg[i] = load
	}
	return loadavg, nil
}

// loadMemPercent uses /usr/bin/free to load mem usage percent.
func loadMemPercent() (int, error) {
	text, err := execCommand("/usr/bin/free", "-k")
	if err != nil {
		return 0, fmt.Errorf("cannot execCommand: %w", err)
	}
	memPercent, err := parseMemPercent(text)
	if err != nil {
		return 0, fmt.Errorf("cannot parseMemPercent: %w", err)
	}
	return memPercent, nil
}

func parseMemPercent(text string) (int, error) {
	//                total        used        free      shared  buff/cache   available
	// Mem:        16072456     2864000      301288      433084    13681804    13208456
	// Swap:        1000444      161024      839420
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = trimText(line)
		after, found := strings.CutPrefix(line, "Mem: ")
		if found {
			toks := strings.Split(after, " ")
			if len(toks) < 3 {
				return 0, fmt.Errorf("want min 3 tokens in %q but was %d", line, len(toks))
			}
			totalStr := toks[0]
			total, err := strconv.ParseInt(totalStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("cannot parse totalStr %q in line %q: %s", totalStr, line, err)
			}
			if total <= 0 {
				return 0, fmt.Errorf("invalid total <= 0 in line %q", line)
			}
			usedStr := toks[1]
			used, err := strconv.ParseInt(usedStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("cannot parse usedStr %q in line %q: %s", usedStr, line, err)
			}
			if used <= 0 {
				return 0, fmt.Errorf("invalid used <= 0 in line %q", line)
			}
			if used > total {
				return 0, fmt.Errorf("invalid used > total in line %q", line)
			}
			return percentOf(float64(used), float64(total)), nil
		}
	}
	return 0, fmt.Errorf("prefix \"Mem: \" not found")
}

// loadDiskPercent uses /usr/bin/df to load disk usage percent.
func loadDiskPercent() (int, error) {
	// /usr/bin/df --exclude-type=tmpfs --total --output=source,size,used
	text, err := execCommand("/usr/bin/df", "--exclude-type=tmpfs", "--total", "--output=source,size,used")
	if err != nil {
		return 0, fmt.Errorf("cannot execCommand: %w", err)
	}
	percent, err := parseDiskPercent(text)
	if err != nil {
		return 0, fmt.Errorf("cannot parseDiskPercent: %w", err)
	}
	return percent, nil
}

func parseDiskPercent(text string) (int, error) {
	// Filesystem     1K-blocks      Used
	// udev             7995232         0
	// /dev/nvme0n1p2 981876212 235000596
	// /dev/nvme0n1p1    523248      5976
	// total          990394692 235006572
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = trimText(line)
		after, found := strings.CutPrefix(line, "total ")
		if found {
			toks := strings.Split(after, " ")
			if len(toks) < 2 {
				return 0, fmt.Errorf("want 2 toks in %q but has only %d", line, len(toks))
			}
			totalStr := toks[0]
			total, err := strconv.ParseInt(totalStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("parse totalStr %q from %q: %w", totalStr, line, err)
			}
			if total <= 0 {
				return 0, fmt.Errorf("invalid total %d from %q", total, line)
			}
			usedStr := toks[1]
			used, err := strconv.ParseInt(usedStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("parse usedStr %q from %q: %w", usedStr, line, err)
			}
			if used <= 0 {
				return 0, fmt.Errorf("invalid used %d from %q", used, line)
			}
			if used > total {
				return 0, fmt.Errorf("invalid used %d > total %d from %q", used, total, line)
			}
			return percentOf(float64(used), float64(total)), nil
		}
	}
	return 0, fmt.Errorf("prefix \"total \" not found")
}

// loadCpuPercent loads current cpu usage percent.
// It reads /proc/stat, waits a bit, then reads /proc/stat again.
// Then it calculates CPU usage percent between first and second read.
func loadCpuPercent(lastStat cpuStat) (int, cpuStat, error) {
	if lastStat.isZero() {
		// load /proc/stat
		stat, err := loadCpuStat()
		if err != nil {
			return 0, stat, fmt.Errorf("cannot loadCpuStat: %w", err)
		}
		lastStat = stat
		// wait a bit
		time.Sleep(5 * time.Second)
	}
	// load /proc/stat
	stat, err := loadCpuStat()
	if err != nil {
		return 0, stat, fmt.Errorf("cannot loadCpuStat: %w", err)
	}
	// calc percent
	total := stat.total - lastStat.total
	idle := stat.idle - lastStat.idle
	used := total - idle
	percent := percentOf(float64(used), float64(total))
	return percent, stat, nil
}

func loadCpuStat() (cpuStat, error) {
	// parse /proc/stat
	filename := "/proc/stat"
	data, err := os.ReadFile(filename)
	if err != nil {
		return cpuStat{}, fmt.Errorf("cannot read %s: %w", filename, err)
	}
	stat, err := parseCpuStat(string(data))
	if err != nil {
		return cpuStat{}, fmt.Errorf("cannot parseCpuStat: %w", err)
	}
	return stat, nil
}

func parseCpuStat(text string) (cpuStat, error) {
	// cpu  611762 30 136480 16065151 13896 0 5946 0 0 0
	// cpu0 75636 5 17226 2003361 1647 0 2358 0 0 0
	// cpu1 77105 6 16617 2009808 1793 0 689 0 0 0
	// ...
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = trimText(line)
		after, found := strings.CutPrefix(line, "cpu ")
		if found {
			toks := strings.Split(after, " ")
			if len(toks) < 5 {
				return cpuStat{}, fmt.Errorf("invalid len(toks) < 5 in %q", line)
			}
			var total int64
			var idle int64
			for i := range toks {
				n, err := strconv.ParseInt(toks[i], 10, 64)
				if err != nil {
					return cpuStat{}, fmt.Errorf("cannot parse toks[%d] %q from line %q: %w", i, toks[i], line, err)
				}
				if i == 3 {
					idle = n
				}
				total += n
			}
			return cpuStat{total, idle}, nil
		}
	}
	return cpuStat{}, fmt.Errorf("prefix \"cpu \" not found")
}

// cpuStat holds data for a cpu usage stat from /proc/stat.
type cpuStat struct {
	total int64
	idle  int64
}

func (s cpuStat) isZero() bool {
	return s.total == 0 && s.idle == 0
}

// percentOf calculates percentage of used compared to total.
// The result is always in the closed interval [0;100].
func percentOf(used, total float64) int {
	percent := int(math.Round(float64(used) * 100.0 / float64(total)))
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	return int(percent)
}
