package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/inimei/backup/log"
)

type CacheSettings struct {
	Backend  string `toml:"backend"`
	Expire   int    `toml:"expire"`
	Maxcount int    `toml:"maxcount"`
}

type RedisSettings struct {
	Host   string `toml:"host"`
	Passwd string `toml:"passwd"`
}

type ResolvSettings struct {
	Enable     bool   `toml:"enable"`
	ResolvFile string `toml:"resolv-file"`
	Timeout    int
	Interval   int
}

type HostsSettings struct {
	Enable    bool
	HostsFile string `toml:"host-file"`
	TTL       uint32 `toml:"ttl"`
}

type ServerSetting struct {
	Debug     bool   `toml:"debug"`
	Addr      string `toml:"addr"`
	Port      int    `toml:"port"`
	EnableWeb bool   `toml:"enableweb"`
	EnableDNS bool   `toml:"enabledns"`
	Master    bool   `toml:"master"`
}

type SlaveSetting struct {
	MasterHost string `toml:"master_host"`
	Accesskey  string `toml:"accesskey"`
	SecretKey  string `toml:"secretKey"`
	UpdateTime int64  `toml:"update_time"`
}

type WebSetting struct {
	Port            int    `toml:"port"`
	Admin           string `toml:"admin"`
	Passwd          string `toml:"passwd"`
	AssetsImageHost string `toml:"assets_image_host"`
}

type SqliteSetting struct {
	Path string `toml:"path"`
}

type cfgData struct {
	Server ServerSetting  `toml:"server"`
	Slave  SlaveSetting   `toml:"slave"`
	Web    WebSetting     `toml:"web"`
	Sqlite SqliteSetting  `toml:"sqlite"`
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

	if Data.Server.Master == false {
		if strings.HasSuffix(Data.Slave.MasterHost, "/") {
			Data.Slave.MasterHost = strings.TrimSuffix(Data.Slave.MasterHost, "/")
		}
		if Data.Slave.UpdateTime == 0 {
			Data.Slave.UpdateTime = 300
		}
	}

	if strings.HasSuffix(Data.Web.AssetsImageHost, "/") {
		Data.Web.AssetsImageHost = strings.TrimSuffix(Data.Web.AssetsImageHost, "/")
	}

	return nil
}

// CurDir current dir
func CurDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return dir
}

func init() {
	dir := CurDir()

	err := initialize(dir + "/ddns.toml")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
