package ddns

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/config"
	"github.com/inimei/ddns/ddns/container"
	"github.com/inimei/ddns/errs"
	"github.com/miekg/dns"
)

const (
	nodata = "nodata"
)

func init() {
	gob.Register((*dns.Msg)(nil))
	gob.Register((*dns.A)(nil))
	gob.Register((*dns.AAAA)(nil))
	gob.Register((*dns.CNAME)(nil))
	gob.Register((*dns.ANY)(nil))
	gob.Register((*dns.HINFO)(nil))
	gob.Register((*dns.MB)(nil))
	gob.Register((*dns.MG)(nil))
	gob.Register((*dns.MINFO)(nil))
	gob.Register((*dns.MR)(nil))
	gob.Register((*dns.MF)(nil))
	gob.Register((*dns.MD)(nil))
	gob.Register((*dns.MX)(nil))
	gob.Register((*dns.AFSDB)(nil))
	gob.Register((*dns.X25)(nil))
	gob.Register((*dns.RT)(nil))
	gob.Register((*dns.NS)(nil))
	gob.Register((*dns.PTR)(nil))
	gob.Register((*dns.RP)(nil))
	gob.Register((*dns.SOA)(nil))
	gob.Register((*dns.TXT)(nil))
	gob.Register((*dns.SPF)(nil))
	gob.Register((*dns.SRV)(nil))
	gob.Register((*dns.NAPTR)(nil))
	gob.Register((*dns.CERT)(nil))
	gob.Register((*dns.DNAME)(nil))
	gob.Register((*dns.PX)(nil))
	gob.Register((*dns.GPOS)(nil))
	gob.Register((*dns.LOC)(nil))
	gob.Register((*dns.SIG)(nil))
	gob.Register((*dns.RRSIG)(nil))
	gob.Register((*dns.NSEC)(nil))
	gob.Register((*dns.DLV)(nil))
	gob.Register((*dns.CDS)(nil))
	gob.Register((*dns.DS)(nil))
	gob.Register((*dns.KX)(nil))
	gob.Register((*dns.TA)(nil))
	gob.Register((*dns.TALINK)(nil))
	gob.Register((*dns.SSHFP)(nil))
	gob.Register((*dns.KEY)(nil))
	gob.Register((*dns.CDNSKEY)(nil))
	gob.Register((*dns.DNSKEY)(nil))
	gob.Register((*dns.RKEY)(nil))
	gob.Register((*dns.NSAPPTR)(nil))
	gob.Register((*dns.NSEC3)(nil))
	gob.Register((*dns.NSEC3PARAM)(nil))
	gob.Register((*dns.TKEY)(nil))
	gob.Register((*dns.RFC3597)(nil))
	gob.Register((*dns.URI)(nil))
	gob.Register((*dns.DHCID)(nil))
	gob.Register((*dns.TLSA)(nil))
	gob.Register((*dns.SMIMEA)(nil))
	gob.Register((*dns.HIP)(nil))
	gob.Register((*dns.NINFO)(nil))
	gob.Register((*dns.NID)(nil))
	gob.Register((*dns.L32)(nil))
	gob.Register((*dns.L64)(nil))
	gob.Register((*dns.LP)(nil))
	gob.Register((*dns.EUI48)(nil))
	gob.Register((*dns.EUI64)(nil))
	gob.Register((*dns.CAA)(nil))
	gob.Register((*dns.UID)(nil))
	gob.Register((*dns.GID)(nil))
	gob.Register((*dns.UINFO)(nil))
	gob.Register((*dns.EID)(nil))
	gob.Register((*dns.NIMLOC)(nil))
	gob.Register((*dns.OPENPGPKEY)(nil))
}

type Cache interface {
	Get(key string) (*dns.Msg, error)
	Set(key string, Msg *dns.Msg) error
	Exists(key string) bool
	Remove(key string)
	Length() int
}

func NewRedisCache(key string, db int, defExpire int) (Cache, error) {
	cfg := `{ "key":"%s", "dbNum":"%v", "password": "%s", "conn":"%s", "scancount": "500"}`

	ccfg := fmt.Sprintf(cfg, key, db, config.Data.Redis.Passwd, config.Data.Redis.Host)
	c, err := container.NewContainer(container.CTRedis, ccfg)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return &rCache{c, time.Duration(defExpire) * time.Second}, nil
}

func NewMemCache(defExpire int) (Cache, error) {
	c, _ := container.NewContainer(container.CTMemery, "")
	return &rCache{c, time.Duration(defExpire) * time.Second}, nil
}

type rCache struct {
	Backend container.Container
	Expire  time.Duration
}

func (c *rCache) Get(key string) (*dns.Msg, error) {

	data, err := c.Backend.Get(key)
	if err != nil {
		return nil, err
	}

	if data == nodata {
		return nil, nil
	}

	if len(data) == 0 {
		return nil, errs.ErrKeyNotFound
	}

	msg, err := c.decode(data)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return msg, nil
}

func (c *rCache) Set(key string, msg *dns.Msg) error {

	var expire int64
	if msg != nil && len(msg.Answer) > 0 && msg.Answer[0].Header() != nil {
		expire = int64(msg.Answer[0].Header().Ttl)
	} else {
		expire = int64(c.Expire / time.Second)
	}

	var err error
	val := nodata
	if msg != nil {
		val, err = c.encode(msg)
		if err != nil {
			return err
		}
	}

	err = c.Backend.Set(key, val, expire)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (c *rCache) Remove(key string) {
	err := c.Backend.Delete(key)
	if err != nil {
		log.Error(err.Error())
	}
}

func (c *rCache) Exists(key string) bool {
	return c.Backend.IsExist(key)
}

func (c *rCache) Length() int {
	len, err := c.Backend.Count()
	if err != nil {
		log.Error(err.Error())
		return 0
	}
	return int(len)
}

func (c *rCache) encode(msg *dns.Msg) (string, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(msg)
	if err != nil {
		log.Error("encode msg [%v] failed: %v", msg, err)
		return "", err
	}

	data := buf.Bytes()
	encoded := make([]byte, base64.URLEncoding.EncodedLen(len(data)))
	base64.URLEncoding.Encode(encoded, data)
	return string(encoded), nil
}

func (c *rCache) decode(msg string) (*dns.Msg, error) {
	value := []byte(msg)
	decoded := make([]byte, base64.URLEncoding.DecodedLen(len(value)))
	b, err := base64.URLEncoding.Decode(decoded, value)
	if err != nil {
		log.Error("decode msg failed:%v", err)
		return nil, err
	}

	value = decoded[:b]
	buf := bytes.NewBuffer(value)
	dec := gob.NewDecoder(buf)
	var out *dns.Msg
	err = dec.Decode(&out)
	if err != nil {
		log.Error("decode msg failed:%v", err)
		return nil, err
	}
	return out, nil
}
