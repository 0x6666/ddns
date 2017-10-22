package controllers

import (
	"crypto/sha1"
	"net/http"

	"fmt"

	"github.com/0x6666/backup/log"
	"github.com/0x6666/ddns/app/model"
	"github.com/0x6666/ddns/app/sessions"
	"github.com/0x6666/ddns/config"
)

const (
	PRoot = "/"

	PFiles = "/files"

	PRecodes = "/domain/:did/recodes"
	PRecode  = "/recode/:rid"
	PLogin   = "/login"
	PLogout  = "/logout"
	PAbout   = "/about"

	PDomains = "/domains"
	PDomain  = "/domain/:did"

	PUpdate = "/update"

	PDownloads   = "/downloads"
	PDownloadDel = "/download/del"

	PRRecodes    = "/api/recodes"
	PDataversion = "/api/dataversion"
	PApiUpdate   = "/api/update"
)

type AppCtrl struct {
	ContrllerBase
}

//Root ...
func (a AppCtrl) Root() {
	a.C.Redirect(http.StatusFound, PDomains)
}

// GetLogin -> [GET] :/login
//
// Ret Code:[200]
//
func (a AppCtrl) GetLogin() {
	if sessions.IsLogined(a.C.Request) {
		to := a.C.Query("to")
		if len(to) == 0 {
			to = PRoot
		}
		a.C.Redirect(http.StatusPermanentRedirect, to)
		return
	}

	parameters := a.getTemplateParameter()
	a.C.HTML(http.StatusOK, "login.html", parameters)
	return
}

// GetAbout -> [GET] :/about
//
// Ret Code:[200]
//
//
func (a AppCtrl) GetAbout() {
	parameters := a.getTemplateParameter()

	parameters["BreadcrumbSecs"] = SectionItems{
		&SectionItem{"About", "/about"},
	}
	parameters["View"] = "view_about"
	a.HTML(http.StatusOK, parameters)
	return
}

// Login -> [POST] :/login?to=xxxx
//
// Ret Code:[200]
//
// 成功返回值
//	{
//		"code": "OK"
//	}
//
// 失败返回值
//		code: xxx
//
func (a AppCtrl) Login() {
	username := a.C.PostForm("username")
	password := a.C.PostForm("password")

	if username != config.Data.Web.Admin {
		a.rspErrorCode(CodeUserNameError, "username error")
		return
	}

	if fmt.Sprintf("%x", sha1.Sum([]byte(password))) != config.Data.Web.Passwd {
		a.rspErrorCode(CodePasswordError, "password error")
		return
	}

	sessions.Login(a.C.Writer, a.C.Request, model.DefUserID)

	a.rspOk()
}

// Logout -> [POST] :/logout
//
// Ret Code:[200]
//
// 成功返回值
//	{
//		"code": "OK"
//	}
//
// 失败返回值
//		code: xxx
//
func (a AppCtrl) Logout() {
	b := sessions.IsLogined(a.C.Request)
	if !b {
		a.JSON(http.StatusOK, JsonMap{"code": CodeInvalidSession})
		return
	}

	err := sessions.Logout(a.C.Writer, a.C.Request)
	if err != nil {
		log.Error(err.Error())
		a.JSON(http.StatusOK, JsonMap{"code": CodeUnknowError})
		return
	}
	a.JSON(http.StatusOK, JsonMap{"code": CodeOK})
}
