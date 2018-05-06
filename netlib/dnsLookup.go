package netlib

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/miekg/dns"
)

const RETRY = 5

func init() {
	rand.Seed(time.Now().UnixNano())
}

func DnsLookUp(domainName string, dnsList map[string]string) (net.IP, error) {
	client := &dns.Client{Net: "tcp"}
	msg := &dns.Msg{}
	msg.SetQuestion(domainName+".", dns.TypeA)

	for dnsName, dnsAddress := range dnsList {
		for i := 1; i <= RETRY; i++ {
			r, t, err := client.Exchange(msg, dnsAddress)
			if err == nil && len(r.Answer) > 0 {
				log.Printf("使用%s解析域名成功，耗时: %v\n", dnsName, t)
				return r.Answer[rand.Int()%len(r.Answer)].(*dns.A).A, nil
			} else {
				log.Println(err)
			}
		}
		log.Printf("使用%s解析域名失败...", dnsName)
	}
	return nil, errors.New("域名解析失败")
}

func HttpLookup(domain string) (net.IP, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	r, err := client.Get("https://dns-api.org/A/" + domain)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var ip string
	json, err := simplejson.NewFromReader(r.Body)
	if err != nil {
		return nil, err
	}
	arr, err := json.Array()
	if err != nil {
		return nil, err
	}
	for i := range arr {
		ans := json.GetIndex(i)
		ip, err = ans.Get("value").String()
		if err == nil {
			return net.ParseIP(ip), nil
		}
	}

	return nil, errors.New("could not get the IP address")
}

type LookupResult struct {
	Status int `json:"Status"`
	Answer []struct {
		Name string `json:"name"`
		Type int    `json:"type"`
		TTL  int    `json:"TTL"`
		Data string `json:"data"`
	} `json:"Answer"`
}

func CloudflareHTTPLookup(domain string) (net.IP, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	r, err := client.Get("https://cloudflare-dns.com/dns-query?ct=application/dns-json&type=A&name=" + domain)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	res := LookupResult{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	if res.Status == 0 {
		return net.ParseIP(res.Answer[0].Name), nil
	} else {
		return nil, errors.New("could not get the IP address")
	}
}

func GetIP(domainName string, dnsList map[string]string) (ipAddr string) {
	// name resolution by DNS
	address, err := DnsLookUp(domainName, dnsList)
	if err == nil && address != nil {
		ipAddr = address.String()
		log.Println("域名解析成功：", ipAddr)
		return ipAddr
	}

	// online lookup via http
	address, err = CloudflareHTTPLookup(domainName)
	if err == nil && address != nil {
		ipAddr = address.String()
		log.Println("获取IP地址成功：", ipAddr)
		return ipAddr
	}

	// online lookup via http
	address, err = HttpLookup(domainName)
	if err == nil && address != nil {
		ipAddr = address.String()
		log.Println("获取IP地址成功：", ipAddr)
		return ipAddr
	}
	return ""
}
