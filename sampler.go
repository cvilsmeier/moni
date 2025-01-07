package main

import (
	"math"

	"github.com/cvilsmeier/monibot-go"
)

type samplerPlatform interface {
	UnixMilli() int64
	CpuPercent() (float64, error)
	MemPercent() (float64, error)
	DiskPercent() (float64, error)
	Load() ([3]float64, error)
	DiskBytes() (uint64, uint64, error)
	NetBytes() (uint64, uint64, error)
}

type Sampler struct {
	platform             samplerPlatform
	lastDiskReadSectors  uint64
	lastDiskWriteSectors uint64
	lastNetRecvBytes     uint64
	lastNetSendBytes     uint64
}

func NewSampler(platform samplerPlatform) *Sampler {
	return &Sampler{platform: platform}
}

// Sample calculates a MachineSample for the current resource usage.
func (s *Sampler) Sample() (monibot.MachineSample, error) {
	// CpuPercent
	var cpuPercent int
	{
		percent, err := s.platform.CpuPercent()
		if err != nil {
			return monibot.MachineSample{}, err
		}
		cpuPercent = toPercent(percent)
	}
	// MemPercent
	var memPercent int
	{
		percent, err := s.platform.MemPercent()
		if err != nil {
			return monibot.MachineSample{}, err
		}
		memPercent = toPercent(percent)
	}
	// DiskPercent
	var diskPercent int
	{
		percent, err := s.platform.DiskPercent()
		if err != nil {
			return monibot.MachineSample{}, err
		}
		diskPercent = toPercent(percent)
	}
	// Load1/5/15
	var load1, load5, load15 float64
	{
		avg, err := s.platform.Load()
		if err != nil {
			return monibot.MachineSample{}, err
		}
		load1 = avg[0]
		load5 = avg[1]
		load15 = avg[2]
	}
	// Disk IO sectors
	var diskReadSectors, diskWriteSectors int64
	{
		readBytes, writeBytes, err := s.platform.DiskBytes()
		if err != nil {
			return monibot.MachineSample{}, err
		}
		readSectors := readBytes / 512
		if readSectors > s.lastDiskReadSectors {
			diskReadSectors = int64(readSectors - s.lastDiskReadSectors)
			s.lastDiskReadSectors = readSectors
		}
		writeSectors := writeBytes / 512
		if writeSectors > s.lastDiskWriteSectors {
			diskWriteSectors = int64(writeSectors - s.lastDiskWriteSectors)
			s.lastDiskWriteSectors = writeSectors
		}
	}
	// Net IO
	var netRecvBytes, netSendBytes int64
	{
		recvBytes, sendBytes, err := s.platform.NetBytes()
		if err != nil {
			return monibot.MachineSample{}, err
		}
		if recvBytes > s.lastNetRecvBytes {
			netRecvBytes = int64(recvBytes - s.lastNetRecvBytes)
			s.lastNetRecvBytes = recvBytes
		}
		if sendBytes > s.lastNetSendBytes {
			netSendBytes = int64(sendBytes - s.lastNetSendBytes)
			s.lastNetSendBytes = sendBytes
		}
	}
	// sample done
	return monibot.MachineSample{
		Tstamp:      s.platform.UnixMilli(),
		Load1:       load1,
		Load5:       load5,
		Load15:      load15,
		CpuPercent:  cpuPercent,
		MemPercent:  memPercent,
		DiskPercent: diskPercent,
		DiskReads:   diskReadSectors,
		DiskWrites:  diskWriteSectors,
		NetRecv:     netRecvBytes,
		NetSend:     netSendBytes,
	}, nil
}

func toPercent(v float64) int {
	p := int(math.Round(v))
	if p < 0 {
		return 0
	}
	if p > 100 {
		return 100
	}
	return p
}
