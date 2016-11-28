package web

import (
	"net"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/data/model"
)

// apiGetDataSchemaVersion -> [POST] :/api/dataversion
//
// Ret Code:[200]
//
// 成功返回值
//	{
//		"code": "OK",
//		"version": "0.1"
//	}
//
// 失败返回值
//	code: xxx
//
func (h *handler) apiGetDataSchemaVersion(c *gin.Context) {

	res := map[string]interface{}{}
	res["code"] = CodeOK

	v := model.Version{}
	v.SchemaVersion = model.CurrentVersion
	v.DataVersion = h.ws.db.GetVersion()

	res["version"] = v

	c.JSON(http.StatusOK, res)
}

// apiGetRecodes -> [POST] :/api/recodes
//
// Ret Code:[200]
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
func (h *handler) apiGetRecodes(c *gin.Context) {

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
		dataJosn["type"] = d.RecordType
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

// apiUpdateRecode -> [POST] :/api/update
//
// Ret Code:[200]
//
// 更新一个记录
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
func (h *handler) apiUpdateRecode(c *gin.Context) {

	key := c.PostForm("key")
	if len(key) == 0 {
		h.rspErrorCode(c, CodeKeyIsEmpty, "key is empty")
		return
	}

	ip := c.PostForm("ip")
	if len(ip) != 0 {
		if i := net.ParseIP(ip); i == nil {
			h.rspErrorCode(c, CodeInvalidIP, "invalid ip address")
			return
		}
	}

	recode, err := h.ws.db.FindByKey(key)
	if err != nil {
		log.Error(err.Error())
		h.rspError(c, err)
		return
	}

	if recode.RecodeValue == ip {
		h.rspOk(c)
		return
	}

	recode.RecodeValue = ip
	err = h.ws.db.UpdateRecode(recode)
	if err != nil {
		log.Error(err.Error())
		h.rspError(c, err)
		return
	}

	h.rspOk(c)
}
