package main

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

type Platform struct{}

func (p Platform) UnixMilli() int64 {
	return time.Now().UnixMilli()
}

func (p Platform) CpuPercent() (float64, error) {
	percents, err := cpu.Percent(0, false)
	if err != nil {
		return 0, fmt.Errorf("cannot cpu.Percent(): %w", err)
	}
	if len(percents) == 0 {
		return 0, fmt.Errorf("cpu.Percent(): empty result")
	}
	return percents[0], err
}

func (p Platform) MemPercent() (float64, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("cannot mem.VirtualMemory(): %w", err)
	}
	return vm.UsedPercent, nil
}

func (p Platform) DiskPercent() (float64, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return 0, fmt.Errorf("cannot disk.Partitions(): %w", err)
	}
	var total uint64
	var used uint64
	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			return 0, fmt.Errorf("cannot disk.Usage(%q): %w", p.Mountpoint, err)
		}
		total += usage.Total
		used += usage.Used
	}
	return float64(used) * 100.0 / float64(total), nil
}

func (p Platform) Load() ([3]float64, error) {
	avg, err := load.Avg()
	if err != nil {
		return [3]float64{0, 0, 0}, fmt.Errorf("cannot load.Avg(): %w", err)
	}
	return [3]float64{avg.Load1, avg.Load5, avg.Load15}, nil
}

func (p Platform) DiskBytes() (uint64, uint64, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot disk.Partitions(): %w", err)
	}
	var readBytes, writeBytes uint64
	for _, p := range parts {
		iocMap, err := disk.IOCounters(p.Device)
		if err != nil {
			return 0, 0, fmt.Errorf("cannot disk.IOCounters(%q): %w", p.Device, err)
		}
		for _, ioc := range iocMap {
			readBytes += ioc.ReadBytes
			writeBytes += ioc.WriteBytes
		}
	}
	return readBytes, writeBytes, nil
}

func (p Platform) NetBytes() (uint64, uint64, error) {
	iocs, err := net.IOCounters(false)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot net.IOCounters(): %w", err)
	}
	var recvBytes, sendBytes uint64
	for _, ioc := range iocs {
		recvBytes += ioc.BytesRecv
		sendBytes += ioc.BytesSent
	}
	return recvBytes, sendBytes, nil
}
