package main

import (
	"net"
	"strings"
	"sync"
)

var (
	iranCidrBlocksOnce sync.Once
	iranCidrNets       []*net.IPNet
)

// List of major Iranian domestic service domains, banking portals, and payment gateways
var iranDomesticDomains = map[string]struct{}{
	// TLD & Second-Level
	".ir": {}, ".xn--mgba3a4f16a": {}, // .ایران
	".co.ir": {}, ".gov.ir": {}, ".ac.ir": {}, ".org.ir": {}, ".net.ir": {}, ".sch.ir": {}, ".id.ir": {},

	// Banking & Payment Systems
	"shaparak.ir": {}, "pep.shaparak.ir": {}, "sep.shaparak.ir": {}, "pec.shaparak.ir": {},
	"bmi.ir": {}, "bankmellat.ir": {}, "bki.ir": {}, "tejaratbank.ir": {}, "sb24.ir": {},
	"parsian-bank.ir": {}, "bsi.ir": {}, "bpi.ir": {}, "rb24.ir": {}, "postbank.ir": {},
	"edbi.ir": {}, "enbank.ir": {}, "sina-bank.ir": {}, "bank-maskan.ir": {}, "izbank.ir": {},
	"karafarinbank.ir": {}, "shahr-bank.ir": {}, "day24.ir": {}, "ansarbank.ir": {},
	"ghavamin.ir": {}, "mebank.ir": {}, "refah-bank.ir": {}, "ttbank.ir": {}, "ba24.ir": {},
	"ayandeh-bank.ir": {}, "middleeastbank.ir": {}, "asanpardakht.ir": {}, "zarinpal.com": {},
	"idpay.ir": {}, "pay.ir": {}, "nextpay.ir": {}, "sep.ir": {}, "sadadpsp.ir": {}, "pec.ir": {},
	"behbardakht.ir": {}, "fanavacard.ir": {}, "sayancard.ir": {},

	// Domestic Portals, Media & Services
	"divar.ir": {}, "torob.com": {}, "digikala.com": {}, "snapp.ir": {}, "tapsi.ir": {},
	"aparat.com": {}, "telewebion.com": {}, "filimo.com": {}, "namava.ir": {}, "rubika.ir": {},
	"bale.ai": {}, "eitaa.com": {}, "gap.im": {}, "splus.ir": {}, "neshan.org": {}, "balad.ir": {},
	"varzesh3.com": {}, "isna.ir": {}, "farsnews.ir": {}, "tasnimnews.com": {}, "irna.ir": {},
	"yjc.ir": {}, "tabnak.ir": {}, "asriran.com": {}, "khabarfoori.com": {}, "emalls.ir": {},
	"nobat.ir": {}, "snappfood.ir": {}, "alibaba.ir": {}, "flytoday.ir": {}, "safarmarket.com": {},
	"cafebazaar.ir": {}, "myket.ir": {}, "sibapp.com": {}, "iauec.ac.ir": {}, "sanjesh.org": {},
	"medu.ir": {}, "tamin.ir": {}, "post.ir": {}, "sabteahval.ir": {},
}

// Major aggregated IP ranges assigned to Iranian Autonomous Systems (ASNs)
var rawIranCidrs = []string{
	"2.144.0.0/14", "5.56.0.0/15", "5.160.0.0/12", "5.200.0.0/13",
	"31.2.0.0/15", "37.98.0.0/15", "37.152.0.0/13", "37.254.0.0/15",
	"46.100.0.0/15", "46.143.0.0/16", "46.209.0.0/16", "46.224.0.0/15",
	"77.36.0.0/15", "77.77.64.0/18", "78.38.0.0/15", "78.157.0.0/16",
	"78.158.0.0/16", "79.127.0.0/16", "80.191.0.0/16", "80.253.144.0/20",
	"85.185.0.0/16", "89.165.0.0/16", "89.196.0.0/16", "91.98.0.0/15",
	"91.240.60.0/22", "94.182.0.0/15", "94.232.168.0/21", "95.38.0.0/16",
	"151.232.0.0/13", "151.240.0.0/14", "164.138.128.0/18", "176.101.0.0/16",
	"178.131.0.0/16", "178.252.128.0/17", "185.0.0.0/12", "188.136.0.0/15",
	"188.210.0.0/15", "188.211.0.0/16", "188.212.0.0/15", "188.245.0.0/16",
	"188.253.0.0/16", "194.225.0.0/16", "217.218.0.0/15", "217.66.192.0/18",
}

func initIranCidrs() {
	for _, cidr := range rawIranCidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil && ipNet != nil {
			iranCidrNets = append(iranCidrNets, ipNet)
		}
	}
}

// IsDomesticIranTarget determines whether target host/IP belongs to Iranian services or banking networks
func IsDomesticIranTarget(target string) bool {
	host := target
	if h, _, err := net.SplitHostPort(target); err == nil {
		host = h
	}
	hostLower := strings.ToLower(strings.TrimSpace(host))

	// 1. Check TLD suffix
	if strings.HasSuffix(hostLower, ".ir") || strings.HasSuffix(hostLower, ".ایران") {
		return true
	}

	// 2. Check GeoSite domestic list
	if _, ok := iranDomesticDomains[hostLower]; ok {
		return true
	}
	for domain := range iranDomesticDomains {
		if strings.HasSuffix(hostLower, "."+domain) {
			return true
		}
	}

	// 3. Check GeoIP CIDRs
	ip := net.ParseIP(hostLower)
	if ip != nil {
		iranCidrBlocksOnce.Do(initIranCidrs)
		for _, ipNet := range iranCidrNets {
			if ipNet.Contains(ip) {
				return true
			}
		}
	}

	return false
}
