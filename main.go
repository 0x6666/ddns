package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/0x6666/ddnsx/ddns"
	"github.com/0x6666/util/log"
)

func main() {

	log.SetLevel(log.LevelAll)
	defer log.Close()

	svr := ddns.Server{
		Host:     "",
		Port:     53,
		RTimeout: 5 * time.Second,
		WTimeout: 5 * time.Second,
	}
	svr.Host = ""
	svr.Port = 53
	svr.Run()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

forever:
	for {
		select {
		case <-sig:
			log.Debug("signal received, stopping")
			// stop dns
			svr.Stop()
			break forever
		}
	}
}
