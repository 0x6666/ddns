package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gpmgo/gopm/modules/log"
	"github.com/yangsongfwd/ddns/app/model"
	"github.com/yangsongfwd/ddns/config"
	"github.com/yangsongfwd/ddns/server"
)

const (
	CodeOK                 = "ok"
	CodeInvalidSession     = "InvalidSession"
	CodeDBError            = "DBError"
	CodeUnknowError        = "UnknowError"
	CodeInvalidParam       = "InvalidParam"
	CodeRecodeExist        = "RecodeExist"
	CodeKeyIsEmpty         = "KeyIsEmpty"
	CodeInvalidIP          = "InvalidIP"
	CodeNoAuthorization    = "NoAuthorization"
	CodeAuthorizationError = "AuthorizationError"
	CodeVerifySignature    = "VerifySignature"
	CodeGetSecretKeyError  = "GetSecretKeyError"
	CodeUserNameError      = "UserNameError"
	CodePasswordError      = "PasswordError"
	CodeInvalidURL         = "InvalidURL"
)

type SectionItem struct {
	Name string
	Href string
}

type JsonMap map[string]interface{}

type SectionItems []*SectionItem

type ContrllerBase struct {
	*server.CtrlBase
}

func (c ContrllerBase) db() model.IDatabase {
	idb := server.Server.GetGlobalData("db")
	if idb == nil {
		log.Error("get db error")
		return nil
	}
	return idb.(model.IDatabase)
}

func (c ContrllerBase) getTemplateParameter() JsonMap {
	envData := JsonMap{}
	if gin.IsDebugging() {
		envData["Debug"] = true
	} else {
		envData["Production"] = true
	}
	envData["AssetsHost"] = config.Data.Web.AssetsImageHost

	parameters := JsonMap{}
	parameters["Env"] = envData
	parameters["Layout"] = "app_layout.html"

	if strings.ToLower(c.C.Request.Header.Get("DDNS-View")) == "true" {
		parameters["Layout"] = "app_view.html"
	}

	val, exist := c.C.Get(MwUserid)
	if exist {
		//TODO: impl user
		parameters["CurUser"] = JsonMap{"UserID": val, "UserName": "YangSong"}
	}

	return parameters
}

func (c ContrllerBase) HTML(code int, params JsonMap) {
	c.C.HTML(code, fmt.Sprintf("%v", params["Layout"]), params)
}

func (c ContrllerBase) JSON(code int, data JsonMap) {
	c.C.JSON(code, data)
}

func (c ContrllerBase) rspOk() {
	c.C.JSON(http.StatusOK, gin.H{"code": CodeOK})
}

func (c ContrllerBase) rspError(err error) {
	c.C.JSON(http.StatusOK, gin.H{
		"code": CodeUnknowError,
		"msg":  err.Error(),
	})
}

func (c ContrllerBase) rspErrorCode(code, msg string) {
	c.C.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  msg,
	})
}
