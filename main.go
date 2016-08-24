package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// A option type for accepting multiple strings via flag module
type stringSlice []string

func (ss *stringSlice) String() string {
	return fmt.Sprint(*ss)
}
func (ss *stringSlice) Set(value string) error {
	*ss = append(*ss, value)
	return nil
}

var (
	localc            *dns.Client
	conf              *dns.ClientConfig
	performed_queries map[string]bool
	retries           int
	timeout           time.Duration
	addtionalNS       stringSlice

	ErrEmptyResponse = errors.New("Empty response")
	ErrNXDOMAIN      = errors.New("NXDOMAIN")
)

func init() {
	flag.Var(&addtionalNS, "ns", "Additional name servers to query")
	flag.IntVar(&retries, "retries", 3, "Number of retries")
	flag.DurationVar(&timeout, "timeout", 5*time.Second, "timeout for queries")
	performed_queries = make(map[string]bool)
}

func ExchangeWithRetries(server string, query string, qtype uint16, recursionDesired bool, allowEmpty bool) (*dns.Msg, error) {
	m := new(dns.Msg)
	m.SetQuestion(query, qtype)
	m.RecursionDesired = recursionDesired
	delay := 1 * time.Second

	var err error
	var response *dns.Msg
	for attempt := 1; attempt <= retries; attempt++ {
		err = nil
		response, _, err = localc.Exchange(m, server)
		if err == nil {
			if len(response.Answer) == 0 && !allowEmpty {
				err = ErrEmptyResponse
			}
			if response.Rcode == dns.RcodeNameError {
				err = ErrNXDOMAIN
			}
		}
		if err == nil {
			break
		}
		if attempt != retries {
			log.Printf("    Error (retry %d/%d) looking up %s against %s: %s", attempt, retries, query, server, err)
			time.Sleep(delay)
			delay *= 2
		}
	}
	return response, err
}

func check_server(q string, server string) error {
	log.Printf("  Looking up %s using local server %s", q, server)
	response, err := ExchangeWithRetries(server, q, dns.TypeA, true, false)
	if err != nil {
		log.Printf("    Error looking up %s against %s: %s", q, server, err)
		return err
	}
	for _, r := range response.Answer {
		log.Printf("    Got response %s", r.String())
	}
	return err
}

func find_ns_records(q, server string) ([]string, error) {
	var err error
	var name_servers []string
	for {
		log.Printf("  looking up NS records for %s against %s", q, server)
		response, err := ExchangeWithRetries(server, q, dns.TypeNS, true, true)
		if err != nil && err != ErrEmptyResponse {
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
		//FIXME: This doesn't work for .co.uk etc
		if q == "" || strings.Count(q, ".") == 1 {
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
	response, err := ExchangeWithRetries(server, ns, dns.TypeA, true, false)
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
	if _, exists := performed_queries[q+ns]; exists {
		log.Printf("      Skipping duplicate lookup A record for %s using server %s", q, ns)
		return nil
	}
	performed_queries[q+ns] = true
	var err error

	log.Printf("      Looking up A record for %s using server %s", q, ns)
	response, err := ExchangeWithRetries(ns, q, dns.TypeA, false, false)
	if err != nil {
		log.Printf("        Error looking up %s against %s: %s", q, ns, err)
		return err
	}
	for _, r := range response.Answer {
		log.Printf("        Got response %s", r)
	}
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

	flag.Parse()

	var err error
	conf, err = dns.ClientConfigFromFile("/etc/resolv.conf")
	if conf == nil {
		log.Fatal("Cannot initialize the local resolver: %s\n", err)
	}
	conf.Servers = append(conf.Servers, addtionalNS...)
	localc = new(dns.Client)
	localc.ReadTimeout = timeout
	for _, domain := range flag.Args() {
		check(domain)
	}
}
