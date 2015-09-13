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
)

/**
Created on 2015-09-12

@author : MSK
description :
	1. 逐个ping备选服务器。
	2. 使用rasdial 连接最优服务器
*/

func runCommand(command string, args ...string) string {
	cmd := exec.Command(command, args...)
	out := bytes.Buffer{}
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	cmdOutput := out.String()
	enc := mahonia.NewDecoder("gbk")
	cmdOutput = enc.ConvertString(cmdOutput)
	return cmdOutput
}

//func getPingCmd() string {
//	if path, err := exec.LookPath("ping"); err == nil {
//		return path
//	}
//	return "C:/Windows/System32/ping"
//}
var avgRegex = regexp.MustCompile(`平均\ =\ (?P<avg>\d+)ms`)

func getAvgTime(data string) int {
	matched := avgRegex.FindStringSubmatch(data)
	r, _ := strconv.ParseInt(matched[1], 10, 0)
	return int(r)
}

var lostRegex = regexp.MustCompile(`已发送 = \d+，已接收 = \d+，丢失 = (?P<lost>\d+)`)

func getLostNum(data string) int {
	matched := lostRegex.FindStringSubmatch(data)
	r, _ := strconv.ParseInt(matched[1], 10, 0)
	return int(r)
}

func ping(address string) (int, int, error) {
	result := runCommand("ping", address, "-n", "10")
	return getAvgTime(result), getLostNum(result), nil
}

type VPNInfo struct {
	Name        string
	Host        string
	AvgTime     int
	LostPackage int
}

func (this *VPNInfo) string() string {
	return "Name:" + this.Name + " Host:" + this.Host + " AvgTime:" + string(this.AvgTime) + " Lost:" + string(this.LostPackage)
}

func getVpnList() []*VPNInfo {
	return []*VPNInfo{
		&VPNInfo{Name: "Yunti-HK1", Host: "hk1.seehey.com"},
		&VPNInfo{Name: "Yunti-HK2", Host: "hk2.seehey.com"},
		&VPNInfo{Name: "Yunti-TW1", Host: "tw1.seehey.com"},
		&VPNInfo{Name: "Yunti-JP1", Host: "jp1.seehey.com"},
		&VPNInfo{Name: "Yunti-JP3", Host: "jp2.seehey.com"},
		&VPNInfo{Name: "Yunti-JP2", Host: "jp3.seehey.com"},
	}
}

func runVPNPinger(vpninfo *VPNInfo, waitGroup *sync.WaitGroup) {
	go func() {
		defer waitGroup.Done()
		vpninfo.AvgTime, vpninfo.LostPackage, _ = ping(vpninfo.Host)
	}()
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

var token string

func getVPNAuth() (string, string) {
	v := strings.Split(token, "|")
	return v[0], v[1]
}

func dialVPN(vpnInfo *VPNInfo) {
	log.Println(runCommand("rasdial", "/DISCONNECT"))
	username, password := getVPNAuth()
	time.Sleep(1 * time.Second)
	log.Println(runCommand("rasdial", vpnInfo.Name, username, password))
}

func main() {
	flag.Parse()

	vpnList := getVpnList()
	vpnSize := len(vpnList)

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(vpnSize)
	for _, vpninfo := range vpnList {
		runVPNPinger(vpninfo, waitGroup)
	}
	log.Println("Waiting for speed result")
	waitGroup.Wait()
	log.Println("===============================")
	log.Println("speed summary:")
	for _, vpninfo := range vpnList {
		log.Println(vpninfo)
	}
	log.Println("===============================")
	log.Println("dialing vpn:")
	fastVpn := chooseVPN(vpnList)
	log.Println(fastVpn)
	dialVPN(fastVpn)
}
