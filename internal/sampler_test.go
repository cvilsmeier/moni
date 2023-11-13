package internal

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
		t.Fatalf("wrong %q", have)
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
		t.Fatalf("wrong %v", percent)
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
		t.Fatalf("wrong %v", percent)
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
		t.Fatalf("wrong %v", stat.total)
	}
	if stat.idle != 16649013 {
		t.Fatalf("wrong %v", stat.idle)
	}
}

func TestParseDiskStat(t *testing.T) {
	text := `
259       0 nvme0n1   348631 57325 49778168 51034 237722 390973 34542122 662471 0 262444 729800 0 0 0 0 14038 16295
259       1 nvme0n1p1    187  1000    13454    31      2      0        2      7 0     60     39 0 0 0 0 0 0
259       2 nvme0n1p2 348152 56277 49752186 50957 237639 388315 34512056 662230 0 262220 713187 0 0 0 0 0 0
`
	stat, err := parseDiskStat(text)
	if err != nil {
		t.Fatal(err)
	}
	if stat.read != 49778168 {
		t.Fatalf("wrong %v", stat.read)
	}
	if stat.written != 34542122 {
		t.Fatalf("wrong %v", stat.written)
	}
	// parse multiple devices must add their numbers
	text = `
259       0 nvme0n1   348631 57325 49778168 51034 237722 390973 34542122 662471 0 262444 729800 0 0 0 0 14038 16295
259       1 nvme0n1p1    187  1000    13454    31      2      0        2      7 0     60     39 0 0 0 0 0 0
259       2 nvme0n1p2 348152 56277 49752186 50957 237639 388315 34512056 662230 0 262220 713187 0 0 0 0 0 0
 13       3 sda        48631   7325  778168  1034   7722  90973  4542122  62471 0 62444   29800 0 0 0 0 4038 6295
`
	stat, err = parseDiskStat(text)
	if err != nil {
		t.Fatal(err)
	}
	if stat.read != 49778168+778168 {
		t.Fatalf("wrong %v", stat.read)
	}
	if stat.written != 34542122+4542122 {
		t.Fatalf("wrong %v", stat.written)
	}
	// one more - big time
	text = `
  259       0 nvme0n1 362239 46299 56104016 50812 164382 271075 31063762 427750 0 237596 488351 0 0 0 0 8754 9787
  259       1 nvme0n1p1 187 1000 13454 29 2 0 2 0 0 52 29 0 0 0 0 0 0
  259       2 nvme0n1p2 361764 45248 56078058 50743 164357 270919 31062336 427726 0 237536 478470 0 0 0 0 0 0
  259       3 nvme0n1p3 109 51 5312 17 22 156 1424 23 0 116 41 0 0 0 0 0 0
  259       4 nvme1n1 328 0 18150 55 0 0 0 0 0 56 55 0 0 0 0 0 0
  259       5 nvme1n1p1 58 0 4192 10 0 0 0 0 0 28 10 0 0 0 0 0 0
  259       6 nvme1n1p2 46 0 1168 1 0 0 0 0 0 16 1 0 0 0 0 0 0
  259       7 nvme1n1p3 60 0 4214 7 0 0 0 0 0 20 7 0 0 0 0 0 0
  259       8 nvme1n1p4 58 0 4176 13 0 0 0 0 0 20 13 0 0 0 0 0 0
  7       0 loop0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
  7       1 loop1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
  7       2 loop2 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
  7       3 loop3 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
  7       4 loop4 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
  7       5 loop5 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
  7       6 loop6 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
  7       7 loop7 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
`
	stat, err = parseDiskStat(text)
	if err != nil {
		t.Fatal(err)
	}
	if stat.read != 56104016+18150 {
		t.Fatalf("wrong %v", stat.read)
	}
	if stat.written != 31063762+0 {
		t.Fatalf("wrong %v", stat.written)
	}
	// another one - cpx11
	text = `
	8       0 sda 80084 15703 18492198 19522 5844738 2423893 86294228 1739631 0 4086808 1850825 15827 2 56772622 5738 759399 85932
	8       1 sda1 79680 10018 18464405 19458 5799604 2423893 86294226 1735890 0 4086728 1761086 15823 0 56281328 5737 0 0
	8      14 sda14 55 0 440 6 0 0 0 0 0 36 6 0 0 0 0 0 0
	8      15 sda15 257 5685 23633 46 2 0 2 0 0 140 47 4 2 491294 0 0 0
   11       0 sr0 9 0 3 0 0 0 0 0 0 16 0 0 0 0 0 0 0
`
	stat, err = parseDiskStat(text)
	if err != nil {
		t.Fatal(err)
	}
	if stat.read != 18492198 {
		t.Fatalf("wrong %v", stat.read)
	}
	if stat.written != 86294228 {
		t.Fatalf("wrong %v", stat.written)
	}
	// another one - cpx11
	text = `
	8       0 sda 80084 15703 18492198 19522 5844738 2423893 86294228 1739631 0 4086808 1850825 15827 2 56772622 5738 759399 85932
	8       1 sda1 79680 10018 18464405 19458 5799604 2423893 86294226 1735890 0 4086728 1761086 15823 0 56281328 5737 0 0
	8      14 sda14 55 0 440 6 0 0 0 0 0 36 6 0 0 0 0 0 0
	8      15 sda15 257 5685 23633 46 2 0 2 0 0 140 47 4 2 491294 0 0 0
   11       0 sr0 9 0 3 0 0 0 0 0 0 16 0 0 0 0 0 0 0
`
	stat, err = parseDiskStat(text)
	if err != nil {
		t.Fatal(err)
	}
	if stat.read != 18492198 {
		t.Fatalf("wrong %v", stat.read)
	}
	if stat.written != 86294228 {
		t.Fatalf("wrong %v", stat.written)
	}
	// another one - ex42
	text = `
   8       0 sda 116758444 39021759 19926482661 203182717 26000897 8268143 3426527240 237899876 0 54951320 275172048 0 0 0 0
   8       1 sda1 932947 115896 134102032 1387743 1249 1439 17766 32427 0 327904 1378300 0 0 0 0
   8       2 sda2 15092 1983 2119140 52397 2737 5770 52588 10398 0 10092 54044 0 0 0 0
   8       3 sda3 64849174 34220711 12675170465 101666451 14292719 3994843 150399979 84243006 0 25791576 114454232 0 0 0 0
   8       4 sda4 50960886 4683169 7115075064 100074792 10416893 4266091 3276055891 131993487 0 28484460 164878240 0 0 0 0
   8       5 sda5 126 0 6344 201 126 0 1008 41 0 248 248 0 0 0 0
   8      16 sdb 110318118 11905564 15643223942 1882821133 29859093 37404888 7649927176 643942724 0 119562340 2134899904 0 0 0 0
   8      17 sdb1 866789 180858 134088392 14575763 1222 1466 17766 104537 0 933184 14547152 0 0 0 0
   8      18 sdb2 14262 2290 2100724 136191 2724 5783 52588 27358 0 11588 142428 0 0 0 0
   8      19 sdb3 59827538 6178591 8448306053 885862199 18171943 33110705 4373799915 236072791 0 57767124 963847484 0 0 0 0
   8      20 sdb4 49609178 5543825 7058712365 982246509 10395905 4286934 3276055891 372326516 0 65280276 1201968476 0 0 0 0
   8      21 sdb5 126 0 6344 153 126 0 1008 25 0 168 168 0 0 0 0
   9       0 md0 1286 0 15896 0 2154 0 17232 0 0 0 0 0 0 0 0
   9       2 md2 100419 0 6455106 0 13784159 0 127435352 0 0 0 0 0 0 0 0
   9       3 md3 735950 0 86009266 0 9973230 0 3252823592 0 0 0 0 0 0 0 0
   9       1 md1 625 0 10952 0 8451 0 52532 0 0 0 0 0 0 0 0	
`
	stat, err = parseDiskStat(text)
	if err != nil {
		t.Fatal(err)
	}
	if stat.read != 19926482661+15643223942 {
		t.Fatalf("wrong %v", stat.read)
	}
	if stat.written != 3426527240+7649927176 {
		t.Fatalf("wrong %v", stat.written)
	}
	// parse unknown device must return only zeros
	text = `
  0       0 loop1 348631 57325 49778168 51034 237722 390973 34542122 662471 0 262444 729800 0 0 0 0 14038 16295
  0       1 loop1p1 187 1000 13454 31 2 0 2 7 0 60 39 0 0 0 0 0 0
  0       2 loop1p2 348152 56277 49752186 50957 237639 388315 34512056 662230 0 262220 713187 0 0 0 0 0 0
 14       3 fda        48631 7325 9778168 1034 37722 90973 4542122 62471 0 62444 29800 0 0 0 0 4038 6295
`
	stat, err = parseDiskStat(text)
	if err != nil {
		t.Fatal(err)
	}
	if stat.read != 0 {
		t.Fatalf("wrong %v", stat.read)
	}
	if stat.written != 0 {
		t.Fatalf("wrong %v", stat.written)
	}
}

func TestParsenetStat(t *testing.T) {
	// cat /proc/net/dev
	text := `
Inter-|   Receive                                                      |  Transmit
 face |       bytes packets errs  drop fifo frame compressed multicast |    bytes packets errs drop fifo colls carrier compressed
enp4s0:    21640725   46246    0 13520    0     0          0      1053   13613968   31281    0    0    0     0       0          0
`
	stat, err := parseNetStat(text)
	if err != nil {
		t.Fatal(err)
	}
	if stat.recv != 21640725 {
		t.Fatalf("wrong %v", stat.recv)
	}
	if stat.send != 13613968 {
		t.Fatalf("wrong %v", stat.send)
	}
	// parse multiple devices must add their numbers
	text = `
Inter-|   Receive                                                      |  Transmit
 face |bytes        packets errs  drop fifo frame compressed multicast |    bytes packets errs drop fifo colls carrier compressed
    lo:   117864359   32173    0     0    0     0          0         0  117864359   32173    0    0    0     0       0          0
enp4s0:    21640725   46246    0 13520    0     0          0      1053   13613968   31281    0    0    0     0       0          0
wlp0s20f3:        1       2    3     4    5     6          7         8          9      10   11   12   13    14      15         16
`
	stat, err = parseNetStat(text)
	if err != nil {
		t.Fatal(err)
	}
	if stat.recv != 21640725+1 {
		t.Fatalf("wrong %v", stat.recv)
	}
	if stat.send != 13613968+9 {
		t.Fatalf("wrong %v", stat.send)
	}
	// do not parse loopback device
	text = `
Inter-|   Receive                                                      |  Transmit
 face |bytes        packets errs  drop fifo frame compressed multicast |    bytes packets errs drop fifo colls carrier compressed
    lo:   117864359   32173    0     0    0     0          0         0  117864359   32173    0    0    0     0       0          0
`
	stat, err = parseNetStat(text)
	if err != nil {
		t.Fatal(err)
	}
	if stat.recv != 0 {
		t.Fatalf("wrong %v", stat.recv)
	}
	if stat.send != 0 {
		t.Fatalf("wrong %v", stat.send)
	}
	// ex42
	text = `
Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
enp0s31f6: 26234713008 236646952    0    0    0     0          0         3 758050187292 527772993    0    0    0     0       0          0
    lo: 77517386200 43681074    0    0    0     0          0         0 77517386200 43681074    0    0    0     0       0          0
`
	stat, err = parseNetStat(text)
	if err != nil {
		t.Fatal(err)
	}
	if stat.recv != 26234713008 {
		t.Fatalf("wrong %v", stat.recv)
	}
	if stat.send != 758050187292 {
		t.Fatalf("wrong %v", stat.send)
	}
	// cpx11
	text = `
Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo: 1770506809  849779    0    0    0     0          0         0 1770506809  849779    0    0    0     0       0          0
  eth0: 1056266878 3444386    0    0    0     0          0         0 2143361438 3516390    0    0    0     0       0          0
`
	stat, err = parseNetStat(text)
	if err != nil {
		t.Fatal(err)
	}
	if stat.recv != 1056266878 {
		t.Fatalf("wrong %v", stat.recv)
	}
	if stat.send != 2143361438 {
		t.Fatalf("wrong %v", stat.send)
	}
	// many NICs
	text = `
Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo: 1770506809  849779    0    0    0     0          0         0 1770506809  849779    0    0    0     0       0          0
  eth0: 1056266878 3444386    0    0    0     0          0         0 2143361438 3516390    0    0    0     0       0          0
  eth1:   56266878 3444386    0    0    0     0          0         0   43361438 3516390    0    0    0     0       0          0
  wlp0:     266878 3444386    0    0    0     0          0         0     361438 3516390    0    0    0     0       0          0
`
	stat, err = parseNetStat(text)
	if err != nil {
		t.Fatal(err)
	}
	if stat.recv != 1056266878+56266878+266878 {
		t.Fatalf("wrong %v", stat.recv)
	}
	if stat.send != 2143361438+43361438+361438 {
		t.Fatalf("wrong %v", stat.send)
	}
}
