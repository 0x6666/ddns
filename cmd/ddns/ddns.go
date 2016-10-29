package main

import (
	"os"
	"os/signal"
	"runtime/pprof"
	"time"

	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/config"
	"github.com/inimei/ddns/ddns"
)

func main() {

	log.SetLevel(log.LevelInfo | log.LevelDebug | log.LevelWarn | log.LevelError)
	defer log.Close()

	server := &ddns.Server{
		Host:     config.Data.Server.Addr,
		Port:     config.Data.Server.Port,
		RTimeout: 5 * time.Second,
		WTimeout: 5 * time.Second,
	}

	server.Run()

	if config.Data.Server.Debug {
		go profileCPU()
		go profileMEM()
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

forever:
	for {
		select {
		case <-sig:
			log.Debug("signal received, stopping")
			server.Stop()
			break forever
		}
	}
}

func profileCPU() {
	f, err := os.Create("godns.cprof")
	if err != nil {
		log.Error("%v", err)
		return
	}

	pprof.StartCPUProfile(f)
	time.AfterFunc(6*time.Minute, func() {
		pprof.StopCPUProfile()
		f.Close()
	})
}

func profileMEM() {
	f, err := os.Create("godns.mprof")
	if err != nil {
		log.Error("%v", err)
		return
	}

	time.AfterFunc(5*time.Minute, func() {
		pprof.WriteHeapProfile(f)
		f.Close()
	})

}
