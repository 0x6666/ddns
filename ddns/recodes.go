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

type recodeValue struct {
	val string
	ttl int
}

type recode map[uint16]recodeValue //map[recodeType]RecodeValue
type recodes map[string]recode     //map[RecordName]recode

type DBRecodes struct {
	db data.IDatabase

	sync.RWMutex
	cache        map[string]recodes     //map[domain]recodes
	rcache       map[string]recodeValue //map[ip.in-addr.arpa.]domain
	cacheVersion int64
}

func NewDBRecodes(db data.IDatabase) *DBRecodes {

	dr := &DBRecodes{db: db}
	dr.cache = map[string]recodes{}
	dr.cacheVersion = -1
	go func() { dr.update() }()
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

	_getVal := func(host, domain string) ([]net.IP, int, bool) {
		d, exist := d.cache[domain]
		if !exist {
			return nil, 0, false
		}

		switch qtype {
		case dns.TypeA, dns.TypeAAAA:
			if r, e := d[host]; e {
				if v, e := r[qtype]; e {
					ip := net.ParseIP(v.val)
					if ip == nil {
						log.Error("invalid ip [%v], type [%v]", v.val, qtype)
						return nil, 0, false
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
						return []net.IP{ip}, v.ttl, true
					}
				}
			}
		}
		return nil, 0, false
	}

	dm := strings.ToLower(domain)
	ips, ttl, exist := _getVal("@", dm)
	if exist {
		return ips, ttl, exist
	}

	hosts := strings.SplitN(dm, ".", 2)
	if len(hosts) == 2 {
		return _getVal(hosts[0], hosts[1])
	}

	return nil, 0, false
}

func (d *DBRecodes) ReverseGet(name string) (string, int, bool) {

	d.RLock()
	defer d.RUnlock()

	if v, b := d.rcache[dns.Fqdn(name)]; b {
		return v.val, v.ttl, true
	}

	return "", 0, false
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

	rcache := map[string]recodeValue{}
	cache := map[string]recodes{}
	for _, domain := range ds {
		recodes, err := d.db.GetRecodes(domain.ID, 0, -1)
		if err != nil {
			log.Error("get recodes domainID [%v]: %v", domain.ID, err)
			return
		}

		rc := map[string]recode{}
		for _, r := range recodes {
			v := recodeValue{val: r.RecodeValue, ttl: r.TTL}
			if v.ttl <= 0 {
				v.ttl = 600
			}

			if r.RecordType == model.A {
				rc[r.RecordHost] = recode{dns.TypeA: v}
			} else if r.RecordType == model.AAAA {
				rc[r.RecordHost] = recode{dns.TypeAAAA: v}
			}

			raddr, err := dns.ReverseAddr(r.RecodeValue)
			if err != nil {
				log.Error("get revers addr failed: %v", err)
			} else {
				if _, b := rcache[raddr]; b {
					log.Warn("revers recode [%v] already exist", raddr)
				} else {
					dname := domain.DomainName
					if r.RecordHost != "@" {
						dname = r.RecordHost + "." + dname
					}

					rcache[raddr] = recodeValue{val: dname, ttl: r.TTL}
				}
			}
		}
		cache[domain.DomainName] = rc
	}

	d.Lock()
	defer d.Unlock()
	d.cache = cache
	d.rcache = rcache
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
