package web

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/inimei/backup/log"
)

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
