package main

import (
	"log"
	"os"
	"time"

	"github.com/miekg/dns"
)

var (
	localc *dns.Client
	conf   *dns.ClientConfig
)

func check_server(q string, server string) error {
	log.Printf("  Looking up %s using server %s", q, server)

	m := new(dns.Msg)
	m.SetQuestion(q, dns.TypeA)
	m.RecursionDesired = true
	response, _, err := localc.Exchange(m, server)
	if err != nil {
		log.Printf("  Error looking up %s against %s: %s", q, server, err)
		return err
	}
	for _, r := range response.Answer {
		log.Printf("  Got response %s", r.(*dns.A).A)
	}
	return nil
}
func check_server_ns(q, server string) error {
	log.Printf("  Looking up NS record for %s using server %s", q, server)

	m := new(dns.Msg)
	m.SetQuestion(q, dns.TypeNS)
	m.RecursionDesired = true
	response, _, err := localc.Exchange(m, server)
	if err != nil {
		log.Printf("  Error looking up %s against %s: %s", q, server, err)
		return err
	}
	for _, r := range response.Answer {
		ns := r.(*dns.NS).Ns
		log.Printf("    Got response %q", ns)
		check_server_ns_resolve(q, server, ns)
	}
	return nil
}
func check_server_ns_resolve(q, server, ns string) error {
	log.Printf("      Looking up A record for NS %s using server %s", ns, server)
	m := new(dns.Msg)
	m.SetQuestion(ns, dns.TypeA)
	m.RecursionDesired = true
	response, _, err := localc.Exchange(m, server)
	if err != nil {
		log.Printf("      Error looking up %s against %s: %s", ns, server, err)
		return err
	}
	for _, r := range response.Answer {
		ns := r.(*dns.A).A
		log.Printf("      Got response %s", ns)
		check_server_ns_resolve_a(q, ns.String()+":"+conf.Port)
	}
	return nil
}

func check_server_ns_resolve_a(q, ns string) error {
	log.Printf("        Looking up A record for %s using server %s", q, ns)
	m := new(dns.Msg)
	m.SetQuestion(q, dns.TypeA)
	m.RecursionDesired = false
	response, _, err := localc.Exchange(m, ns)
	if err != nil {
		log.Printf("        Error looking up %s against %s: %s", q, ns, err)
		return err
	}
	for _, r := range response.Answer {
		rec := r.(*dns.A).A
		log.Printf("      Got response %s", rec)
	}
	return nil
}

func check(q string) {
	log.Printf("Checking %s", q)
	for _, server := range conf.Servers {
		check_server(q, server+":"+conf.Port)
		check_server_ns(q, server+":"+conf.Port)
	}
}

func main() {
	var err error
	conf, err = dns.ClientConfigFromFile("/etc/resolv.conf")
	if conf == nil {
		log.Fatal("Cannot initialize the local resolver: %s\n", err)
	}
	localc = new(dns.Client)
	localc.ReadTimeout = 5 * time.Second
	check(os.Args[1])
}
