package ddns

import (
	"net"

	"github.com/miekg/dns"
)

type recode struct {
	Host string // address or domain
	TTL  uint32
	Type uint16 // dns.Type
}

func (s *recode) newA(name string, ip net.IP) *dns.A {
	return &dns.A{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: s.TTL}, A: ip}
}

func (s *recode) newAAAA(name string, ip net.IP) *dns.AAAA {
	return &dns.AAAA{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: s.TTL}, AAAA: ip}
}

func (s *recode) newCNAME(name string) *dns.CNAME {
	return &dns.CNAME{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: s.TTL}, Target: s.Host}
}

func (s *recode) newNS(name string, target string) *dns.NS {
	return &dns.NS{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: s.TTL}, Ns: target}
}

func (s *recode) newPTR(name string) *dns.PTR {
	return &dns.PTR{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: s.TTL}, Ptr: dns.Fqdn(s.Host)}
}
