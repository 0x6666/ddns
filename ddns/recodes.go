package ddns

import (
	"net"
	"time"

	"sync"

	"strings"

	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/data"
)

type recode struct {
	domain string
	ip     string
	ttl    int
}

type DBRecodes struct {
	db data.IDatabase

	sync.RWMutex

	cache map[string]recode

	cacheVersion int64
}

func NewDBRecodes(db data.IDatabase) *DBRecodes {

	dr := &DBRecodes{db: db}
	dr.cache = map[string]recode{}
	dr.refresh()
	return dr
}

func (d *DBRecodes) Get(domain string, family int) ([]net.IP, int, bool) {

	d.RLock()
	defer d.RUnlock()

	dm := strings.ToLower(domain)
	if r, exist := d.cache[dm]; exist {
		var ip net.IP
		switch family {
		case _IP4Query:
			ip = net.ParseIP(r.ip).To4()
		case _IP6Query:
			ip = net.ParseIP(r.ip).To16()

		}

		return []net.IP{ip}, r.ttl, true
	}
	return nil, 0, false
}

func (d *DBRecodes) update() {

	if d.cacheVersion == d.db.GetVersion() {
		return
	}

	datas, err := d.db.ReadData(-1, -1)
	if err != nil {
		log.Error("ReadData failed: %v", err)
		return
	}

	cache := map[string]recode{}
	for _, data := range datas {
		if len(data.RecordName) == 0 || len(data.RecodeValue) == 0 {
			continue
		}
		r := recode{}
		r.domain = strings.ToLower(data.RecordName)
		r.ip = data.RecodeValue
		r.ttl = data.TTL
		if r.ttl <= 0 {
			r.ttl = 600
		}
		cache[r.domain] = r
	}

	d.Lock()
	defer d.Unlock()
	d.cache = cache
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
