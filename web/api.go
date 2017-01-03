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
//		"domains": [
//			{
//				"id" : xxx,
//				"domain": "xxxx",
//				"recodes": [
//					{
//						"id" : xxx,
//						"value": "xxxx",
//						"type":	type,
//						"name":	"name"
//						"ttl": xxx,
//						"dynamic": true,
//						"key": "xxxx"
//					}
//				]
//			}
//		]
//	}
// 失败返回值
//		code: xxx
//
func (h *handler) apiGetRecodes(c *gin.Context) {

	l := c.DefaultQuery("limit", "10")
	o := c.DefaultQuery("offset", "0")

	limit, _ := strconv.ParseInt(l, 10, 64)
	offset, _ := strconv.ParseInt(o, 10, 64)

	domains, err := h.ws.db.GetAllDomains(int(offset), int(limit))
	if err != nil {
		log.Error(err.Error())
		h.rspError(c, err)
		return
	}

	domainsJson := []map[string]interface{}{}
	for _, d := range domains {
		dJson := map[string]interface{}{}
		dJson["id"] = d.ID
		dJson["domain"] = d.DomainName

		recodes, err := h.ws.db.GetRecodes(d.ID, 0, -1)
		if err != nil {
			log.Error(err.Error())
		} else {
			recodesJson := []map[string]interface{}{}
			for _, r := range recodes {
				rJson := map[string]interface{}{}
				rJson["id"] = r.ID
				rJson["value"] = r.RecodeValue
				rJson["type"] = r.RecordType
				rJson["name"] = r.RecordHost
				rJson["ttl"] = r.TTL
				rJson["dynamic"] = r.Dynamic
				rJson["key"] = r.UpdateKey

				recodesJson = append(recodesJson, rJson)
			}
			dJson["recodes"] = recodesJson
		}

		domainsJson = append(domainsJson, dJson)
	}

	res := map[string]interface{}{}
	res["code"] = CodeOK
	res["domains"] = domainsJson

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

	err = h.ws.db.UpdateRecode(recode.ID, recode)
	if err != nil {
		log.Error(err.Error())
		h.rspError(c, err)
		return
	}

	h.rspOk(c)
}
