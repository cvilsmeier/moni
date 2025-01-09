package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

type Platform struct {
	verbose bool
}

func NewPlatform(verbose bool) *Platform {
	return &Platform{verbose}
}

func (p *Platform) UnixMilli() int64 {
	return time.Now().UnixMilli()
}

func (p *Platform) CpuPercent() (float64, error) {
	percents, err := cpu.Percent(0, false)
	if err != nil {
		return 0, fmt.Errorf("cannot cpu.Percent(): %w", err)
	}
	if len(percents) == 0 {
		return 0, fmt.Errorf("cpu.Percent(): empty result")
	}
	return percents[0], err
}

func (p *Platform) MemPercent() (float64, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("cannot mem.VirtualMemory(): %w", err)
	}
	return vm.UsedPercent, nil
}

func (p *Platform) DiskPercent() (float64, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return 0, fmt.Errorf("cannot disk.Partitions(): %w", err)
	}
	var total uint64
	var used uint64
	for _, partition := range partitions {
		p.debugf("Platform.DiskPercent(): include %s (mnt %s)", partition.Device, partition.Mountpoint)
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			log.Printf("WARNING: cannot disk.Usage(%q): %s", partition.Mountpoint, err)
			continue
		}
		total += usage.Total
		used += usage.Used
	}
	return float64(used) * 100.0 / float64(total), nil
}

func (p *Platform) Load() ([3]float64, error) {
	avg, err := load.Avg()
	if err != nil {
		return [3]float64{0, 0, 0}, fmt.Errorf("cannot load.Avg(): %w", err)
	}
	return [3]float64{avg.Load1, avg.Load5, avg.Load15}, nil
}

func (p *Platform) DiskBytes() (uint64, uint64, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot disk.Partitions(): %w", err)
	}
	var readBytes, writeBytes uint64
	for _, partition := range partitions {
		p.debugf("Platform.DiskBytes(): include %s (mnt %s)", partition.Device, partition.Mountpoint)
		iocMap, err := disk.IOCounters(partition.Device)
		if err != nil {
			log.Printf("WARNING: cannot disk.IOCounters(%q): %s", partition.Device, err)
			continue
		}
		for _, ioc := range iocMap {
			readBytes += ioc.ReadBytes
			writeBytes += ioc.WriteBytes
		}
	}
	return readBytes, writeBytes, nil
}

func (p *Platform) NetBytes() (uint64, uint64, error) {
	iocs, err := net.IOCounters(true)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot net.IOCounters(): %w", err)
	}
	var recvBytes, sendBytes uint64
	for _, ioc := range iocs {
		skip := strings.HasPrefix(ioc.Name, "lo")
		if skip {
			p.debugf("Platform.NetBytes(): skip %q", ioc.Name)
			continue
		}
		p.debugf("Platform.NetBytes(): include %q", ioc.Name)
		recvBytes += ioc.BytesRecv
		sendBytes += ioc.BytesSent
	}
	return recvBytes, sendBytes, nil
}

func (p *Platform) debugf(f string, a ...any) {
	if p.verbose {
		log.Printf("DEBUG: "+f, a...)
	}
}
