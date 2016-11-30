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

	ws.e.Static("/login", curPath+"/login")
	ws.e.Static("/assets", curPath+"/assets")

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
}

func (ws *WebServer) regWebHandler() {

	group := ws.e.Group("")

	group.POST(pLogin, ws.h.login)
	group.GET(pLogin, ws.h.getLogin)

	auth := group.Group("", ws.h.CookieAuthMiddleware)

	auth.GET(pRoot, ws.h.root)
	auth.GET(pRecodes, ws.h.getRecodes)
	auth.POST(pRecodes, ws.h.newRecode)
	auth.GET(pRecode, ws.h.getRecode)
	auth.POST(pRecode, ws.h.getRecode)
	auth.PATCH(pRecode, ws.h.getRecode)
	auth.DELETE(pRecode, ws.h.deleteRecode)
}

func (ws *WebServer) regAPIHandler() {
	group := ws.e.Group("/api", ws.h.SignMiddleware)

	group.GET("/recodes", ws.h.apiGetRecodes)
	group.GET("/dataversion", ws.h.apiGetDataSchemaVersion)
	group.POST("/update", ws.h.apiUpdateRecode)
}
