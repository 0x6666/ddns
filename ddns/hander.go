package ddns

import (
	"crypto/md5"
	"net"
	"strings"

	"fmt"

	"github.com/0x6666/ddnsx/errs"
	"github.com/0x6666/util/log"
	"github.com/miekg/dns"
)

type NetType int

const (
	NetTCP NetType = 1
	NetUDP         = 2
)

const (
	notIPQuery = 0
	_IP4Query  = 4
	_IP6Query  = 6
)

func KeyGen(q Question) string {
	h := md5.New()
	h.Write([]byte(q.String()))
	x := h.Sum(nil)
	key := fmt.Sprintf("%x", x)
	return key
}

func UnFqdn(s string) string {
	if dns.IsFqdn(s) {
		return s[:len(s)-1]
	}
	return s
}

type Question struct {
	qname  string
	qtype  string
	qclass string
}

func (q *Question) String() string {
	return q.qname + " " + q.qclass + " " + q.qtype
}

type DDNSHandler struct {
	resolver        *Resolver
	cache, negCache Cache
	hosts           Hosts
}

func NewHandler() *DDNSHandler {

	var (
		resolver        *Resolver
		cache, negCache Cache
		err             error
	)

	clientConfig, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		log.Error("%v", err)
		panic(err)
	}
	clientConfig.Timeout = 5
	resolver = &Resolver{clientConfig}

	cache, _ = NewMemCache(1200)
	negCache, _ = NewMemCache(1200)

	var hosts Hosts
	hosts = NewHosts()

	return &DDNSHandler{resolver, cache, negCache, hosts}
}

func (h *DDNSHandler) close() {
	if h.hosts.hostWatcher != nil {
		h.hosts.hostWatcher.Close()
	}
}

func (h *DDNSHandler) do(netType NetType, w dns.ResponseWriter, req *dns.Msg) {

	q := req.Question[0]
	if q.Qtype == dns.TypeANY {
		m := new(dns.Msg)
		m.SetReply(req)
		m.Authoritative = false
		m.Rcode = dns.RcodeRefused
		m.RecursionAvailable = false
		m.RecursionDesired = false
		m.Compress = false
		w.WriteMsg(m)
		return
	}

	Q := Question{strings.ToLower(UnFqdn(q.Name)), dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass]}

	var remote net.IP
	if netType == NetTCP {
		remote = w.RemoteAddr().(*net.TCPAddr).IP
	} else {
		remote = w.RemoteAddr().(*net.UDPAddr).IP
	}
	log.Info("remote: %s, lookup: %s", remote, Q.String())

	//
	// query in cache
	//
	key := KeyGen(Q)
	mesg, err := h.cache.Get(key)
	if err != nil {
		if _, err = h.negCache.Get(key); err != nil {
			log.Debug("%s didn't hit cache", Q.String())
		} else {
			log.Debug("%s hit negative cache", Q.String())
			dns.HandleFailed(w, req)
			return
		}
	} else if mesg != nil {
		log.Debug("%s hit cache", Q.String())
		// we need this copy against concurrent modification of Id
		msg := *mesg
		msg.Id = req.Id
		w.WriteMsg(&msg)
		return
	}
	IPQuery := h.isIPQuery(q)

	rspByIps := func(ips []net.IP, ttl uint32) {
		m := new(dns.Msg)
		m.SetReply(req)
		m.RecursionAvailable = true

		switch IPQuery {
		case _IP4Query:
			rr_header := dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    ttl,
			}
			for _, ip := range ips {
				a := &dns.A{rr_header, ip}
				m.Answer = append(m.Answer, a)
			}
		case _IP6Query:
			rr_header := dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    ttl,
			}
			for _, ip := range ips {
				aaaa := &dns.AAAA{rr_header, ip}
				m.Answer = append(m.Answer, aaaa)
			}
		}

		w.WriteMsg(m)
	}

	//
	//	query in host file
	//
	if IPQuery > 0 {
		if ips, ok := h.hosts.Get(Q.qname, q.Qtype); ok {
			rspByIps(ips, 600)
			log.Debug("%s found in hosts file", Q.String())
			return
		}
		log.Debug("%s didn't found in hosts file", Q.String())
	}

	//
	//	external resolution
	//
	mesg, err = h.resolver.Lookup(netType, req)

	if err != nil {
		log.Debug("Resolve query error %s, insert to negative cache", err)
		dns.HandleFailed(w, req)

		// cache the failure, too!
		if err = h.negCache.Set(key, nil); err != nil {
			log.Debug("Set %s negative cache failed: %v", Q.String(), err)
		}
		return
	}

	w.WriteMsg(mesg)

	if len(mesg.Answer) > 0 {
		log.Debug("resolv res: %v", mesg.Answer)
		err = h.cache.Set(key, mesg)
		if err != nil {
			log.Debug("Set %s cache failed: %s", Q.String(), err.Error())
		}
		log.Debug("Insert %s into cache", Q.String())
	}
}

func (h *DDNSHandler) isDuplicateCNAME(r *dns.CNAME, records []dns.RR) bool {
	for _, rec := range records {
		if v, ok := rec.(*dns.CNAME); ok {
			if v.Target == r.Target {
				return true
			}
		}
	}
	return false
}

func (h *DDNSHandler) CNAMERecode(name string) (rs []dns.RR) {
	return rs
}

// ServeDNSReverse is the handler for DNS requests for the reverse zone. If nothing is found
// locally the request is forwarded to the forwarder for resolution.
func (h *DDNSHandler) ServeDNSReverse(w dns.ResponseWriter, req *dns.Msg) *dns.Msg {
	m := new(dns.Msg)
	m.SetReply(req)
	m.Compress = true
	m.Authoritative = false // Set to false, because I don't know what to do wrt DNSSEC.
	m.RecursionAvailable = true
	var err error
	if m.Answer, err = h.PTRRecords(req.Question[0]); err == nil {
		// TODO Reverse DNSSEC. We should sign this, but requires a key....and more
		// Probably not worth the hassle?
		if err := w.WriteMsg(m); err != nil {
			log.Error("failure to return reply %q", err)
		}
		return m
	}

	return nil
}

func (h *DDNSHandler) PTRRecords(q dns.Question) (records []dns.RR, err error) {
	return nil, errs.ErrPtrRecodeNotFound
}

func (h *DDNSHandler) DoTCP(w dns.ResponseWriter, req *dns.Msg) {
	h.do(NetTCP, w, req)
}

func (h *DDNSHandler) DoUDP(w dns.ResponseWriter, req *dns.Msg) {
	h.do(NetUDP, w, req)
}

func (h *DDNSHandler) isIPQuery(q dns.Question) int {
	if q.Qclass != dns.ClassINET {
		return notIPQuery
	}

	switch q.Qtype {
	case dns.TypeA:
		return _IP4Query
	case dns.TypeAAAA:
		return _IP6Query
	default:
		return notIPQuery
	}
}
