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
	"github.com/inimei/ddns/web/sessions"
)

const (
	pRoot = "/"

	pRecodes = "/domain/:did/recodes"
	pRecode  = "/recode/:rid"
	pLogin   = "/login"
	pLogout  = "/logout"
	pAbout   = "/about"

	pDomains = "/domains"
	pDomain  = "/domain/:did"

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
	r.RecordHost = c.PostForm("host")
	r.RecodeValue = c.PostForm("value")
	t, _ := strconv.Atoi(c.DefaultPostForm("type", "1"))
	r.RecordType = model.RecodeType(t)
	if len(r.RecodeValue) == 0 {
		r.Dynamic = true
	}

	ttl, _ := strconv.Atoi(c.DefaultPostForm("ttl", "600"))

	r.TTL = uint32(ttl)
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

func (h *handler) root(c *gin.Context) {
	//c.Redirect(http.StatusFound, "/html/index.html")
	c.Redirect(http.StatusFound, pDomains)
}

// getLogin -> [GET] :/login
//
// Ret Code:[200]
//
func (h *handler) getLogin(c *gin.Context) {
	if sessions.IsLogined(c.Request) {
		to := c.Query("to")
		if len(to) == 0 {
			to = pRoot
		}
		c.Redirect(http.StatusPermanentRedirect, to)
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

	parameters["BreadcrumbSecs"] = SectionItems{
		&SectionItem{"About", "/about"},
	}
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

	sessions.Login(c.Writer, c.Request, model.DefUserID)

	h.rspOk(c)
}

// logout -> [POST] :/logout
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
func (h *handler) logout(c *gin.Context) {
	b := sessions.IsLogined(c.Request)
	if !b {
		c.JSON(http.StatusOK, JsonMap{"code": CodeInvalidSession})
		return
	}

	err := sessions.Logout(c.Writer, c.Request)
	if err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusOK, JsonMap{"code": CodeUnknowError})
		return
	}
	c.JSON(http.StatusOK, JsonMap{"code": CodeOK})
}

// getDomains -> [GET] :/domains?r=true&offset=10&limit=0
//
// Ret Code:[200]
//
// [request=json] 获取指定区间的记录数
// [request=html] 获取域名列表页面
//
// 成功返回值
//	{
//		"code": "OK"
//		"domains" : [
//			"id": id,
//			"domain"; "domain"
//		]
//	}
//
// 失败返回值
//		code: xxx
//
func (h *handler) getDomains(c *gin.Context) {
	if t := requestType(c.Request); t != MIMEJSON {
		parameters := h.getTemplateParameter(c)
		parameters["BreadcrumbSecs"] = SectionItems{
			&SectionItem{"Domain", "/domains"},
		}

		parameters["View"] = "domains_view"
		h.HTML(c, http.StatusOK, parameters)
		return
	}

	/*
		l := c.DefaultQuery("limit", "10")
		o := c.DefaultQuery("offset", "0")

		limit, _ := strconv.ParseInt(l, 10, 64)
		offset, _ := strconv.ParseInt(o, 10, 64)
	*/

	userid, _ := sessions.GetUserID(c.Request)
	ds, err := h.ws.db.GetDomains(userid)
	if err != nil {
		log.Error(err.Error())
		h.rspError(c, err)
		return
	}

	domainsJson := []JsonMap{}
	for _, d := range ds {
		domainsJson = append(domainsJson, JsonMap{
			"id":     d.ID,
			"domain": d.DomainName,
		})
	}

	res := map[string]interface{}{}
	res["code"] = CodeOK
	res["domains"] = domainsJson

	c.JSON(http.StatusOK, res)
}

// newDomain -> [POST] :/domains
//
// Ret Code:[200]
//
// 创建一个新的域名
//
// 成功返回值
//	{
//		"code": "OK",
//		"id": newID,
//	}
//
// 失败返回值
//		code: xxx
//
func (h *handler) newDomain(c *gin.Context) {

	d := h.createDomainFromForm(c)
	if len(d.DomainName) == 0 {
		c.JSON(http.StatusOK, JsonMap{
			"code": CodeInvalidParam,
			"msg":  "domain name can't be empty",
		})
		return
	}

	_, err := h.ws.db.FindDomainByName(d.DomainName)
	if err == nil {
		c.JSON(http.StatusOK, JsonMap{
			"code": CodeInvalidParam,
			"msg":  "domain name [" + d.DomainName + "] already exist",
		})
		return
	}

	userid, _ := sessions.GetUserID(c.Request)
	id, err := h.ws.db.NewDomain(userid, d)
	if err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusOK, JsonMap{
			"code": CodeDBError,
			"msg":  "create domain [" + d.DomainName + "] failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, JsonMap{
		"code": CodeOK,
		"id":   id,
	})
}

func (h *handler) getDomain(c *gin.Context) {

}

// updateDomain -> [PATCH] :/domain/:did
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
func (h *handler) updateDomain(c *gin.Context) {
	d, err := h.getDomainFromParam(c)
	if err != nil {
		return
	}

	newD := h.createDomainFromForm(c)
	if len(newD.DomainName) == 0 {
		c.JSON(http.StatusOK, JsonMap{
			"code": CodeInvalidParam,
			"msg":  "domain name can't be empty",
		})
		return
	}

	if newD.DomainName == d.DomainName {
		c.JSON(http.StatusOK, JsonMap{"code": CodeOK})
		return
	}

	_, err = h.ws.db.FindDomainByName(newD.DomainName)
	if err == nil {
		c.JSON(http.StatusOK, JsonMap{
			"code": CodeInvalidParam,
			"msg":  "domain name [" + newD.DomainName + "] already exist",
		})
		return
	}

	err = h.ws.db.UpdateDomain(d.ID, newD.DomainName)
	if err != nil {
		h.rspError(c, err)
	} else {
		h.rspOk(c)
	}
}

// deleteDomain -> [DELETE] :/domain/:did
//
// Ret Code:[200]
//
// 删除一个domain
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
func (h *handler) deleteDomain(c *gin.Context) {
	d, err := h.getDomainFromParam(c)
	if err != nil {
		return
	}

	err = h.ws.db.DeleteDomain(d.ID)
	if err != nil {
		h.rspError(c, err)
	} else {
		h.rspOk(c)
	}
}

// getRecodes -> [GET] :/domain/:did/recodes
//
// Ret Code:[200]
//
// [request=json] 获取指定区间的记录数
// [request=html] 获取记录列表页面
//
// 成功返回值
//	{
//		"code": "OK",
//		"domainID": did,
//		"domainName": "name",
//		"recodes": [
//			{
//				"id" : xxx,
//				"host": "xxxx",
//				"type":	type
//				"value": "xxxx"
//				"ttl": xxx,
//				"dynamic": true,
//				"key": "xxxx",
//			}
//		]
//	}
//
// 失败返回值
//		code: xxx
//
func (h *handler) getRecodes(c *gin.Context) {

	d, _ := h.getDomainFromParam(c)
	if t := requestType(c.Request); t != MIMEJSON {
		parameters := h.getTemplateParameter(c)
		parameters["BreadcrumbSecs"] = SectionItems{
			&SectionItem{"Domain", "/domains"},
			&SectionItem{"Recode", fmt.Sprintf("/domain/%v/recodes", d.ID)},
		}
		parameters["View"] = "recode_list"
		parameters["Did"] = d.ID
		h.HTML(c, http.StatusOK, parameters)
		return
	}

	l := c.DefaultQuery("limit", "10")
	o := c.DefaultQuery("offset", "0")

	limit, _ := strconv.ParseInt(l, 10, 64)
	offset, _ := strconv.ParseInt(o, 10, 64)

	rs, err := h.ws.db.GetRecodes(d.ID, int(offset), int(limit))
	if err != nil {
		c.JSON(http.StatusOK, JsonMap{
			"code": CodeDBError,
			"msg":  "get recodes failed: " + err.Error(),
		})
		return
	}

	res := JsonMap{}
	res["domainID"] = d.ID
	res["domainName"] = d.DomainName

	recodesJson := []map[string]interface{}{}
	for _, r := range rs {
		rJson := map[string]interface{}{}
		rJson["id"] = r.ID
		rJson["host"] = r.RecordHost
		rJson["type"] = r.RecordType
		rJson["value"] = r.RecodeValue
		rJson["ttl"] = r.TTL
		rJson["dynamic"] = r.Dynamic
		if r.UpdateKey.Valid {
			rJson["key"] = r.UpdateKey.String
		}

		recodesJson = append(recodesJson, rJson)
	}
	res["recodes"] = recodesJson
	res["code"] = CodeOK

	c.JSON(http.StatusOK, res)
}

// newRecode -> [POST] :/domain/:did/recodes
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

	d, err := h.getDomainFromParam(c)
	if err != nil {
		return
	}

	recode := createRecodeFromForm(c)
	if len(recode.RecordHost) == 0 {
		c.JSON(http.StatusOK, JsonMap{
			"code": CodeInvalidParam,
			"msg":  "recode name can't be empty",
		})
		return
	}

	_, err = h.ws.db.FindByName(d.ID, recode.RecordHost)
	if err == nil {
		c.JSON(http.StatusOK, JsonMap{
			"code": CodeRecodeExist,
			"msg":  "recode [" + recode.RecordHost + "] exist",
		})
		return
	}

	if recode.Dynamic {
		recode.UpdateKey.Valid = true
		recode.UpdateKey.String = uuid.New().String()
	}

	_, err = h.ws.db.NewRecode(d.ID, recode)
	if err != nil {
		c.JSON(http.StatusOK, JsonMap{
			"code": CodeUnknowError,
			"msg":  err.Error()})
		return
	}

	res := JsonMap{
		"code": CodeOK,
		"name": recode.RecordHost,
		"id":   recode.ID,
	}
	if recode.UpdateKey.Valid {
		res["key"] = recode.UpdateKey.String
	}

	c.JSON(http.StatusOK, res)
}

func (h *handler) getRecode(c *gin.Context) {
	c.Redirect(http.StatusFound, "/html/index.html")
}

// updateRecode -> [PATCH] :/recode/:rid
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

	r.RecordHost = c.DefaultPostForm("name", r.RecordHost)
	def := "false"
	if r.Dynamic {
		def = "true"
	}
	def = c.DefaultPostForm("dynamic", def)
	r.Dynamic = (def == "true")

	if !r.Dynamic {
		r.RecodeValue = c.DefaultPostForm("value", r.RecodeValue)
	}

	ttl, _ := strconv.Atoi(c.DefaultPostForm("ttl", fmt.Sprintf("%v", r.TTL)))

	r.TTL = uint32(ttl)
	if r.TTL == 0 {
		r.TTL = 600
	}

	t, _ := strconv.Atoi(c.DefaultPostForm("type", fmt.Sprintf("%v", model.CNAME)))
	r.RecordType = model.RecodeType(t)

	err = h.ws.db.UpdateRecode(r.ID, r)
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
