package web

import (
	"net/http"
	"strings"

	"strconv"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/data/model"
	"github.com/inimei/ddns/errs"
)

const (
	pRoot = "/"

	pRecodes = "/recodes"
	pRecode  = "/recode/:id"
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
	CodeOK           = "ok"
	CodeDBError      = "DBError"
	CodeUnknowError  = "UnknowError"
	CodeInvalidParam = "InvalidParam"
	CodeRecodeExist  = "RecodeExist"
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
	r.Dynamic = true
	r.TTL, _ = strconv.Atoi(c.DefaultPostForm("ttl", "600"))
	if r.TTL == 0 {
		r.TTL = 600
	}

	return r
}

type handler struct {
	ws *WebServer
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
	return map[string]interface{}{}
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

// newRecode -> [POST] :/recodes
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
		c.HTML(http.StatusOK, "app_layout.html", parameters)
		return
	}

	l := c.DefaultQuery("limit", "10")
	o := c.DefaultQuery("offset", "0")

	limit, _ := strconv.ParseInt(l, 10, 64)
	offset, _ := strconv.ParseInt(o, 10, 64)

	data, err := h.ws.db.ReadData(int(offset), int(limit))
	if err != nil {
		log.Error(err.Error())
		h.rspError(c, err)
		return
	}

	dastamap := []map[string]interface{}{}
	for _, d := range data {
		dataJosn := map[string]interface{}{}
		dataJosn["id"] = d.ID
		dataJosn["name"] = d.RecordName
		dataJosn["dynamic"] = d.Dynamic
		dataJosn["ttl"] = d.TTL
		dataJosn["value"] = d.RecodeValue
		if d.UpdateKey.Valid {
			dataJosn["key"] = d.UpdateKey.String
		}
		dastamap = append(dastamap, dataJosn)
	}

	res := map[string]interface{}{}
	res["code"] = CodeOK
	res["recodes"] = dastamap

	c.JSON(http.StatusOK, res)
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
