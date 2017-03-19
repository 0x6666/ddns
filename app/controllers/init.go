package controllers

import "github.com/yangsongfwd/ddns/server"

func init() {
	server.RegisterController((*AppCtrl)(nil))
	server.RegisterController((*Downloader)(nil))
	server.RegisterController((*ApiCtrl)(nil))
	server.RegisterController((*DomainsCtrl)(nil))
}
