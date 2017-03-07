package ddns

import (
	"crypto/md5"
	"errors"
	"net"
	"strings"

	"fmt"

	"github.com/miekg/dns"
	"github.com/yangsongfwd/backup/log"
	"github.com/yangsongfwd/ddns/config"
	"github.com/yangsongfwd/ddns/data"
	"github.com/yangsongfwd/ddns/errs"
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
	dbrecodes       *DBRecodes
}

func NewHandler(db data.IDatabase) *DDNSHandler {

	var (
		cacheConfig     config.CacheSettings
		resolver        *Resolver
		cache, negCache Cache
		err             error
	)

	if config.Data.Resolv.Enable {
		resolvConfig := config.Data.Resolv
		path := resolvConfig.ResolvFile
		if path[0] != '/' {
			path = config.CurDir() + "/" + path
		}

		clientConfig, err := dns.ClientConfigFromFile(path)
		if err != nil {
			log.Warn(":%s is not a valid resolv.conf file\n", path)
			log.Error("%v", err)
			panic(err)
		}
		clientConfig.Timeout = resolvConfig.Timeout
		resolver = &Resolver{clientConfig}
	}

	cacheConfig = config.Data.Cache
	switch cacheConfig.Backend {
	case "memory":
		cache, _ = NewMemCache(cacheConfig.Expire)
		negCache, _ = NewMemCache(cacheConfig.Expire)
	case "redis":
		cache, err = NewRedisCache("dns_cache", 1, cacheConfig.Expire)
		if err != nil {
			log.Error(err.Error())
			return nil
		}

		negCache, _ = NewRedisCache("dns_negcache", 2, cacheConfig.Expire)
		if err != nil {
			log.Error(err.Error())
			return nil
		}
	default:
		log.Error("Invalid cache backend %s", cacheConfig.Backend)
		panic("Invalid cache backend")
	}

	var hosts Hosts
	if config.Data.Hosts.Enable {
		hosts = NewHosts(config.Data.Hosts)
	}

	var recodes *DBRecodes
	if db != nil {
		recodes = NewDBRecodes(db)
	}

	return &DDNSHandler{resolver, cache, negCache, hosts, recodes}
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
	log.Info("%s lookup　%s", remote, Q.String())

	//
	// query in cache
	//
	key := KeyGen(Q)
	mesg, err := h.cache.Get(key)
	if err != nil {
		if mesg, err = h.negCache.Get(key); err != nil {
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
	//	query in database
	//
	if h.dbrecodes != nil {
		if q.Qtype == dns.TypePTR && strings.HasSuffix(Q.qname, ".in-addr.arpa") || strings.HasSuffix(Q.qname, ".ip6.arpa") {
			resp := h.ServeDNSReverse(w, req)
			if resp != nil {
				log.Debug("%s found in database", Q.String())
				return
			}
		} else {
			m := new(dns.Msg)
			m.SetReply(req)
			m.RecursionAvailable = true

			switch q.Qtype {
			case dns.TypeA, dns.TypeAAAA:
				rs := h.AddrRecode(Q.qname, q.Qtype, nil)
				if len(rs) > 0 {
					m.Answer = append(m.Answer, rs...)
				}
			case dns.TypeCNAME:
				rs := h.CNAMERecode(Q.qname)
				if len(rs) > 0 {
					m.Answer = append(m.Answer, rs...)
				}
			}

			if len(m.Answer) > 0 {
				if err := w.WriteMsg(m); err != nil {
					log.Error(err.Error())
				} else {
					log.Debug("%s found in database", Q.String())
				}
				return
			}
		}
		log.Debug("%s didn't found in database", Q.String())
	}

	//
	//	query in host file
	//
	if config.Data.Hosts.Enable && IPQuery > 0 {
		if ips, ok := h.hosts.Get(Q.qname, q.Qtype); ok {
			rspByIps(ips, config.Data.Hosts.TTL)
			log.Debug("%s found in hosts file", Q.String())
			return
		}
		log.Debug("%s didn't found in hosts file", Q.String())
	}

	//
	//	external resolution
	//
	if config.Data.Resolv.Enable {
		mesg, err = h.resolver.Lookup(netType, req)
	} else {
		err = errors.New("no external resolution")
	}

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

func (h *DDNSHandler) AddrRecode(domain string, qtype uint16, prrs []dns.RR) (rrs []dns.RR) {
	domain = dns.Fqdn(domain)
	rs := h.dbrecodes.GetAddrCname(domain, qtype)
	if len(rs) == 0 {
		return rrs
	}

	for _, r := range rs {
		if r.Type == qtype { // A | AAAA
			ip := net.ParseIP(r.Host)
			if ip != nil {
				if ip.To4() != nil && (qtype == dns.TypeA) {
					ip = ip.To4()
				} else if ip.To4() == nil && (qtype == dns.TypeAAAA) {
					ip = ip.To16()
				} else {
					continue
				}
			}

			rrs = append(rrs, &dns.A{
				Hdr: dns.RR_Header{
					Name:   domain,
					Rrtype: qtype,
					Class:  dns.ClassINET,
					Ttl:    r.TTL,
				},
				A: ip})
		} else if r.Type == dns.TypeCNAME {
			if domain == r.Host {
				log.Warn("CNAME loop detected: %q -> %q", domain, domain)
				continue
			}

			cname := r.newCNAME(domain)
			if len(prrs) > 7 {
				log.Warn("CNAME lookup limit of 8 exceeded for %s", cname)
				continue
			}

			if h.isDuplicateCNAME(cname, prrs) {
				log.Warn("CNAME loop detected for record %s", cname)
				continue
			}

			nextRecords := h.AddrRecode(r.Host, qtype, append(prrs, cname))
			if len(nextRecords) > 0 {
				rrs = append(rrs, cname)
				rrs = append(rrs, nextRecords...)
			}
			continue
		} else {
			log.Warn("invlid type [%v]", r.Type)
		}
	}
	return rrs
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
	if r := h.dbrecodes.GetCNAMERecode(name); r != nil {
		rs = append(rs, r.newCNAME(dns.Fqdn(name)))
	}
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
	name := strings.ToLower(q.Name)
	if h.dbrecodes != nil {
		vs := h.dbrecodes.ReverseGet(name)
		if len(vs) == 0 {
			return nil, errs.ErrPtrRecodeNotFound
		}

		for _, r := range vs {
			records = append(records, r.newPTR(name))
		}

		return records, nil
	}

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
