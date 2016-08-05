package main

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
)

var (
	localc             *dns.Client
	conf               *dns.ClientConfig
	successful_queries map[string]bool
)

func init() {
	successful_queries = make(map[string]bool)
}

func check_server(q string, server string) error {
	log.Printf("  Looking up %s using local server %s", q, server)

	m := new(dns.Msg)
	m.SetQuestion(q, dns.TypeA)
	m.RecursionDesired = true
	response, _, err := localc.Exchange(m, server)
	if err != nil {
		log.Printf("    Error looking up %s against %s: %s", q, server, err)
		return err
	}
	for _, r := range response.Answer {
		log.Printf("    Got response %s", r.String())
	}
	return nil
}

func find_ns_records(q, server string) ([]string, error) {
	var err error
	var name_servers []string
	for {
		log.Printf("  looking up NS records for %s against %s", q, server)
		m := new(dns.Msg)
		m.SetQuestion(q, dns.TypeNS)
		m.RecursionDesired = true
		response, _, err := localc.Exchange(m, server)
		if err != nil {
			log.Printf("  Error looking up %s against %s: %s", q, server, err)
			return name_servers, err
		}
		for _, r := range response.Answer {
			if ns, ok := r.(*dns.NS); ok {
				name_servers = append(name_servers, ns.Ns)
			}
		}
		if len(name_servers) > 0 {
			return name_servers, nil
		}
		q = q[strings.Index(q, ".")+1:]
		if q == "" {
			return name_servers, errors.New("No servers found")
		}
	}
	return name_servers, err
}

func check_server_ns(q, server string) error {
	name_servers, err := find_ns_records(q, server)
	if err != nil {
		log.Printf("  Error looking up name servers for %s against %s: %s", q, server, err)
		return err
	}
	for _, ns := range name_servers {
		log.Printf("    Got response %s", ns)
		check_server_ns_resolve(q, server, ns)
	}
	return nil
}
func check_server_ns_resolve(q, server, ns string) error {
	log.Printf("    Looking up A record for NS %s using local server %s", ns, server)
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
	if _, exists := successful_queries[q+ns]; exists {
		log.Printf("      Skipping duplicate lookup A record for %s using server %s", q, ns)
		return nil
	}

	log.Printf("      Looking up A record for %s using server %s", q, ns)
	m := new(dns.Msg)
	m.SetQuestion(q, dns.TypeA)
	m.RecursionDesired = false
	response, _, err := localc.Exchange(m, ns)
	if err != nil {
		log.Printf("        Error looking up %s against %s: %s", q, ns, err)
		return err
	}
	for _, r := range response.Answer {
		log.Printf("        Got response %s", r)
	}
	successful_queries[q+ns] = true
	return nil
}

func check(q string) {
	q = dns.Fqdn(q)
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
