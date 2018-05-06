package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/weaming/HProxy/netlib"
)

var (
	version = "0.1.2"
	dnsList = map[string]string{
		"Cloudflare": "1.1.1.1:53",
		"Google 8.8": "8.8.8.8:53",
		"Google 4.4": "8.8.4.4:53",
		// "OpenDNS_1":    "208.67.222.222:5353",
		// "OpenDNS_2":    "208.67.220.220:443",
		// "OpenDNS_2-fs": "208.67.220.123:443",
	}

	fixedIP string
	domain  = "www.google.com"
	port    = ":443"
)

func init() {
	flag.StringVar(&domain, "domain", domain, "set your proxy domain")
	flag.StringVar(&port, "port", port, "remote port to proxy")
	flag.StringVar(&fixedIP, "ip", "", "set your domain IP")

	flag.Parse()
	fmt.Printf("Version %v\n", version)
}

func main() {
	ProxyDomain(domain)
	HandlerSignal()
}

func ProxyDomain(domain string) {
	var ipAddr string
	if fixedIP == "" {
		log.Println("正在获取IP地址，请稍候~")
		ipAddr = netlib.GetIP(domain, dnsList)
	} else {
		ipAddr = fixedIP
		log.Println("使用手动指定的IP地址：", ipAddr)
	}

	go netlib.StartServingTCPProxy(port, ipAddr+port)
}

func HandlerSignal() {
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Kill)
	signal.Notify(interrupt, os.Interrupt)

	select {
	case <-interrupt:
		fmt.Println("ByeBye!")
	}
}
