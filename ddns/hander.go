package ddns

import (
	"errors"
	"net"
	"strings"
	"time"

	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/config"
	"github.com/inimei/ddns/data"
	"github.com/inimei/ddns/errs"
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
		cache = &MemoryCache{
			Backend:  make(map[string]Mesg, cacheConfig.Maxcount),
			Expire:   time.Duration(cacheConfig.Expire) * time.Second,
			Maxcount: cacheConfig.Maxcount,
		}
		negCache = &MemoryCache{
			Backend:  make(map[string]Mesg),
			Expire:   time.Duration(cacheConfig.Expire) * time.Second / 2,
			Maxcount: cacheConfig.Maxcount,
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

	key := KeyGen(Q)

	//
	//	query in database
	//
	if h.dbrecodes != nil {
		if q.Qtype == dns.TypePTR && strings.HasSuffix(Q.qname, ".in-addr.arpa") || strings.HasSuffix(Q.qname, ".ip6.arpa") {
			resp := h.ServeDNSReverse(w, req)
			if resp != nil {
				err := h.cache.Set(key, resp)
				if err != nil {
					log.Error(err.Error())
				}
			}

			return
		}

		if ips, ttl, ok := h.dbrecodes.Get(Q.qname, q.Qtype); ok {
			rspByIps(ips, uint32(ttl))
			log.Debug("%s found in database", Q.String())
			return
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
	// query in cache
	//
	if IPQuery > 0 {
		mesg, err := h.cache.Get(key)
		if err != nil {
			if mesg, err = h.negCache.Get(key); err != nil {
				log.Debug("%s didn't hit cache", Q.String())
			} else {
				log.Debug("%s hit negative cache", Q.String())
				dns.HandleFailed(w, req)
				return
			}
		} else {
			log.Debug("%s hit cache", Q.String())
			// we need this copy against concurrent modification of Id
			msg := *mesg
			msg.Id = req.Id
			w.WriteMsg(&msg)
			return
		}
	}

	//
	//	external resolution
	//
	var err error
	var mesg *dns.Msg
	if config.Data.Resolv.Enable {
		mesg, err = h.resolver.Lookup(netType, req)
	} else {
		err = errors.New("no external resolution")
	}

	if err != nil {
		log.Debug("Resolve query error %s", err)
		dns.HandleFailed(w, req)

		// cache the failure, too!
		if err = h.negCache.Set(key, nil); err != nil {
			log.Debug("Set %s negative cache failed: %v", Q.String(), err)
		}
		return
	}

	w.WriteMsg(mesg)

	if IPQuery > 0 && len(mesg.Answer) > 0 {
		err = h.cache.Set(key, mesg)
		if err != nil {
			log.Debug("Set %s cache failed: %s", Q.String(), err.Error())
		}
		log.Debug("Insert %s into cache", Q.String())
	}
}

// ServeDNSReverse is the handler for DNS requests for the reverse zone. If nothing is found
// locally the request is forwarded to the forwarder for resolution.
func (s *DDNSHandler) ServeDNSReverse(w dns.ResponseWriter, req *dns.Msg) *dns.Msg {
	m := new(dns.Msg)
	m.SetReply(req)
	m.Compress = true
	m.Authoritative = false // Set to false, because I don't know what to do wrt DNSSEC.
	m.RecursionAvailable = true
	var err error
	if m.Answer, err = s.PTRRecords(req.Question[0]); err == nil {
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
		d, ttl, exist := h.dbrecodes.ReverseGet(name)
		if !exist {
			return nil, errs.ErrPtrRecodeNotFound
		}

		ptr := &dns.PTR{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: uint32(ttl)}, Ptr: dns.Fqdn(d)}
		records = append(records, ptr)
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

func UnFqdn(s string) string {
	if dns.IsFqdn(s) {
		return s[:len(s)-1]
	}
	return s
}
