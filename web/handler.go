package web

import (
	"crypto/sha1"
	"net/http"
	"strings"

	"strconv"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/config"
	"github.com/inimei/ddns/data/model"
	"github.com/inimei/ddns/errs"
	"github.com/inimei/ddns/web/sessions"
)

const (
	pRoot = "/"

	pRecodes = "/recodes"
	pRecode  = "/recode/:id"
	pLogin   = "/login"
	pAbout   = "/about"

	pUpdate = "/update"
)

const (
	MIMEJSON              = "application/json"
	MIMEHTML              = "text/html"
	MIMEXML               = "application/xml"
	MIMEXML2              = "text/xml"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
	MIMEPROTOBUF          = "application/x-protobuf"
)

const (
	CodeOK                 = "ok"
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
)

func requestType(r *http.Request) string {
	accept := r.Header["Accept"]
	if len(accept) == 0 {
		return MIMEHTML
	}

	accept = strings.Split(accept[0], ",")

	return accept[0]
}

func createRecodeFromForm(c *gin.Context) *model.Recode {

	r := &model.Recode{}
	r.RecordName = c.PostForm("name")
	r.RecodeValue = c.PostForm("value")
	r.RecordType = 1
	if len(r.RecodeValue) == 0 {
		r.Dynamic = true
	}

	r.TTL, _ = strconv.Atoi(c.DefaultPostForm("ttl", "600"))
	if r.TTL == 0 {
		r.TTL = 600
	}

	return r
}

type handler struct {
	ws      *WebServer
	envData map[string]interface{}
}

func newHandler(ws *WebServer) *handler {
	h := &handler{ws: ws}

	h.envData = map[string]interface{}{}
	if gin.IsDebugging() {
		h.envData["Debug"] = true
	} else {
		h.envData["Production"] = true
	}

	h.envData["AssetsHost"] = config.Data.Web.AssetsImageHost

	return h
}

func (h *handler) rspOk(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": CodeOK,
	})
}

func (h *handler) rspError(c *gin.Context, err error) {
	c.JSON(http.StatusOK, gin.H{
		"code": CodeUnknowError,
		"msg":  err.Error(),
	})
}

func (h *handler) rspErrorCode(c *gin.Context, code, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  msg,
	})
}

func (h *handler) getTemplateParameter(c *gin.Context) map[string]interface{} {

	parameters := map[string]interface{}{}
	parameters["Env"] = h.envData
	parameters["Layout"] = "app_layout.html"

	if strings.ToLower(c.Request.Header.Get("DDNS-View")) == "true" {
		parameters["Layout"] = "app_view.html"
	}

	return parameters
}

func (h *handler) HTML(c *gin.Context, code int, params map[string]interface{}) {
	c.HTML(code, fmt.Sprintf("%v", params["Layout"]), params)
}

func (h *handler) root(c *gin.Context) {
	//c.Redirect(http.StatusFound, "/html/index.html")
	c.Redirect(http.StatusFound, pRecodes)
}

func (h *handler) getRecodeFromParam(c *gin.Context) (*model.Recode, error) {
	rid, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		err = fmt.Errorf("parse recode id failed: %v", err.Error())
		log.Error(err.Error())
		h.rspErrorCode(c, CodeInvalidParam, err.Error())
		return nil, errs.ErrInvalidParam
	}

	r, err := h.ws.db.GetRecode(rid)
	if err != nil {
		err = fmt.Errorf("get recode [%v] failed: %v", rid, err.Error())
		log.Error(err.Error())
		h.rspErrorCode(c, CodeDBError, err.Error())
		return nil, errs.ErrInvalidParam
	}
	return r, err
}

// getLogin -> [GET] :/login
//
// Ret Code:[200]
//
func (h *handler) getLogin(c *gin.Context) {
	if sessions.IsLogined(c.Request) {
		c.Redirect(http.StatusPermanentRedirect, pRoot)
		return
	}

	parameters := h.getTemplateParameter(c)
	c.HTML(http.StatusOK, "login.html", parameters)
	return
}

// getAbout -> [GET] :/about
//
// Ret Code:[200]
//
//
func (h *handler) getAbout(c *gin.Context) {
	parameters := h.getTemplateParameter(c)
	parameters["View"] = "view_about"
	h.HTML(c, http.StatusOK, parameters)
	return
}

// login -> [POST] :/login?to=xxxx
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
func (h *handler) login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username != config.Data.Web.Admin {
		h.rspErrorCode(c, CodeUserNameError, "username error")
		return
	}

	if fmt.Sprintf("%x", sha1.Sum([]byte(password))) != config.Data.Web.Passwd {
		h.rspErrorCode(c, CodePasswordError, "password error")
		return
	}

	sessions.Login(c.Writer, c.Request, username)

	h.rspOk(c)
}

// getRecode -> [POST] :/recodes
//
// Ret Code:[200]
//
// [request=json] 获取指定区间的记录数
//
// 成功返回值
//	{
//		"code": "OK",
//		"recodes": [
//			{
//				"id" : xxx,
//				"name": "xxxx",
//				"dynamic": true,
//				"ttl": xxx,
//				"key": "xxxx",
//				"value": "xxxx"
//			}
//		]
//	}
//
// 失败返回值
//		code: xxx
//
func (h *handler) getRecodes(c *gin.Context) {

	if t := requestType(c.Request); t != MIMEJSON {
		parameters := h.getTemplateParameter(c)
		parameters["View"] = "recode_list"
		h.HTML(c, http.StatusOK, parameters)
		return
	}

	h.apiGetRecodes(c)
}

// newRecode -> [POST] :/recodes
//
// Ret Code:[200]
//
// 创建一条记录
//
// 成功返回值
//	{
//		"code": "OK",
//		"id": newID,
//		"key": "xxxx"
//	}
//
// 失败返回值
//		code: xxx
//
func (h *handler) newRecode(c *gin.Context) {

	recode := createRecodeFromForm(c)
	if len(recode.RecordName) == 0 {
		res := map[string]interface{}{}
		res["code"] = CodeInvalidParam
		res["msg"] = "recode name can't be empty"
		c.JSON(http.StatusOK, res)
		return
	}

	_, err := h.ws.db.FindByName(recode.RecordName)
	if err == nil {
		res := map[string]interface{}{}
		res["code"] = CodeRecodeExist
		res["msg"] = "recode [" + recode.RecordName + "] exist"
		c.JSON(http.StatusOK, res)
		return
	}

	if recode.Dynamic {
		recode.UpdateKey.Valid = true
		recode.UpdateKey.String = uuid.New().String()
	}

	_, err = h.ws.db.CreateRecode(recode)
	if err != nil {
		res := map[string]interface{}{}
		res["code"] = CodeUnknowError
		res["msg"] = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res := map[string]interface{}{}
	res["code"] = CodeOK
	res["name"] = recode.RecordName
	if recode.UpdateKey.Valid {
		res["key"] = recode.UpdateKey.String
	}
	res["id"] = recode.ID

	c.JSON(http.StatusOK, res)
}

func (h *handler) getRecode(c *gin.Context) {
	c.Redirect(http.StatusFound, "/html/index.html")
}

// updateRecode -> [PATCH] :/recode/:id
//
// Ret Code:[200]
//
// 更新一条记录
//
// 成功返回值
// 	{
//		"result": "ok"
//	}
//
// 失败返回值
//	{
//		"result": "xxx"
//	}
//
func (h *handler) updateRecode(c *gin.Context) {
	r, err := h.getRecodeFromParam(c)
	if err != nil {
		return
	}

	r.RecordName = c.DefaultPostForm("name", r.RecordName)
	def := "false"
	if r.Dynamic {
		def = "true"
	}
	def = c.DefaultPostForm("dynamic", def)
	r.Dynamic = (def == "true")

	if !r.Dynamic {
		r.RecodeValue = c.DefaultPostForm("value", r.RecodeValue)
	}

	r.TTL, _ = strconv.Atoi(c.DefaultPostForm("ttl", fmt.Sprintf("%v", r.TTL)))
	if r.TTL == 0 {
		r.TTL = 600
	}

	err = h.ws.db.UpdateRecode(r)
	if err != nil {
		h.rspError(c, err)
	} else {
		h.rspOk(c)
	}
}

// deleteRecode -> [DELETE] :/recode/:id
//
// Ret Code:[200]
//
// 删除一条记录
//
// 成功返回值
// 	{
//		"result": "ok"
//	}
//
// 失败返回值
//	{
//		"result": "xxx"
//	}
//
func (h *handler) deleteRecode(c *gin.Context) {
	recode, err := h.getRecodeFromParam(c)
	if err != nil {
		return
	}

	err = h.ws.db.DeleteRecode(recode.ID)
	if err != nil {
		h.rspError(c, err)
	} else {
		h.rspOk(c)
	}
}
