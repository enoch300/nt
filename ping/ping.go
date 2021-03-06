package ping

import (
	"bytes"
	"fmt"
	"github.com/enoch300/nt/common"
	"github.com/enoch300/nt/icmp"
	"time"
)

// Ping 输入参数包括 目的地址 数据包数量 超时时间 发包间隔
func Ping(srcaddr string, destAddr string, count, timeout, interval int) (result string, pingReturn PingReturn, err error) {
	pingOptions := &PingOptions{}
	pingOptions.SetCount(count)
	pingOptions.SetDestAddr(destAddr)
	pingOptions.SetSrcAddr(srcaddr)
	pingOptions.SetTimeoutMs(timeout)
	pingOptions.SetIntervalMs(interval)

	// 针对域名进行解析
	destAddrs, err := common.DestAddrs(pingOptions.DestAddr())
	if err != nil || len(destAddrs) == 0 {
		return
	}

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Start %v, FROM %v -> %v (%v)\n", time.Now().Format("2006-01-02 15:04:05"), pingOptions.SrcAddr(), pingOptions.DestAddr(), destAddrs[0]))
	begin := time.Now().UnixNano() / 1e6
	pingReturn = runPing(pingOptions.SrcAddr(), destAddrs[0], pingOptions)
	end := time.Now().UnixNano() / 1e6
	buffer.WriteString(fmt.Sprintf("%v packets transmitted, %v packet loss, time %vms\n", count, pingReturn.DropRate, end-begin))
	buffer.WriteString(fmt.Sprintf("rtt min/avg/max = %v/%v/%v ms\n", common.Time2Float(pingReturn.WrstTime), common.Time2Float(pingReturn.AvgTime), common.Time2Float(pingReturn.BestTime)))

	result = buffer.String()
	return
}

func runPing(srcAddr string, destAddr string, option *PingOptions) (pingReturn PingReturn) {
	pingReturn = PingReturn{}
	pingReturn.DestAddr = destAddr

	pid := common.Goid()
	timeout := time.Duration(option.TimeoutMs()) * time.Millisecond
	interval := option.IntervalMs()
	ttl := DEFAULT_TTL

	pingResult := PingResult{}

	seq := 0
	for cnt := 0; cnt < option.Count(); cnt++ {
		icmpReturn, err := icmp.Icmp(srcAddr, destAddr, ttl, pid, timeout, seq)
		if err != nil || !icmpReturn.Success || !common.IsEqualIp(destAddr, icmpReturn.Addr) {
			continue
		}

		pingResult.succSum++
		if pingResult.wrstTime == time.Duration(0) || icmpReturn.Elapsed > pingResult.wrstTime {
			pingResult.wrstTime = icmpReturn.Elapsed
		}
		if pingResult.bestTime == time.Duration(0) || icmpReturn.Elapsed < pingResult.bestTime {
			pingResult.bestTime = icmpReturn.Elapsed
		}
		pingResult.allTime += icmpReturn.Elapsed
		pingResult.avgTime = time.Duration((int64)(pingResult.allTime/time.Microsecond)/(int64)(pingResult.succSum)) * time.Microsecond
		pingResult.success = true

		seq++
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}

	//if !pingResult.success {
	//	pingReturn.Success = false
	//	pingReturn.DropRate = 100.0
	//
	//	return
	//}
	//pingReturn.Success = pingResult.success

	pingReturn.DropRate = float64(option.Count()-pingResult.succSum) / float64(option.Count()) * 100
	pingReturn.AvgTime = pingResult.avgTime
	pingReturn.BestTime = pingResult.bestTime
	pingReturn.WrstTime = pingResult.wrstTime

	return
}
