package mtr

import (
	"bytes"
	"fmt"
	"time"

	"github.com/enoch300/nettool/common"
	"github.com/enoch300/nettool/icmp"
)

// Mtr 执行traceroute操作 新增ipv6操作
func Mtr(ipAddr string, maxHops, sntSize, timeoutMs int) (result string, hops []Hop, err error) {
	options := MtrOptions{}
	options.SetMaxHops(maxHops)
	options.SetSntSize(sntSize)
	options.SetTimeoutMs(timeoutMs)

	var out MtrResult
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Start: %v, DestAddr: %v\n", time.Now().Format("2006-01-02 15:04:05"), ipAddr))
	out, err = runMtr(ipAddr, &options)

	if err == nil {
		if len(out.Hops) == 0 {
			buffer.WriteString("TestMtr failed. Expected at least one hop\n")
			return buffer.String(), hops, nil
		}
	} else {
		buffer.WriteString(fmt.Sprintf("TestMtr failed due to an error: %v\n", err))
		return buffer.String(), hops, err
	}

	buffer.WriteString(fmt.Sprintf("%-3v %-48v  %10v%c  %10v  %10v  %10v  %10v  %10v\n", "", "HOST", "Loss", '%', "Snt", "Last", "Avg", "Best", "Wrst"))

	// 根据原生的linux mtr结果，格式化mtr输出
	var hop_str string
	var last_hop int
	for index, hop := range out.Hops {
		if hop.Success {
			if hop_str != "" {
				buffer.WriteString(hop_str)
				hop_str = ""
			}

			hops = append(hops, Hop{
				RouteNo: hop.TTL,
				Addr:    hop.Address,
				Loss:    hop.Loss,
				Snt:     hop.Snt,
				Last:    common.Time2Float(hop.LastTime),
				Avg:     common.Time2Float(hop.AvgTime),
				Best:    common.Time2Float(hop.BestTime),
				Wrst:    common.Time2Float(hop.WrstTime),
			})

			buffer.WriteString(fmt.Sprintf("%-3d %-48v  %10.1f%c  %10v  %10.2f  %10.2f  %10.2f  %10.2f\n", hop.TTL, hop.Address, hop.Loss, '%', hop.Snt, common.Time2Float(hop.LastTime), common.Time2Float(hop.AvgTime), common.Time2Float(hop.BestTime), common.Time2Float(hop.WrstTime)))
			last_hop = hop.TTL
		} else {
			if index != len(out.Hops)-1 {
				hop_str += fmt.Sprintf("%-3d %-48v  %10.1f%c  %10v  %10.2f  %10.2f  %10.2f  %10.2f\n", hop.TTL, "???", float32(100), '%', int(0), float32(0), float32(0), float32(0), float32(0))
			} else {
				last_hop++
				buffer.WriteString(fmt.Sprintf("%-3d %-48v\n", last_hop, "???"))
			}
			hops = append(hops, Hop{RouteNo: hop.TTL, Addr: hop.Host, Loss: 100})
		}
	}

	return buffer.String(), hops, nil
}

// mtr的实现
func runMtr(destAddr string, options *MtrOptions) (result MtrResult, err error) {
	result.Hops = []common.IcmpHop{}
	result.DestAddress = destAddr

	// 用于避免多协程发起mtr造成的干扰
	pid := common.Goid()
	timeout := time.Duration(options.TimeoutMs()) * time.Millisecond

	mtrResults := make([]*MtrReturn, options.MaxHops()+1)

	unkownHopMax := 6
	// 用于验证数据包
	seq := 0
	for snt := 0; snt < options.SntSize(); snt++ {
		unkownHopCount := 0
		for ttl := 1; ttl < options.MaxHops(); ttl++ {
			if mtrResults[ttl] == nil {
				mtrResults[ttl] = &MtrReturn{TTL: ttl, Host: "???", SuccSum: 0, Success: false, LastTime: time.Duration(0), AllTime: time.Duration(0), BestTime: time.Duration(0), WrstTime: time.Duration(0), AvgTime: time.Duration(0)}
			}

			hopReturn, err := icmp.Icmp(destAddr, ttl, pid, timeout, seq)
			if err != nil || !hopReturn.Success {
				if unkownHopCount == unkownHopMax {
					break
				}

				unkownHopCount++
				continue
			}

			mtrResults[ttl].SuccSum = mtrResults[ttl].SuccSum + 1
			mtrResults[ttl].Host = hopReturn.Addr
			mtrResults[ttl].LastTime = hopReturn.Elapsed
			if mtrResults[ttl].WrstTime == time.Duration(0) || hopReturn.Elapsed > mtrResults[ttl].WrstTime {
				mtrResults[ttl].WrstTime = hopReturn.Elapsed
			}
			if mtrResults[ttl].BestTime == time.Duration(0) || hopReturn.Elapsed < mtrResults[ttl].BestTime {
				mtrResults[ttl].BestTime = hopReturn.Elapsed
			}
			mtrResults[ttl].AllTime += hopReturn.Elapsed
			mtrResults[ttl].AvgTime = time.Duration((int64)(mtrResults[ttl].AllTime/time.Microsecond)/(int64)(mtrResults[ttl].SuccSum)) * time.Microsecond
			mtrResults[ttl].Success = true

			if common.IsEqualIp(hopReturn.Addr, destAddr) {
				break
			}
		}
	}

	for index, mtrResult := range mtrResults {
		if index == 0 {
			continue
		}

		if mtrResult == nil {
			break
		}

		hop := common.IcmpHop{TTL: mtrResult.TTL, Snt: options.SntSize()}
		hop.Address = mtrResult.Host
		hop.Host = mtrResult.Host
		hop.AvgTime = mtrResult.AvgTime
		hop.BestTime = mtrResult.BestTime
		hop.LastTime = mtrResult.LastTime
		failSum := options.SntSize() - mtrResult.SuccSum
		loss := (float32)(failSum) / (float32)(options.SntSize()) * 100
		hop.Loss = float32(loss)
		hop.WrstTime = mtrResult.WrstTime
		hop.Success = mtrResult.Success

		result.Hops = append(result.Hops, hop)

		// 主要用于避免ipv6省略的情况
		if common.IsEqualIp(hop.Host, destAddr) {
			break
		}
	}

	return result, nil
}
