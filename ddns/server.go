package ddns

import (
	"strconv"
	"time"

	"github.com/inimei/backup/log"
	"github.com/miekg/dns"
)

type Server struct {
	Host     string
	Port     int
	RTimeout time.Duration
	WTimeout time.Duration
}

func (s *Server) Addr() string {
	if len(s.Host) == 0 {
		return ""
	}
	return s.Host + ":" + strconv.Itoa(s.Port)
}

func (s *Server) Run() {

	Handler := NewHandler()

	tcpHandler := dns.NewServeMux()
	tcpHandler.HandleFunc(".", Handler.DoTCP)

	udpHandler := dns.NewServeMux()
	udpHandler.HandleFunc(".", Handler.DoUDP)

	tcpServer := &dns.Server{Addr: s.Addr(),
		Net:          "tcp",
		Handler:      tcpHandler,
		ReadTimeout:  s.RTimeout,
		WriteTimeout: s.WTimeout}

	udpServer := &dns.Server{Addr: s.Addr(),
		Net:          "udp",
		Handler:      udpHandler,
		UDPSize:      65535,
		ReadTimeout:  s.RTimeout,
		WriteTimeout: s.WTimeout}

	go s.start(udpServer)
	go s.start(tcpServer)
}

func (s *Server) start(ds *dns.Server) {

	log.Info("Start %s listener on %s", ds.Net, s.Addr())
	err := ds.ListenAndServe()
	if err != nil {
		log.Error("Start %s listener on %s failed:%s", ds.Net, s.Addr(), err.Error())
	}
}
