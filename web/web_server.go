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

	ws.regWebHandler()
	ws.regAPIHandler()
}

func (ws *WebServer) regWebHandler() {

	group := ws.e.Group("")
	group.GET(pRoot, ws.h.root)

	group.GET(pRecodes, ws.h.getRecodes)
	group.POST(pRecodes, ws.h.newRecode)
	group.GET(pRecode, ws.h.getRecode)
	group.POST(pRecode, ws.h.getRecode)
	group.PATCH(pRecode, ws.h.getRecode)
	group.DELETE(pRecode, ws.h.deleteRecode)

	group.POST(pUpdate, ws.h.updateRecode)
}

func (ws *WebServer) regAPIHandler() {
	group := ws.e.Group("/api", ws.h.SignMiddleware)

	group.GET("/recodes", ws.h.apiGetRecodes)
	group.GET("/schemaversion", ws.h.apiGetDataSchemaVersion)
}
