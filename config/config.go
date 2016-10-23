package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/inimei/backup/log"
)

type SqlConfig struct {
	Username     string `toml:"username"`
	Password     string `toml:"password"`
	URL          string `toml:"host"`
	Port         string `toml:"port"`
	DatabaseName string `toml:"database"`
}

type CacheSettings struct {
	Backend  string `toml:"backend"`
	Expire   int    `toml:"expire"`
	Maxcount int    `toml:"maxcount"`
}

type ResolvSettings struct {
	Enable     bool   `toml:"enable"`
	ResolvFile string `toml:"resolv-file"`
	Timeout    int
	Interval   int
}

type HostsSettings struct {
	Enable          bool
	HostsFile       string `toml:"host-file"`
	RedisEnable     bool   `toml:"redis-enable"`
	RedisKey        string `toml:"redis-key"`
	TTL             uint32 `toml:"ttl"`
	RefreshInterval uint32 `toml:"refresh-interval"`
}

type RedisSettings struct {
	Host     string
	Port     int
	DB       int
	Password string
}

func (s RedisSettings) Addr() string {
	return s.Host + ":" + strconv.Itoa(s.Port)
}

type cfgData struct {
	Server struct {
		Debug bool   `toml:"debug"`
		Addr  string `toml:"addr"`
		Port  int    `toml:"port"`
	} `toml:"server"`

	Sql    SqlConfig      `toml:"mysql"`
	Cache  CacheSettings  `toml:"cache"`
	Resolv ResolvSettings `toml:"resolv"`
	Hosts  HostsSettings  `toml:"hosts"`
	Redis  RedisSettings  `toml:"redis"`
}

var Data cfgData

func initialize(configFilePath string) error {
	if configFilePath == "" {
		configFilePath = "ddns.toml"
	}

	f, err := os.Open(configFilePath)
	if err != nil {
		cwd, _ := os.Getwd()
		log.Error("os.Stat fail, %s ,please ensure ddns.toml exist. ddns.toml path:%s, cwd:%s", err, configFilePath, cwd)
		return err
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		log.Error("read config file error, %s", err)
		return err
	}

	if err := toml.Unmarshal(buf, &Data); err != nil {
		log.Error("unmarshal config failed, %s", err)
		return err
	}

	return nil
}

// CurDir current dir
func CurDir() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	return dir, nil
}

func init() {
	dir, _ := CurDir()

	err := initialize(dir + "/ddns.toml")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
