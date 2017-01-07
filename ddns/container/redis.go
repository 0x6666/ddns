package container

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

type RedisContainer struct {
	p        *redis.Pool
	conninfo string
	dbNum    int
	name     string
	password string

	scancount int
	cursor    int
	eof       bool
}

func newRedisContainer() *RedisContainer {
	return &RedisContainer{}
}

func (rc *RedisContainer) do(commandName string, args ...interface{}) (reply interface{}, err error) {
	c := rc.p.Get()
	defer c.Close()
	return c.Do(commandName, args...)
}

//Get 。。。
func (rc *RedisContainer) Get(key string) (string, error) {
	if !rc.IsExist(key) {
		return "", ContainerErr{param: key, t: KeyNotFound}
	}
	return redis.String(rc.do("GET", key))
}

// Delete cache in redis.
func (rc *RedisContainer) Delete(key string) error {
	var err error
	if _, err = rc.do("DEL", key); err != nil {
		return err
	}
	_, err = rc.do("HDEL", rc.name, key)
	return err
}

/*
//Append 。。。
func (rc *RedisContainer) Append(key, val string) error {
	if _, err := rc.do("SADD", key, val); err != nil {
		return err
	}
	if _, err := rc.do("HSET", rc.name, key, true); err != nil {
		return err
	}
	return nil
}
*/
// Set ...
func (rc *RedisContainer) Set(key string, val string, expire int64) error {
	if _, err := rc.do("SETEX", key, expire, val); err != nil {
		return err
	}
	if _, err := rc.do("HSET", rc.name, key, true); err != nil {
		return err
	}
	return nil
}

/*
//DeleteValue 。。。
func (rc *RedisContainer) DeleteValue(key, val string) error {
	if _, err := rc.do("SREM", key, val); err != nil {
		return err
	}
	return nil
}

//ValueCount 。。。
func (rc *RedisContainer) ValueCount(key string) (int, error) {
	return redis.Int(rc.do("SCARD", key))
}
*/

//IsExist check cache's existence in redis.
func (rc *RedisContainer) IsExist(key string) bool {
	v, err := redis.Bool(rc.do("EXISTS", key))
	if err != nil {
		return false
	}
	if v == false {
		rc.do("HDEL", rc.name, key)
	} else {
		v, err = redis.Bool(rc.do("HEXISTS", rc.name, key))
		if err != nil {
			return false
		}
	}
	return v
}

//Count ...
func (rc *RedisContainer) Count() (int64, error) {
	return redis.Int64(rc.do("HLEN", rc.name))
}

/*
//IsValueExist ...
func (rc *RedisContainer) IsValueExist(key, val string) bool {
	b, err := redis.Bool(rc.do("SISMEMBER", key, val))
	if err != nil {
		return false
	}
	return b
}
*/

//ClearAll clean all cache in redis. delete this redis collection.
func (rc *RedisContainer) ClearAll() error {
	//卧槽， HKEYS不能用在生产环境。。。。。
	cachedKeys, err := redis.Strings(rc.do("HKEYS", rc.name))
	if err != nil {
		return err
	}
	for _, str := range cachedKeys {
		if _, err = rc.do("DEL", str); err != nil {
			return err
		}
	}
	_, err = rc.do("DEL", rc.name)
	return err
}

//StartScan 。。。
func (rc *RedisContainer) StartScan() {
	rc.cursor = 0
	rc.eof = false
}

//ScanNext ...
func (rc *RedisContainer) ScanNext() ([]string, error) {
	if rc.eof {
		return nil, errors.New("Scan EOF")
	}

	res, err := redis.Values(rc.do("HSCAN", rc.name, rc.cursor, "COUNT", rc.scancount))
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, errors.New("HSCAN failed")
	}

	rc.cursor, err = redis.Int(res[0], nil)
	if err != nil {
		return nil, err
	}
	rc.eof = (rc.cursor == 0)

	if len(res) > 1 {
		kes, err := redis.Strings(res[1], nil)
		if err != nil {
			return nil, err
		}
		return kes, nil
	}

	return nil, nil
}

func (rc *RedisContainer) Eof() bool {
	return rc.eof
}

//StartAndGC start redis Counter adapter.
// config is like {"key":"collection key","conn":"connection info","dbNum":"0", "password", ""}
// the Counter item in redis are stored forever,
// so no gc operation.
func (rc *RedisContainer) start(config string) error {
	var cf map[string]string
	err := json.Unmarshal([]byte(config), &cf)
	if err != nil {
		return err
	}

	if _, ok := cf["conn"]; !ok {
		return errors.New("config has no conn key")
	}
	if _, ok := cf["key"]; !ok {
		cf["key"] = "DDNS-Resis"
	}
	if _, ok := cf["dbNum"]; !ok {
		cf["dbNum"] = "0"
	}
	if _, ok := cf["password"]; !ok {
		cf["password"] = ""
	}
	if c, ok := cf["scancount"]; !ok {
		rc.scancount = 20
	} else {
		rc.scancount, err = strconv.Atoi(c)
		if err != nil {
			return err
		}
	}

	rc.name = cf["key"]
	rc.conninfo = cf["conn"]
	rc.dbNum, _ = strconv.Atoi(cf["dbNum"])
	rc.password = cf["password"]

	rc.connectInit()

	c := rc.p.Get()
	defer c.Close()

	return c.Err()
}

// connect to redis.
func (rc *RedisContainer) connectInit() {
	dialFunc := func() (c redis.Conn, err error) {
		c, err = redis.Dial("tcp", rc.conninfo)
		if err != nil {
			return nil, err
		}

		if rc.password != "" {
			if _, err := c.Do("AUTH", rc.password); err != nil {
				c.Close()
				return nil, err
			}
		}

		_, selecterr := c.Do("SELECT", rc.dbNum)
		if selecterr != nil {
			c.Close()
			return nil, selecterr
		}
		return
	}
	// initialize a new pool
	rc.p = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 180 * time.Second,
		Dial:        dialFunc,
	}
}
