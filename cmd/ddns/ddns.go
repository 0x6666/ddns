package main

import (
	"os"
	"os/signal"
	"runtime/pprof"
	"time"

	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/config"
	"github.com/inimei/ddns/data"
	"github.com/inimei/ddns/data/sqlite"
	"github.com/inimei/ddns/ddns"
	"github.com/inimei/ddns/web"
)

func main() {

	log.SetLevel(log.LevelInfo | log.LevelDebug | log.LevelWarn | log.LevelError)
	defer log.Close()

	var db data.IDatabase
	db = sqlite.NewSqlite()
	if err := db.Init(); err != nil {
		log.Error(err.Error())
		return
	}

	var server *ddns.Server
	if config.Data.Server.EnableDNS {
		server = &ddns.Server{
			Host:     config.Data.Server.Addr,
			Port:     config.Data.Server.Port,
			RTimeout: 5 * time.Second,
			WTimeout: 5 * time.Second,
			Db:       db,
		}
		server.Run()
	}

	var ws *web.WebServer
	if config.Data.Server.EnableWeb {
		ws = &web.WebServer{}
		ws.Start(db)
	}

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

			if ws != nil {
				ws.Stop()
			}

			if server != nil {
				server.Stop()
			}

			if db != nil {
				db.Close()
			}

			break forever
		}
	}
}

func profileCPU() {
	f, err := os.Create("ddns.cprof")
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
	f, err := os.Create("ddns.mprof")
	if err != nil {
		log.Error("%v", err)
		return
	}

	time.AfterFunc(5*time.Minute, func() {
		pprof.WriteHeapProfile(f)
		f.Close()
	})
}
