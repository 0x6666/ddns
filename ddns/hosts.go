package ddns

import (
	"bufio"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/miekg/dns"

	"github.com/0x6666/util/log"
)

type Hosts struct {
	fileHosts *FileHosts

	hostWatcher *fsnotify.Watcher
	c1          chan bool
}

func NewHosts() Hosts {
	fileHosts := &FileHosts{
		file:  "./hosts",
		hosts: make(map[string][]string),
	}

	hosts := Hosts{fileHosts, nil, make(chan bool, 1)}
	hosts.refresh()
	return hosts
}

func (h *Hosts) Get(domain string, qtype uint16) ([]net.IP, bool) {

	var sips []string
	var ips []net.IP

	sips, _ = h.fileHosts.Get(domain)

	if sips == nil {
		return nil, false
	}

	for _, sip := range sips {
		var ip net.IP = net.ParseIP(sip)
		switch {
		case ip == nil:
			log.Error("invalid ip address [%v]", sip)
			return []net.IP{ip}, true
		case qtype == dns.TypeA:
			ip = ip.To4()
		case qtype == dns.TypeAAAA:
			if ip.To4() == nil {
				ip = ip.To16()
			} else {
				ip = nil
			}
		}
		if ip != nil {
			ips = append(ips, ip)
		}
	}

	return ips, (ips != nil)
}

func (h *Hosts) refresh() {

	h.fileHosts.Refresh()

	if h.hostWatcher != nil {
		return
	}

	var err error
	h.hostWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Error(err.Error())
	}

	go func() {
	stopWatch:
		for {
			select {
			case event := <-h.hostWatcher.Events:
				log.Info("event: %v", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Info("modified file: %v", event.Name)
					h.fileHosts.Refresh()
				}
			case err := <-h.hostWatcher.Errors:
				if err != nil {
					log.Info("error: %v", err)

				}
			case <-h.c1:
				break stopWatch
			}
		}
	}()

	err = h.hostWatcher.Add(h.fileHosts.file)
	if err != nil {
		log.Error(err.Error())
	}
}

type FileHosts struct {
	file  string
	hosts map[string][]string
	mu    sync.RWMutex
}

func (f *FileHosts) Get(domain string) ([]string, bool) {
	domain = strings.ToLower(domain)
	f.mu.RLock()
	ip, ok := f.hosts[domain]
	f.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return ip, true
}

func (f *FileHosts) Refresh() {
	buf, err := os.Open(f.file)
	if err != nil {
		log.Warn("Update hosts records from file failed %s", err)
		return
	}
	defer buf.Close()

	f.mu.Lock()
	defer f.mu.Unlock()

	f.clear()

	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {

		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		sli := strings.Split(line, " ")
		if len(sli) == 1 {
			sli = strings.Split(line, "\t")
		}

		if len(sli) < 2 {
			continue
		}

		domain := sli[len(sli)-1]
		ip := sli[0]
		if !f.isDomain(domain) || !f.isIP(ip) {
			continue
		}

		f.hosts[strings.ToLower(domain)] = append(f.hosts[strings.ToLower(domain)], ip)
	}
	log.Debug("update hosts records from %s", f.file)
}

func (f *FileHosts) clear() {
	f.hosts = make(map[string][]string)
}

func (f *FileHosts) isDomain(domain string) bool {
	if f.isIP(domain) {
		return false
	}
	match, _ := regexp.MatchString(`^([a-zA-Z0-9\*]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,6}$`, domain)
	return match
}

func (f *FileHosts) isIP(ip string) bool {
	return (net.ParseIP(ip) != nil)
}
