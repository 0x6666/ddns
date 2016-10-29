package ddns

import (
	"errors"
	"net"
	"time"

	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/config"
	"github.com/miekg/dns"
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
}

func NewHandler() *DDNSHandler {

	var (
		cacheConfig     config.CacheSettings
		resolver        *Resolver
		cache, negCache Cache
	)

	if config.Data.Resolv.Enable {
		resolvConfig := config.Data.Resolv
		clientConfig, err := dns.ClientConfigFromFile(resolvConfig.ResolvFile)
		if err != nil {
			log.Warn(":%s is not a valid resolv.conf file\n", resolvConfig.ResolvFile)
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
		hosts = NewHosts(config.Data.Hosts, config.Data.Redis)
	}

	return &DDNSHandler{resolver, cache, negCache, hosts}
}

func (h *DDNSHandler) close() {
	if h.hosts.hostWatcher != nil {
		h.hosts.hostWatcher.Close()
	}
}

func (h *DDNSHandler) do(Net string, w dns.ResponseWriter, req *dns.Msg) {
	q := req.Question[0]
	Q := Question{UnFqdn(q.Name), dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass]}

	var remote net.IP
	if Net == "tcp" {
		remote = w.RemoteAddr().(*net.TCPAddr).IP
	} else {
		remote = w.RemoteAddr().(*net.UDPAddr).IP
	}
	log.Info("%s lookup　%s", remote, Q.String())

	IPQuery := h.isIPQuery(q)

	// Query hosts
	if config.Data.Hosts.Enable && IPQuery > 0 {
		if ips, ok := h.hosts.Get(Q.qname, IPQuery); ok {
			m := new(dns.Msg)
			m.SetReply(req)

			switch IPQuery {
			case _IP4Query:
				rr_header := dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    config.Data.Hosts.TTL,
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
					Ttl:    config.Data.Hosts.TTL,
				}
				for _, ip := range ips {
					aaaa := &dns.AAAA{rr_header, ip}
					m.Answer = append(m.Answer, aaaa)
				}
			}

			w.WriteMsg(m)
			log.Debug("%s found in hosts file", Q.qname)
			return
		} else {
			log.Debug("%s didn't found in hosts file", Q.qname)
		}
	}

	// Only query cache when qtype == 'A'|'AAAA' , qclass == 'IN'
	key := KeyGen(Q)
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

	var err error
	var mesg *dns.Msg
	if config.Data.Resolv.Enable {
		mesg, err = h.resolver.Lookup(Net, req)
	} else {
		err = errors.New("Local resolution failed with no external resolution")
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

func (h *DDNSHandler) DoTCP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("tcp", w, req)
}

func (h *DDNSHandler) DoUDP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("udp", w, req)
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