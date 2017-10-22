package app

import (
	"github.com/0x6666/ddns/app/controllers"
	"github.com/0x6666/ddns/config"
	"github.com/0x6666/ddns/server"
)

func RegRoute(r server.IRouter) {
	curPath := config.CurDir()
	curPath += "/ddns_static"

	r.Public("/login", curPath+"/login")
	r.Public("/assets", curPath+"/assets")
	if config.Data.Download.Enable {
		r.Public(controllers.PFiles, config.Data.Download.Dest)

		g := r.Group("", controllers.Downloader.CookieAuthMiddleware)
		g.Get(controllers.PDownloads, controllers.Downloader.GetDownloads)
		g.Post(controllers.PDownloads, controllers.Downloader.StartDownloads)
		g.Post(controllers.PDownloadDel, controllers.Downloader.Delete)
	}

	//api
	apig := r.Group("", controllers.ApiCtrl.SignMiddleware)
	apig.Get(controllers.PRRecodes, controllers.ApiCtrl.ApiGetRecodes)
	apig.Get(controllers.PDataversion, controllers.ApiCtrl.ApiGetDataSchemaVersion)
	apig.Post(controllers.PApiUpdate, controllers.ApiCtrl.ApiUpdateRecode)

	//app
	dg := r.Group("")
	dg.Post(controllers.PLogin, controllers.AppCtrl.Login)
	dg.Get(controllers.PLogin, controllers.AppCtrl.GetLogin)
	dg.Get(controllers.PAbout, controllers.AppCtrl.GetAbout)
	dg.Post(controllers.PLogout, controllers.AppCtrl.Logout)
	dg.Get(controllers.PRoot, controllers.AppCtrl.Root)

	//domains
	adg := dg.Group("", controllers.DomainsCtrl.CookieAuthMiddleware)

	adg.Get(controllers.PDomains, controllers.DomainsCtrl.GetDomains)
	adg.Post(controllers.PDomains, controllers.DomainsCtrl.NewDomain)
	adg.Get(controllers.PDomain, controllers.DomainsCtrl.GetDomain)
	adg.Patch(controllers.PDomain, controllers.DomainsCtrl.UpdateDomain)
	adg.Delete(controllers.PDomain, controllers.DomainsCtrl.DeleteDomain)
	adg.Get(controllers.PRecodes, controllers.DomainsCtrl.GetRecodes)
	adg.Post(controllers.PRecodes, controllers.DomainsCtrl.NewRecode)
	adg.Get(controllers.PRecode, controllers.DomainsCtrl.GetRecode)
	adg.Post(controllers.PRecode, controllers.DomainsCtrl.UpdateRecode)
	adg.Patch(controllers.PRecode, controllers.DomainsCtrl.UpdateRecode)
	adg.Delete(controllers.PRecode, controllers.DomainsCtrl.DeleteRecode)
}
