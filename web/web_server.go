package web

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/inimei/ddns/config"
	"github.com/inimei/ddns/data"
	"github.com/inimei/ddns/download"
)

type WebServer struct {
	e    *gin.Engine
	h    *handler
	d    *download.DownloadMgr
	tmpl *tmplHelper

	db data.IDatabase
}

func (ws *WebServer) Start(db data.IDatabase, d *download.DownloadMgr) {

	ws.db = db
	ws.e = gin.Default()
	ws.d = d

	if !config.Data.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	curPath := config.CurDir()
	curPath += "/ddns_static"

	ws.e.Static("/login", curPath+"/login")
	ws.e.Static("/assets", curPath+"/assets")
	if config.Data.Download.Enable {
		ws.e.Static(pFiles, config.Data.Download.Dest)
	}

	ws.loadTemplate()
	ws.regHandler()

	go func() {
		ws.e.Run(":" + fmt.Sprintf("%v", config.Data.Web.Port))
	}()
}

func (ws *WebServer) Stop() {

}

func (ws *WebServer) loadTemplate() {
	ws.tmpl = &tmplHelper{}
	ws.tmpl.e = ws.e

	curPath := config.CurDir()
	curPath += "/ddns_static"
	ws.tmpl.loadMainTmpl(curPath + "/tmpl/*.*")
	ws.tmpl.loadView(curPath + "/tmpl/views/*.*")
}

func (ws *WebServer) regHandler() {
	ws.h = newHandler(ws)

	ws.regWebHandler()
	ws.regAPIHandler()
	if config.Data.Download.Enable {
		ws.regDownload()
	}
}

func (ws *WebServer) regWebHandler() {

	group := ws.e.Group("")

	group.POST(pLogin, ws.h.login)
	group.GET(pLogin, ws.h.getLogin)
	group.GET(pAbout, ws.h.getAbout)
	group.POST(pLogout, ws.h.logout)
	group.GET(pRoot, ws.h.root)

	auth := group.Group("", ws.h.CookieAuthMiddleware)

	auth.GET(pDomains, ws.h.getDomains)
	auth.POST(pDomains, ws.h.newDomain)
	auth.GET(pDomain, ws.h.getDomain)
	auth.PATCH(pDomain, ws.h.updateDomain)
	auth.DELETE(pDomain, ws.h.deleteDomain)

	auth.GET(pRecodes, ws.h.getRecodes)
	auth.POST(pRecodes, ws.h.newRecode)
	auth.GET(pRecode, ws.h.getRecode)
	auth.POST(pRecode, ws.h.updateRecode)
	auth.PATCH(pRecode, ws.h.updateRecode)
	auth.DELETE(pRecode, ws.h.deleteRecode)
}

func (ws *WebServer) regAPIHandler() {
	group := ws.e.Group("/api", ws.h.SignMiddleware)

	group.GET("/recodes", ws.h.apiGetRecodes)
	group.GET("/dataversion", ws.h.apiGetDataSchemaVersion)
	group.POST("/update", ws.h.apiUpdateRecode)
}

func (ws *WebServer) regDownload() {
	auth := ws.e.Group("", ws.h.CookieAuthMiddleware)

	auth.GET(pDownloads, ws.h.getDownloads)
	auth.POST(pDownloads, ws.h.startDownloads)

}
