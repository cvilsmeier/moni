package main

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cvilsmeier/monibot-go"
)

func TestSampler(t *testing.T) {
	str := func(s monibot.MachineSample) string {
		return fmt.Sprintf("Tstamp=%v, CpuPercent=%v, MemPercent=%d, DiskPercent=%v, "+
			"Load1=%v, Load5=%v, Load15=%v, "+
			"DiskRead=%v, DiskWrite=%v, NetRecv=%v, NetSend=%v",
			s.Tstamp, s.CpuPercent, s.MemPercent, s.DiskPercent,
			s.Load1, s.Load5, s.Load15,
			s.DiskRead, s.DiskWrite, s.NetRecv, s.NetSend,
		)
	}
	platform := &fakePlatform{}
	sampler := NewSampler(platform)
	// first sample is empty
	sample, err := sampler.Sample()
	assertNil(t, err)
	assertEqual(t, "Tstamp=0, CpuPercent=0, MemPercent=0, DiskPercent=0, Load1=0, Load5=0, Load15=0, DiskRead=0, DiskWrite=0, NetRecv=0, NetSend=0", str(sample))
	// fake machine activity
	platform.unixMilli = time.Date(2025, 1, 4, 10, 0, 0, 0, time.UTC).UnixMilli()
	platform.cpuPercent = 13
	platform.memPercent = 58
	platform.diskPercent = 76
	platform.load = [3]float64{.5, .6, .7}
	platform.diskReadBytes = 6000
	platform.diskWriteBytes = 7000
	platform.netRecvBytes = 4 * 1024
	platform.netSendBytes = 8 * 1024
	sample, err = sampler.Sample()
	assertNil(t, err)
	assertEqual(t, "Tstamp=1735984800000, CpuPercent=13, MemPercent=58, DiskPercent=76, Load1=0.5, Load5=0.6, Load15=0.7, DiskRead=6000, DiskWrite=7000, NetRecv=4096, NetSend=8192", str(sample))
	// sampler must sanitize percent values
	platform.unixMilli = time.Date(2025, 1, 4, 10, 5, 0, 0, time.UTC).UnixMilli()
	platform.cpuPercent = 130
	platform.memPercent = 580
	platform.diskPercent = -760
	platform.load = [3]float64{-.5, -.6, -.7}
	platform.diskReadBytes = 6000
	platform.diskWriteBytes = 7000
	platform.netRecvBytes = 4 * 1024
	platform.netSendBytes = 8 * 1024
	sample, err = sampler.Sample()
	assertNil(t, err)
	assertEqual(t, "Tstamp=1735985100000, CpuPercent=100, MemPercent=100, DiskPercent=0, Load1=-0.5, Load5=-0.6, Load15=-0.7, DiskRead=0, DiskWrite=0, NetRecv=0, NetSend=0", str(sample))
	// must increment disk/net IO values
	platform.unixMilli = time.Date(2025, 1, 4, 10, 10, 0, 0, time.UTC).UnixMilli()
	platform.diskReadBytes = 6000 + 100
	platform.diskWriteBytes = 7000 + 200
	platform.netRecvBytes = 4*1024 + 3
	platform.netSendBytes = 8*1024 + 4
	sample, err = sampler.Sample()
	assertNil(t, err)
	assertEqual(t, "Tstamp=1735985400000, CpuPercent=100, MemPercent=100, DiskPercent=0, Load1=-0.5, Load5=-0.6, Load15=-0.7, DiskRead=100, DiskWrite=200, NetRecv=3, NetSend=4", str(sample))
	// fake error
	platform.unixMilli = time.Date(2025, 1, 4, 10, 15, 0, 0, time.UTC).UnixMilli()
	platform.err = errors.New("machine locked")
	_, err = sampler.Sample()
	assertEqual(t, "machine locked", err.Error())
}

type fakePlatform struct {
	unixMilli      int64
	cpuPercent     float64
	memPercent     float64
	diskPercent    float64
	load           [3]float64
	diskReadBytes  uint64
	diskWriteBytes uint64
	netRecvBytes   uint64
	netSendBytes   uint64
	err            error
}

func (f *fakePlatform) UnixMilli() int64 {
	return f.unixMilli
}

func (f *fakePlatform) CpuPercent() (float64, error) {
	return f.cpuPercent, f.err
}

func (f *fakePlatform) MemPercent() (float64, error) {
	return f.memPercent, f.err
}

func (f *fakePlatform) DiskPercent() (float64, error) {
	return f.diskPercent, f.err
}

func (f *fakePlatform) Load() ([3]float64, error) {
	return f.load, f.err
}

func (f *fakePlatform) DiskBytes() (uint64, uint64, error) {
	return f.diskReadBytes, f.diskWriteBytes, f.err
}

func (f *fakePlatform) NetBytes() (uint64, uint64, error) {
	return f.netRecvBytes, f.netSendBytes, f.err
}

func assertNil(t *testing.T, v any) {
	t.Helper()
	if v != nil {
		t.Fatal("want nil but have", v)
	}
}

func assertEqual(t *testing.T, want, have any) {
	t.Helper()
	if want != have {
		t.Fatal("want ", want, " but have ", have)
	}
}
