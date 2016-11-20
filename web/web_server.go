package web

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/inimei/ddns/config"
	"github.com/inimei/ddns/data"
)

type WebServer struct {
	e    *gin.Engine
	h    *handler
	tmpl *tmplHelper

	db data.IDatabase
}

func (ws *WebServer) Start(db data.IDatabase) {

	ws.db = db
	ws.e = gin.Default()

	if !config.Data.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	curPath := config.CurDir()
	curPath += "/ddns_static"

	ws.e.Static("/html", curPath+"/html")
	ws.e.Static("/css", curPath+"/css")
	ws.e.Static("/js", curPath+"/js")
	ws.e.Static("/vendors", curPath+"/vendors")

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
	ws.h = &handler{ws: ws}

	ws.e.GET(pRoot, ws.h.root)

	ws.e.GET(pRecodes, ws.h.getRecodes)
	ws.e.POST(pRecodes, ws.h.newRecode)
	ws.e.GET(pRecode, ws.h.getRecode)
	ws.e.POST(pRecode, ws.h.getRecode)
	ws.e.PATCH(pRecode, ws.h.getRecode)
	ws.e.DELETE(pRecode, ws.h.getRecode)
}
