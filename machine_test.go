package main

import (
	"fmt"
	"testing"
)

func TestParseLoadavg(t *testing.T) {
	loadAvg, err := parseLoadAvg("2.01 0.56 0.15 1/1006 176235\n")
	if err != nil {
		t.Fatal(err)
	}
	have := fmt.Sprintf("%v", loadAvg)
	if have != "[2.01 0.56 0.15]" {
		t.Fatalf("have %q", have)
	}
}

func TestParseMemPercent(t *testing.T) {
	text := `
               total        used        free      shared  buff/cache   available
Mem:        16072456     2864000      301288      433084    13681804    13208456
Swap:        1000444      161024      839420
`
	percent, err := parseMemPercent(text)
	if err != nil {
		t.Fatal(err)
	}
	if percent != 18 {
		t.Fatalf("have %v", percent)
	}
}

func TestParseDiskPercent(t *testing.T) {
	text := `
Filesystem     1K-blocks      Used
udev             7995232         0
/dev/nvme0n1p2 981876212 235000596
/dev/nvme0n1p1    523248      5976
total          990394692 235006572
`
	percent, err := parseDiskPercent(text)
	if err != nil {
		t.Fatal(err)
	}
	if percent != 24 {
		t.Fatalf("have %v", percent)
	}
}

func TestParseCpuPercent(t *testing.T) {
	// cpu  611762 30 136480 16065151 13896 0 5946 0 0 0
	// cpu0 75636 5 17226 2003361 1647 0 2358 0 0 0
	// cpu1 77105 6 16617 2009808 1793 0 689 0 0 0
	text := `
cpu  634755 30 142645 16649013 14328 0 6168 0 0 0
cpu0 78454 5 17986 2076297 1702 0 2432 0 0 0
cpu1 79965 6 17364 2082887 1852 0 722 0 0 0
`
	stat, err := parseCpuStat(text)
	if err != nil {
		t.Fatal(err)
	}
	if stat.total != 17446939 {
		t.Fatalf("have %v", stat.total)
	}
	if stat.idle != 16649013 {
		t.Fatalf("have %v", stat.idle)
	}
}
