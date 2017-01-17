package ddns

import (
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

type fqdn string
type domainCache struct {
	domains  map[fqdn][]*model.Recode
	rdomains map[fqdn][]recode
}

func (d domainCache) append(domain string, r *model.Recode) {
	domain = dns.Fqdn(domain)
	exist := false

	if _, exist = d.domains[fqdn(domain)]; !exist {
		d.domains[fqdn(domain)] = []*model.Recode{}
	}

	d.domains[fqdn(domain)] = append(d.domains[fqdn(domain)], r)

	if r.RecordType == model.A || r.RecordType == model.AAAA {
		addr, err := dns.ReverseAddr(r.RecodeValue)
		raddr := fqdn(addr)
		if err != nil {
			log.Error("get reverse addr failed: %v", err)
		} else {
			if r.RecordHost != "@" {
				domain = r.RecordHost + "." + domain
			}

			if _, b := d.rdomains[raddr]; !b {
				d.rdomains[raddr] = []recode{}
			}
			d.rdomains[raddr] = append(d.rdomains[raddr], recode{domain, r.TTL, dns.TypePTR})
		}
	}
}

func (d domainCache) getAddrValue(host, domain string, qtype uint16) []recode {

	if qtype != dns.TypeA && qtype != dns.TypeAAAA {
		return nil
	}

	domain = dns.Fqdn(domain)
	ds, exist := d.domains[fqdn(domain)]
	if !exist {
		return nil
	}

	ips := []recode{}
	for _, r := range ds {
		if toDnsType(r.RecordType) != qtype || r.RecordHost != host {
			continue
		}

		ips = append(ips, recode{r.RecodeValue, r.TTL, qtype})
	}

	return ips
}

func (d domainCache) getCnameValue(host, domain string) *recode {

	domain = dns.Fqdn(domain)
	ds, exist := d.domains[fqdn(domain)]
	if !exist {
		return nil
	}

	for _, r := range ds {
		if r.RecordType == model.CNAME && r.RecordHost == host {
			return &recode{r.RecodeValue, r.TTL, dns.TypeCNAME}
		}
	}

	return nil
}

func (d domainCache) getReverseValue(addr string) []recode {
	name := fqdn(dns.Fqdn(addr))
	return d.rdomains[name]
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

func (d *DBRecodes) GetAddress(domain string, qtype uint16) []recode {
	if qtype != dns.TypeA && qtype != dns.TypeAAAA {
		return nil
	}

	d.RLock()
	defer d.RUnlock()

	d.RLock()
	defer d.RUnlock()

	dm := strings.ToLower(domain)
	ips := d.dcache.getAddrValue("@", dm, qtype)

	hosts := strings.SplitN(dm, ".", 2)
	if len(hosts) == 2 {
		sips := d.dcache.getAddrValue(hosts[0], hosts[1], qtype)
		if len(sips) > 0 {
			ips = append(ips, sips...)
		}
	}

	return ips
}

func (d *DBRecodes) GetCNAMERecode(domain string) *recode {

	dm := strings.ToLower(domain)
	hosts := strings.SplitN(dm, ".", 2)
	if len(hosts) == 2 {
		return d.dcache.getCnameValue(hosts[0], hosts[1])
	}

	return nil
}

func (d *DBRecodes) GetAddrCname(domain string, qtype uint16) []recode {

	if qtype != dns.TypeA && qtype != dns.TypeAAAA || qtype == dns.TypeCNAME {
		return nil
	}

	d.RLock()
	defer d.RUnlock()

	dm := strings.ToLower(domain)
	ips := d.dcache.getAddrValue("@", dm, qtype)

	hosts := strings.SplitN(dm, ".", 2)
	if len(hosts) == 2 {
		sips := d.dcache.getAddrValue(hosts[0], hosts[1], qtype)
		if len(sips) > 0 {
			ips = append(ips, sips...)
		}

		if cname := d.dcache.getCnameValue(hosts[0], hosts[1]); cname != nil {
			ips = append(ips, *cname)
		}
	}

	return ips
}

func (d *DBRecodes) ReverseGet(name string) []recode {
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

	log.Info("update data cache")

	dcache := &domainCache{map[fqdn][]*model.Recode{}, map[fqdn][]recode{}}
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

	d.update()
	d.db.RegListener(d.update)

	/*ticker := time.NewTicker(time.Second * 5)
	go func() {
		for {
			d.update()
			<-ticker.C
		}
	}()
	*/
}
