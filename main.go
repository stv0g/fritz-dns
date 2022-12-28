package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mdlayher/netx/eui64"
	"github.com/miekg/dns"
	"github.com/philippfranke/go-fritzbox/fritzbox"
)

var (
	zone string
	port int

	fbUser     string
	fbPassword string
	fbUrl      string

	fbIPv6ULAPrefixStr string

	soaNS      string
	soaMbox    string
	soaRefresh time.Duration
	soaRetry   time.Duration
	soaExpire  time.Duration
	soaMinTTL  time.Duration

	ttl time.Duration

	mu            sync.Mutex
	soa           dns.RR
	records       = map[uint16]map[string]dns.RR{}
	lastUpdate    time.Time
	ipv6ULAPrefix net.IP
	ipv6Prefix    net.IP
)

func main() {
	var err error

	flag.DurationVar(&soaRefresh, "soa-refresh", 2*time.Hour, "SOA refresh value")
	flag.DurationVar(&soaRetry, "soa-retry", time.Hour, "SOA retry value")
	flag.DurationVar(&soaExpire, "soa-expire", 31*24*time.Hour, "SOA expire value")
	flag.DurationVar(&soaMinTTL, "soa-minttl", time.Hour, "SOA minimum TTL value")
	flag.StringVar(&soaMbox, "soa-mbox", "", "SOA mailbox value")
	flag.StringVar(&soaNS, "soa-ns", "", "Authorative DNS server for the zone")

	flag.StringVar(&zone, "zone", "fritz.box.", "DNS Zone")
	flag.DurationVar(&ttl, "ttl", 5*time.Minute, "default TTL values for records")

	flag.IntVar(&port, "port", 53, "Listen port")

	flag.StringVar(&fbUser, "user", "admin", "FritzBox username")
	flag.StringVar(&fbPassword, "pass", "", "FritzBox password")
	flag.StringVar(&fbUrl, "url", "http://fritz.box/", "FritzBox URL")
	flag.StringVar(&fbIPv6ULAPrefixStr, "ipv6-ula-prefix", "fd00::/64", "Fritz Box IPv6 ULA Prefix")

	flag.Parse()

	zone = dns.Fqdn(zone)

	if soaMbox == "" {
		soaMbox = "hostmaster." + zone
	}

	if soaNS == "" {
		soaNS = "ns1." + zone
	}

	soa = &dns.SOA{
		Hdr:     dns.RR_Header{Name: zone, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: uint32(ttl.Seconds())},
		Ns:      soaNS,
		Mbox:    soaMbox,
		Serial:  uint32(time.Now().UnixMilli()),
		Refresh: uint32(soaRefresh.Seconds()),
		Retry:   uint32(soaRetry.Seconds()),
		Expire:  uint32(soaExpire.Seconds()),
		Minttl:  uint32(soaMinTTL.Seconds()),
	}

	if fbIPv6ULAPrefixStr != "" {
		if _, pfx, err := net.ParseCIDR(fbIPv6ULAPrefixStr); err != nil {
			log.Fatalf("Failed to parse IPv6 ULA prefix: %s", err)
		} else if ones, bits := pfx.Mask.Size(); ones != 64 || bits != net.IPv6len*8 {
			log.Fatalf("Invalid IPv6 prefix")
		} else {
			ipv6ULAPrefix = pfx.IP
		}
	}

	// attach request handler func
	dns.HandleFunc(zone, handleRequest)

	// start udpSvr
	udpSvr := &dns.Server{
		Addr: ":" + strconv.Itoa(port),
		Net:  "udp",
	}

	tcpSvr := &dns.Server{
		Addr: ":" + strconv.Itoa(port),
		Net:  "tcp",
	}

	log.Printf("Start listening for DNS requests at port %d for zone %s\n", port, zone)

	for _, svr := range []*dns.Server{udpSvr, tcpSvr} {
		go func(svr *dns.Server) {
			err = svr.ListenAndServe()
			defer svr.Shutdown()

			if err != nil {
				log.Fatalf("Failed to start server: %s\n ", err.Error())
			}
		}(svr)
	}

	select {}
}

func handleQuery(m *dns.Msg) error {
	if lastUpdate.Add(ttl).Before(time.Now()) {
		newRecords, err := updateRecords()
		if err != nil {
			return fmt.Errorf("failed to update records: %w", err)
		}

		mu.Lock()
		records = newRecords
		lastUpdate = time.Now()
		mu.Unlock()
	}

	for _, q := range m.Question {
		switch q.Qtype {
		// case dns.TypePTR:
		// 	log.Printf("Reverse lookup for %s", q.Name)

		case dns.TypeA, dns.TypeAAAA:
			log.Printf("Forward lookup for %s", q.Name)

			if rr, ok := records[q.Qtype][q.Name]; ok {
				m.Answer = append(m.Answer, rr)
			}

		case dns.TypeAXFR:
			log.Printf("Handle zone transfer for %s", q.Name)

			if q.Name == zone {
				m.Answer = append(m.Answer, soa)

				for _, rrs := range records {
					for _, rr := range rrs {
						m.Answer = append(m.Answer, rr)
					}
				}

				m.Answer = append(m.Answer, soa)
			}
		}
	}

	return nil
}

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		if err := handleQuery(m); err != nil {
			m.Rcode = dns.RcodeServerFailure

			log.Printf("Failed to handle query: %s", err)
		}
	}

	if err := w.WriteMsg(m); err != nil {
		log.Printf("Failed to write message: %s", err)
	}

	if err := w.Close(); err != nil {
		log.Printf("Failed to close connection: %s", err)
	}
}

func updateRecords() (map[uint16]map[string]dns.RR, error) {
	var err error

	c := fritzbox.NewClient(nil)
	if c.BaseURL, err = url.Parse(fbUrl); err != nil {
		log.Fatalf("Failed to parse Fritz Box URL: %s", err)
	}

	if err := c.Auth(fbUser, fbPassword); err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	if ipv6Prefix, err = getIPv6Prefix(); err != nil {
		log.Printf("Failed to get IPv6 prefix: %s", err)

		ipv6Prefix = ipv6ULAPrefix
	}

	u := *c.BaseURL
	u.Host += ":49000"
	u.Path = "/devicehostlist.lua"

	req, err := c.NewRequest("GET", u.String(), url.Values{})
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var respXml Response
	if _, err = c.Do(req, &respXml); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	rrsA := map[string]dns.RR{}
	rrsAAAA := map[string]dns.RR{}

	for _, item := range respXml.Items {
		fqdn := strings.ToLower(item.HostName) + "." + zone

		var rr dns.RR = &dns.A{
			Hdr: dns.RR_Header{Name: fqdn, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: uint32(ttl.Seconds())},
			A:   net.ParseIP(item.IPAddress),
		}

		rrsA[fqdn] = rr

		if ipv6Prefix != nil {
			mac, err := net.ParseMAC(item.MACAddress)
			if err != nil {
				return nil, fmt.Errorf("failed to parse MAC address: %s: %w", item.MACAddress, err)
			}

			aaaa, err := eui64.ParseMAC(ipv6Prefix, mac)
			if err != nil {
				return nil, fmt.Errorf("failed to generate EUI-64 address: %w", err)
			}

			rrsAAAA[fqdn] = &dns.AAAA{
				Hdr:  dns.RR_Header{Name: fqdn, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: uint32(ttl.Seconds())},
				AAAA: aaaa,
			}
		}
	}

	rrs := map[uint16]map[string]dns.RR{
		dns.TypeA:    rrsA,
		dns.TypeAAAA: rrsAAAA,
	}

	log.Printf("Updated %d hosts from Fritz Box", len(respXml.Items))

	return rrs, nil
}

func getIPv6Prefix() (net.IP, error) {
	return nil, errors.New("not supported yet")
}
