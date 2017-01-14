package ddns

import (
	"net"
	"time"

	"sync"

	"strings"

	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/data"
	"github.com/inimei/ddns/data/model"
	"github.com/miekg/dns"
)

func toDnsType(t model.RecodeType) uint16 {
	switch t {
	case model.A:
		return dns.TypeA
	case model.AAAA:
		return dns.TypeAAAA
	case model.CNAME:
		return dns.TypeCNAME
	default:
		return dns.TypeNone
	}
}

type recodeValue struct {
	val string
	ttl int
}

type fqdn string
type domainCache struct {
	domains  map[fqdn][]*model.Recode
	rdomains map[fqdn]recodeValue
}

func (d domainCache) append(domain string, r *model.Recode) {
	domain = dns.Fqdn(domain)
	exist := false

	if _, exist = d.domains[fqdn(domain)]; !exist {
		d.domains[fqdn(domain)] = []*model.Recode{}
	}

	d.domains[fqdn(domain)] = append(d.domains[fqdn(domain)], r)

	if r.RecordType == model.A || r.RecordType == model.AAAA {
		raddr, err := dns.ReverseAddr(r.RecodeValue)
		if err != nil {
			log.Error("get reverse addr failed: %v", err)
		} else {
			if _, b := d.rdomains[fqdn(raddr)]; b {
				log.Warn("reverse recode [%v] already exist, host [%v] domain [%v]", raddr, r.RecordHost, domain)
			} else {
				if r.RecordHost != "@" {
					domain = r.RecordHost + "." + domain
				}

				d.rdomains[fqdn(raddr)] = recodeValue{val: domain, ttl: r.TTL}
			}
		}
	}
}

func (d domainCache) getValue(host, domain string, qtype uint16) ([]net.IP, int, bool) {

	domain = dns.Fqdn(domain)
	ds, exist := d.domains[fqdn(domain)]
	if !exist {
		return nil, 0, false
	}

	switch qtype {
	case dns.TypeA, dns.TypeAAAA:

		ips := []net.IP{}
		ttl := 0
		for _, r := range ds {
			if toDnsType(r.RecordType) != qtype || r.RecordHost != host {
				continue
			}

			ip := net.ParseIP(r.RecodeValue)
			if ip == nil {
				log.Warn("invalid ip [%v], type [%v]", r.RecodeValue, dns.TypeToString[qtype])
				continue
			}
			if qtype == dns.TypeA {
				ip = ip.To4()
			} else {
				if ip.To4() == nil {
					ip = ip.To16()
				} else {
					ip = nil
				}
			}
			if ip != nil {
				ips = append(ips, ip)
				if ttl > r.TTL {
					ttl = r.TTL
				}
			}
		}
		if len(ips) > 0 {
			return ips, ttl, true
		}
	}
	return nil, 0, false
}

func (d domainCache) getReverseValue(name string) (string, int, bool) {
	name = dns.Fqdn(name)
	if v, b := d.rdomains[fqdn(name)]; b {
		return v.val, v.ttl, true
	}
	return "", 0, false
}

type DBRecodes struct {
	db data.IDatabase

	sync.RWMutex

	dcache       *domainCache
	cacheVersion int64
}

func NewDBRecodes(db data.IDatabase) *DBRecodes {

	dr := &DBRecodes{db: db}
	dr.dcache = &domainCache{}
	dr.cacheVersion = -1
	dr.refresh()
	return dr
}

func (d *DBRecodes) Get(domain string, qtype uint16) ([]net.IP, int, bool) {
	if qtype != dns.TypeA && qtype != dns.TypeAAAA {
		log.Debug("not implement for %v", dns.TypeToString[qtype])
		return nil, 0, false
	}

	d.RLock()
	defer d.RUnlock()

	dm := strings.ToLower(domain)
	ips, ttl, exist := d.dcache.getValue("@", dm, qtype)
	if exist {
		return ips, ttl, exist
	}

	hosts := strings.SplitN(dm, ".", 2)
	if len(hosts) == 2 {
		return d.dcache.getValue(hosts[0], hosts[1], qtype)
	}

	return nil, 0, false
}

func (d *DBRecodes) ReverseGet(name string) (string, int, bool) {
	d.RLock()
	defer d.RUnlock()

	return d.dcache.getReverseValue(name)
}

func (d *DBRecodes) update() {
	if d.cacheVersion == d.db.GetVersion() {
		return
	}

	ds, err := d.db.GetAllDomains(0, -1)
	if err != nil {
		log.Error("GetAllDomains failed: %v", err)
		return
	}

	dcache := &domainCache{map[fqdn][]*model.Recode{}, map[fqdn]recodeValue{}}
	for _, domain := range ds {
		recodes, err := d.db.GetRecodes(domain.ID, 0, -1)
		if err != nil {
			log.Error("get recodes domainID [%v]: %v", domain.ID, err)
			return
		}

		for _, r := range recodes {
			dcache.append(domain.DomainName, r)
		}
	}

	d.Lock()
	defer d.Unlock()
	d.dcache = dcache
	d.cacheVersion = d.db.GetVersion()
}

func (d *DBRecodes) refresh() {

	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for {
			d.update()
			<-ticker.C
		}
	}()
}
