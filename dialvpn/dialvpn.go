package main

import (
	"bytes"
	"flag"
	"github.com/axgle/mahonia"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"runtime"
)

/**
Created on 2015-09-12

@author : MSK
description :
	1. 逐个ping备选服务器。
	2. 使用rasdial 连接最优服务器

	go install -ldflags "-X main.token username|password" github.com/mskyspet/gotools/dialvpn
*/
type VPNInfo struct {
	Name        string
	Host        string
	AvgTime     int
	LostPackage int
}

func (this *VPNInfo) string() string {
	return "Name:" + this.Name + " Host:" + this.Host + " AvgTime:" + string(this.AvgTime) + " Lost:" + string(this.LostPackage)
}


type VPNDialer interface {
	PingVpnList([]*VPNInfo)
	Dial(*VPNInfo)
}

func GetVPNDailer() VPNDialer {
	if runtime.GOOS == "windows" {
		return &WindowsDialer{}
	}
	panic("Not support Windows")
}

var lostRegex = regexp.MustCompile(`已发送 = \d+，已接收 = \d+，丢失 = (?P<lost>\d+)`)

func getLostNum(data string) int {
	matched := lostRegex.FindStringSubmatch(data)
	r, _ := strconv.ParseInt(matched[1], 10, 0)
	return int(r)
}

var avgRegex = regexp.MustCompile(`平均\ =\ (?P<avg>\d+)ms`)

func getAvgTime(data string) int {
	matched := avgRegex.FindStringSubmatch(data)
	r, _ := strconv.ParseInt(matched[1], 10, 0)
	return int(r)
}

type WindowsDialer struct {

}

func (this *WindowsDialer) runCommand(command string, args ...string) string {
	cmd := exec.Command(command, args...)
	out := bytes.Buffer{}
	cmd.Stdout = &out
	err := cmd.Run()

	cmdOutput := out.String()
	enc := mahonia.NewDecoder("gbk")
	cmdOutput = enc.ConvertString(cmdOutput)

	if err != nil {
		log.Panic(cmdOutput)
	}
	return cmdOutput
}

func (this *WindowsDialer) ping(address string) (int, int, error) {
	result := this.runCommand("ping", address, "-n", "10")
	return getAvgTime(result), getLostNum(result), nil
}

func (this *WindowsDialer) PingVpnList(vpnList []*VPNInfo) {
	vpnSize := len(vpnList)

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(vpnSize)
	for _, vpninfo := range vpnList {
		go func(vpnInfo *VPNInfo) {
			defer waitGroup.Done()
			vpnInfo.AvgTime, vpnInfo.LostPackage, _ = this.ping(vpnInfo.Host)
		}(vpninfo)
	}

	waitGroup.Wait()
}

var token string

func getVPNAuth() (string, string) {
	v := strings.Split(token, "|")
	return v[0], v[1]
}

func (this *WindowsDialer) Dial(vpnInfo *VPNInfo) {
	log.Println(this.runCommand("rasdial", "/DISCONNECT"))
	username, password := getVPNAuth()
	time.Sleep(1 * time.Second)
	log.Println(this.runCommand("rasdial", vpnInfo.Name, username, password))
}


func getVpnList() []*VPNInfo {
	return []*VPNInfo{
		&VPNInfo{Name: "Yunti-HK1-L2TP", Host: "p2.hk1.seehey.com"},
		&VPNInfo{Name: "Yunti-HK2-L2TP", Host: "p2.hk2.seehey.com"},
		&VPNInfo{Name: "Yunti-TW1-L2TP", Host: "p2.tw1.seehey.com"},
		&VPNInfo{Name: "Yunti-JP1-L2TP", Host: "p2.jp1.seehey.com"},
		&VPNInfo{Name: "Yunti-JP2-L2TP", Host: "p2.jp2.seehey.com"},
		&VPNInfo{Name: "Yunti-JP3-L2TP", Host: "p2.jp3.seehey.com"},
	}
}

func chooseVPN(vpnList []*VPNInfo) *VPNInfo {
	var fastVpn *VPNInfo
	fastVpn = vpnList[0]
	for _, vpnInfo := range vpnList {
		if vpnInfo.LostPackage > fastVpn.LostPackage {
			continue
		}
		if vpnInfo.LostPackage < fastVpn.LostPackage || vpnInfo.AvgTime < fastVpn.AvgTime {
			fastVpn = vpnInfo
		}
	}
	return fastVpn
}

func main() {
	flag.Parse()

	vpnList := getVpnList()

	dialer := GetVPNDailer()

	log.Println("Run for speed analysis")
	dialer.PingVpnList(vpnList)
	log.Println("===============================")
	log.Println("speed summary:")
	for _, vpninfo := range vpnList {
		log.Println(vpninfo)
	}
	log.Println("===============================")
	log.Println("dialing vpn:")
	fastVpn := chooseVPN(vpnList)
	log.Println(fastVpn)
	dialer.Dial(fastVpn)
}
