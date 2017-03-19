package controllers

import (
	"net"
	"net/http"
	"strconv"

	"github.com/yangsongfwd/backup/log"
	"github.com/yangsongfwd/ddns/app/model"
)

type ApiCtrl struct {
	ContrllerBase
}

// ApiGetDataSchemaVersion -> [POST] :/api/dataversion
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
func (a ApiCtrl) ApiGetDataSchemaVersion() {

	res := JsonMap{}
	res["code"] = CodeOK

	v := model.Version{}
	v.SchemaVersion = model.CurrentVersion
	v.DataVersion = a.db().GetVersion()

	res["version"] = v
	a.JSON(http.StatusOK, res)
}

// ApiGetRecodes -> [POST] :/api/recodes
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
func (a ApiCtrl) ApiGetRecodes() {

	l := a.C.DefaultQuery("limit", "10")
	o := a.C.DefaultQuery("offset", "0")

	limit, _ := strconv.ParseInt(l, 10, 64)
	offset, _ := strconv.ParseInt(o, 10, 64)

	domains, err := a.db().GetAllDomains(int(offset), int(limit))
	if err != nil {
		log.Error(err.Error())
		a.rspError(err)
		return
	}

	domainsJson := []JsonMap{}
	for _, d := range domains {
		dJson := JsonMap{}
		dJson["id"] = d.ID
		dJson["domain"] = d.DomainName
		dJson["userid"] = d.UserID

		recodes, err := a.db().GetRecodes(d.ID, 0, -1)
		if err != nil {
			log.Error(err.Error())
		} else {
			recodesJson := []JsonMap{}
			for _, r := range recodes {
				rJson := JsonMap{}
				rJson["id"] = r.ID
				rJson["value"] = r.RecodeValue
				rJson["type"] = r.RecordType
				rJson["name"] = r.RecordHost
				rJson["ttl"] = r.TTL
				rJson["dynamic"] = r.Dynamic
				if r.UpdateKey.Valid {
					rJson["key"] = r.UpdateKey.String
				}

				recodesJson = append(recodesJson, rJson)
			}
			dJson["recodes"] = recodesJson
		}

		domainsJson = append(domainsJson, dJson)
	}

	res := JsonMap{}
	res["code"] = CodeOK
	res["domains"] = domainsJson

	a.JSON(http.StatusOK, res)
}

// ApiUpdateRecode -> [POST] :/api/update
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
func (a ApiCtrl) ApiUpdateRecode() {

	key := a.C.PostForm("key")
	if len(key) == 0 {
		a.rspErrorCode(CodeKeyIsEmpty, "key is empty")
		return
	}

	ip := a.C.PostForm("ip")
	if len(ip) != 0 {
		if i := net.ParseIP(ip); i == nil {
			a.rspErrorCode(CodeInvalidIP, "invalid ip address")
			return
		}
	}

	recode, err := a.db().FindByKey(key)
	if err != nil {
		log.Error(err.Error())
		a.rspError(err)
		return
	}

	if recode.RecodeValue == ip {
		a.rspOk()
		return
	}

	recode.RecodeValue = ip

	err = a.db().UpdateRecode(recode.ID, recode)
	if err != nil {
		log.Error(err.Error())
		a.rspError(err)
		return
	}

	a.rspOk()
}
