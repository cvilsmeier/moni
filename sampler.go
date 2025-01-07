package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/cvilsmeier/monibot-go"
)

type Sampler struct {
	lastCpuStat  cpuStat
	lastDiskStat diskStat
	lastNetStat  netStat
}

func NewSampler() *Sampler {
	return &Sampler{}
}

// Sample calculates a MachineSample for the current resource usage.
func (s *Sampler) Sample() (monibot.MachineSample, error) {
	var sample monibot.MachineSample
	// load loadavg
	loadAvg, err := s.loadLoadAvg()
	if err != nil {
		return sample, fmt.Errorf("cannot loadLoadAvg: %w", err)
	}
	sample.Load1, sample.Load5, sample.Load15 = loadAvg[0], loadAvg[1], loadAvg[2]
	// load cpu usage percent
	cpuPercent, err := s.loadCpuPercent()
	if err != nil {
		return sample, fmt.Errorf("cannot loadCpuPercent: %w", err)
	}
	sample.CpuPercent = cpuPercent
	// load mem usage percent
	memPercent, err := s.loadMemPercent()
	if err != nil {
		return sample, fmt.Errorf("cannot loadMemPercent: %w", err)
	}
	sample.MemPercent = memPercent
	// load disk usage percent
	diskPercent, err := s.loadDiskPercent()
	if err != nil {
		return sample, fmt.Errorf("cannot loadDiskPercent: %w", err)
	}
	sample.DiskPercent = diskPercent
	// load disk activity
	diskAct, err := s.loadDiskActivity()
	if err != nil {
		return sample, fmt.Errorf("cannot loadDiskActivity: %w", err)
	}
	sample.DiskReads, sample.DiskWrites = diskAct[0], diskAct[1]
	// load net activity
	netAct, err := s.loadNetActivity()
	if err != nil {
		return sample, fmt.Errorf("cannot loadNetActivity: %w", err)
	}
	sample.NetRecv, sample.NetSend = netAct[0], netAct[1]
	// load local tstamp
	sample.Tstamp = time.Now().UnixMilli()
	return sample, nil
}

// loadAvg holds system load avg
//
//	[0] = load1 (1m)
//	[1] = load5 (5m)
//	[2] = load15 (15m)
type loadAvg [3]float64

// loadLoadAvg loads loadavg from /proc/loadavg
func (s *Sampler) loadLoadAvg() (loadAvg, error) {
	filename := "/proc/loadavg"
	data, err := os.ReadFile(filename)
	if err != nil {
		return loadAvg{}, fmt.Errorf("cannot read %s: %w", filename, err)
	}
	loadavg, err := parseLoadAvg(string(data))
	if err != nil {
		return loadAvg{}, fmt.Errorf("cannot parse %s %w", filename, err)
	}
	return loadavg, nil
}

// loadCpuPercent loads current cpu usage percent.
// It reads current /proc/stat and calculates CPU usage
// percent between current and lastStat.
func (s *Sampler) loadCpuPercent() (_cpuPercent int, _err error) {
	// load /proc/stat
	stat, err := loadCpuStat()
	if err != nil {
		return 0, fmt.Errorf("cannot loadCpuStat: %w", err)
	}
	// save stat for next time
	lastStat := s.lastCpuStat
	s.lastCpuStat = stat
	// if we have no lastStat, we return 0%
	if lastStat.isZero() {
		return 0, nil
	}
	// calc cpu percent as stat minus lastStat
	total := stat.total - lastStat.total
	idle := stat.idle - lastStat.idle
	used := total - idle
	percent := percentOf(used, total)
	return percent, nil
}

// loadMemPercent uses /usr/bin/free to load mem usage percent.
func (s *Sampler) loadMemPercent() (int, error) {
	filename := "/usr/bin/free"
	text, err := execCommand(filename)
	if err != nil {
		return 0, fmt.Errorf("cannot exec %s: %w", filename, err)
	}
	memPercent, err := parseMemPercent(text)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %s output: %w", filename, err)
	}
	return memPercent, nil
}

// loadDiskPercent uses /usr/bin/df to load disk usage percent.
func (s *Sampler) loadDiskPercent() (int, error) {
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

// diskActivity hold number of sectors read and written. It's used for sampling disk activity.
//
//	[0]=read
//	[1]=writes
type diskActivity [2]int64

// loadDiskActivity loads diskActivity since last invocation.
func (s *Sampler) loadDiskActivity() (diskActivity, error) {
	// load current disk stat
	stat, err := loadDiskStat()
	if err != nil {
		return diskActivity{}, fmt.Errorf("cannot loadDiskStat: %w", err)
	}
	// save stat for next time
	lastStat := s.lastDiskStat
	s.lastDiskStat = stat
	// if we have no lastStat, we return zero
	if lastStat.isZero() {
		return diskActivity{}, nil
	}
	// calc stat minus lastStat
	reads := stat.read - lastStat.read
	writes := stat.written - lastStat.written
	return diskActivity{reads, writes}, nil
}

// netActivity hold number of bytes received and sent. It's used for sampling network activity.
//
//	[0]=recv
//	[1]=send
type netActivity [2]int64

// loadNetActivity loads netActivity since last invocation.
func (s *Sampler) loadNetActivity() (netActivity, error) {
	// load current net stat
	stat, err := loadNetStat()
	if err != nil {
		return netActivity{}, fmt.Errorf("cannot loadNetStat: %w", err)
	}
	// save stat for next time
	lastStat := s.lastNetStat
	s.lastNetStat = stat
	// if we have no lastStat, we return zero
	if lastStat.isZero() {
		return netActivity{}, nil
	}
	// calc stat minus lastStat
	recv := stat.recv - lastStat.recv
	send := stat.send - lastStat.send
	return netActivity{recv, send}, nil
}

// helper functions

// parseLoadAvg parses /proc/loadavg
func parseLoadAvg(text string) (loadAvg, error) {
	// cv@cv:~$ cat /proc/loadavg
	// 0.54 0.56 0.55 1/1006 176235
	loadavg := loadAvg{0, 0, 0}
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

// parseMemPercent parses /usr/bin/free output
func parseMemPercent(text string) (int, error) {
	//                total        used        free      shared  buff/cache   available
	// Mem:        16072456     2864000      301288      433084    13681804    13208456
	// Swap:        1000444      161024      839420
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = normalize(line)
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
			return percentOf(used, total), nil
		}
	}
	return 0, fmt.Errorf("prefix \"Mem: \" not found")
}

// parseDiskPercent parses /usr/bin/df output
func parseDiskPercent(text string) (int, error) {
	// Filesystem     1K-blocks      Used
	// udev             7995232         0
	// /dev/nvme0n1p2 981876212 235000596
	// /dev/nvme0n1p1    523248      5976
	// total          990394692 235006572
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = normalize(line)
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
			return percentOf(used, total), nil
		}
	}
	return 0, fmt.Errorf("prefix \"total \" not found")
}

// cpuStat holds data for a cpu usage stat from /proc/stat.
type cpuStat struct {
	total int64
	idle  int64
}

func (s cpuStat) isZero() bool {
	return s.total == 0 && s.idle == 0
}

// loadCpuStat reads /proc/stat and parses it.
func loadCpuStat() (cpuStat, error) {
	// parse /proc/stat
	filename := "/proc/stat"
	data, err := os.ReadFile(filename)
	if err != nil {
		return cpuStat{}, fmt.Errorf("cannot read %s: %w", filename, err)
	}
	stat, err := parseCpuStat(string(data))
	if err != nil {
		return cpuStat{}, fmt.Errorf("cannot parse %s: %w", filename, err)
	}
	return stat, nil
}

// parseCpuStat parses /proc/stat content.
func parseCpuStat(text string) (cpuStat, error) {
	// cpu  611762 30 136480 16065151 13896 0 5946 0 0 0
	// cpu0 75636 5 17226 2003361 1647 0 2358 0 0 0
	// cpu1 77105 6 16617 2009808 1793 0 689 0 0 0
	// ...
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = normalize(line)
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

// loadDiskStat reads /proc/diskstats and parses it.
func loadDiskStat() (diskStat, error) {
	filename := "/proc/diskstats"
	data, err := os.ReadFile(filename)
	if err != nil {
		return diskStat{}, fmt.Errorf("cannot read %s: %w", filename, err)
	}
	stat, err := parseDiskStat(string(data))
	if err != nil {
		return diskStat{}, fmt.Errorf("cannot parse %s: %w", filename, err)
	}
	return stat, nil
}

// parseDiskStat parses /proc/diskstats
// See https://www.kernel.org/doc/Documentation/admin-guide/iostats.rst
func parseDiskStat(text string) (diskStat, error) {
	// 259       0 nvme0n1 348631 57325 49778168 51034 237722 390973 34542122 662471 0 262444 729800 0 0 0 0 14038 16295
	// 259       1 nvme0n1p1 187 1000 13454 31 2 0 2 7 0 60 39 0 0 0 0 0 0
	// 259       2 nvme0n1p2 348152 56277 49752186 50957 237639 388315 34512056 662230 0 262220 713187 0 0 0 0 0 0
	//  12       3 sda 348631 57325 49778168 51034 237722 390973 34542122 662471 0 262444 729800 0 0 0 0 14038 16295
	//  12       4 sda1 348631 57325 49778168 51034 237722 390973 34542122 662471 0 262444 729800 0 0 0 0 14038 16295
	// ...
	lines := strings.Split(text, "\n")
	var stat diskStat
	var sampledDevices []string
	for _, line := range lines {
		line = normalize(line)
		toks := strings.Split(line, " ")
		/*
			"259",              [0] major number
			"2",                [1] minor number
			"nvme0n1p2",        [2] device name
			"362480",           [3] reads completed successfully
			"45251",            [4] reads merged
			"56219218",         [5] sectors read <------ want this
			"50895",            [6] time spent reading (ms)
			"169828",           [7] writes completed
			"284438",           [8] writes merged
			"31247016",         [9] sectors written <------ and this
			"434359",          [10] time spent writing (ms)
			"0",               [11] I/Os currently in progress
			"241188",          [12] time spent doing I/Os (ms)
			"485254",          [13] weighted time spent doing I/Os (ms)
		*/
		if len(toks) < 14 {
			continue
		}
		device := normalize(toks[2])
		// skip devices we're not interested in
		goodDevice := strings.HasPrefix(device, "sd") || strings.HasPrefix(device, "nvme")
		if !goodDevice {
			continue
		}
		// skip sub-devices
		var deviceSampledBefore bool
		for _, sampledDevice := range sampledDevices {
			if strings.HasPrefix(device, sampledDevice) {
				deviceSampledBefore = true
			}
		}
		if deviceSampledBefore {
			continue
		}
		// sample this device
		sampledDevices = append(sampledDevices, device)
		tok := toks[5] // [5] sectors read
		read, err := strconv.ParseInt(tok, 10, 64)
		if err != nil {
			return diskStat{}, fmt.Errorf("cannot parse read count %q: %w", tok, err)
		}
		tok = toks[9] // [9] sectors written
		written, err := strconv.ParseInt(tok, 10, 64)
		if err != nil {
			return diskStat{}, fmt.Errorf("cannot parse write count %q: %w", tok, err)
		}
		stat.read += read
		stat.written += written
	}
	return stat, nil
}

// diskStat holds read/write counters from /proc/diskstats.
type diskStat struct {
	read    int64 // number of sectors read since boot // TODO these might overflow
	written int64 // number of sectors written since boot // TODO these might overflow
}

func (s diskStat) isZero() bool {
	return s.read == 0 && s.written == 0
}

// loadNetStat reads /proc/net/dev and parses it.
func loadNetStat() (netStat, error) {
	filename := "/proc/net/dev"
	data, err := os.ReadFile(filename)
	if err != nil {
		return netStat{}, fmt.Errorf("cannot read %s: %w", filename, err)
	}
	stat, err := parseNetStat(string(data))
	if err != nil {
		return netStat{}, fmt.Errorf("cannot parse %s: %w", filename, err)
	}
	return stat, nil
}

// parseNetStat parses /proc/net/dev
func parseNetStat(text string) (netStat, error) {
	// Inter-|   Receive                                                      |  Transmit
	//  face |       bytes packets errs  drop fifo frame compressed multicast |    bytes packets errs drop fifo colls carrier compressed
	//     [0]         [1]     [2]  [3]   [4]  [5]   [6]        [7]       [8]        [9]
	//     lo:   117864359   32173    0     0    0     0          0         0  117864359   32173    0    0    0     0       0          0
	// enp4s0:    21640725   46246    0 13520    0     0          0      1053   13613968   31281    0    0    0     0       0          0
	// wlp0s20f3:        1       2    3     4    5     6          7         8          9      10   11   12   13    14      15         16
	lines := strings.Split(text, "\n")
	var stat netStat
	for _, line := range lines {
		line = normalize(line)
		toks := strings.Split(line, " ")
		if len(toks) < 10 {
			continue
		}
		device := normalize(toks[0])
		// skip non-device lines, e.g. header lines
		if !strings.HasSuffix(device, ":") {
			continue
		}
		// skip devices we are not interested in
		goodDevice := strings.HasPrefix(device, "e") || strings.HasPrefix(device, "w")
		if !goodDevice {
			continue
		}
		tok := toks[1]
		recv, err := strconv.ParseInt(tok, 10, 64)
		if err != nil {
			return netStat{}, fmt.Errorf("cannot parse recv %q: %w", tok, err)
		}
		tok = toks[9]
		send, err := strconv.ParseInt(tok, 10, 64)
		if err != nil {
			return netStat{}, fmt.Errorf("cannot parse send %q: %w", tok, err)
		}
		stat.recv += recv
		stat.send += send
	}
	return stat, nil
}

// netStat holds read/write counters from /proc/net/dev.
type netStat struct {
	recv int64 // number of bytes received since device startup // TODO these might overflow
	send int64 // number of bytes sent since device startup // TODO these might overflow
}

func (s netStat) isZero() bool {
	return s.recv == 0 && s.send == 0
}

// percentOf calculates percentage of used compared to total.
// The result is always in the closed interval [0;100].
func percentOf(used, total int64) int {
	percentf := float64(used) * 100.0 / float64(total)
	percent := int(math.Round(percentf))
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	return int(percent)
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

// normalize trims and normalizes a line of text.
func normalize(s string) string {
	s = replaceAll(s, "\t", " ")
	s = replaceAll(s, "\r", "")
	s = replaceAll(s, "\n", "")
	s = replaceAll(s, "  ", " ")
	return strings.TrimSpace(s)
}

// replaceAll replaces strings, even if they occur many times.
func replaceAll(str, old, new string) string {
	var i int
	for strings.Contains(str, old) && i < 100 {
		i++
		str = strings.ReplaceAll(str, old, new)
	}
	return str
}
