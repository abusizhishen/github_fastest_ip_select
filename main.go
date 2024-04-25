package main

import (
	"encoding/json"
	"fmt"
	"github.com/tatsushid/go-fastping"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/user"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	checkUserPermission()
	ipList := getIpList()
	if len(ipList) == 0 {
		panic("no ip found")
	}
	var wg = &sync.WaitGroup{}

	wg.Add(len(ipList))
	var result []IpRtt
	var lock = &sync.Mutex{}
	for _, ip := range ipList {
		go func(ip string, wg *sync.WaitGroup) {
			defer wg.Done()
			var subResult = ping(ip)
			if subResult != nil {
				lock.Lock()
				defer lock.Unlock()
				result = append(result, *subResult)
			}
		}(ip, wg)
	}

	wg.Wait()

	if len(result) == 0 {
		panic("no ip available")
	}
	var ipRtts = IpRtts(result)

	fmt.Println("sort by rtt:")
	sort.Sort(ipRtts)
	minRtt := ipRtts[0]
	fmt.Printf("fastest ip: %s, rtt: %v\n", minRtt.Ip, minRtt.Rtt)
}

func ping(ip string) *IpRtt {
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", ip)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	p.AddIPAddr(ra)

	var ipRtt = &IpRtt{
		Ip: ip,
	}
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
		ipRtt.Rtt = rtt
	}

	err = p.Run()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if ipRtt.Rtt == 0 {
		return nil
	}

	return ipRtt
}

func getIpList() []string {
	// get ips from https://api.github.com/meta
	resp, err := http.Get("https://api.github.com/meta")
	if err != nil {
		panic("access https://api.github.com/meta error: " + err.Error())
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("read response body error: " + err.Error())
	}

	var meta Meta
	err = json.Unmarshal(body, &meta)
	if err != nil {
		panic("unmarshal response body error: " + err.Error())
	}

	if meta.Message != "" {
		panic("get meta error: " + meta.Message)
	}

	ipList := []string{}
	for _, ip := range meta.Web {
		ipList = append(ipList, strings.Split(ip, "/")[0])
	}

	return ipList
}

func checkUserPermission() {
	// 获取当前用户
	currentUser, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get current user: %s\n", err)
		os.Exit(1)
	}

	// 检查 UID 是否为 0
	if currentUser.Uid != "0" {
		// 如果不是 root 用户, 提示使用 sudo 运行
		fmt.Println("This program requires root privileges. Please run it with 'sudo'.")
		os.Exit(1)
	}
}
