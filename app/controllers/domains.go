package controllers

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/0x6666/backup/log"
	"github.com/0x6666/ddns/app/model"
	"github.com/0x6666/ddns/app/sessions"
	"github.com/0x6666/ddns/errs"
	"github.com/google/uuid"
)

type DomainsCtrl struct {
	ContrllerBase
}

func validateDomainName(domain string) bool {
	//todo
	return true

	/*	log.Info(domain)

		RegExp := regexp.MustCompile(`^(([a-zA-Z]{1})|([a-zA-Z]{1}[a-zA-Z]{1})|([a-zA-Z]{1}[0-9]{1})|([0-9]{1}[a-zA-Z]{1})|([a-zA-Z0-9][a-zA-Z0-9-_]{1,61}[a-zA-Z0-9]))\.([a-zA-Z]{2,6}|[a-zA-Z0-9-]{2,30}\.[a-zA-Z]{2,3})$`)
		res := RegExp.MatchString(domain)
		log.Info("%v", res)

		return RegExp.MatchString(domain)
	*/
}

func validIP4(ipAddress string) bool {
	ipAddress = strings.Trim(ipAddress, " ")

	re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	if re.MatchString(ipAddress) {
		return true
	}
	return false
}

func (d DomainsCtrl) getDomainFromParam() (*model.Domain, error) {
	did, err := strconv.ParseInt(d.C.Param("did"), 10, 64)
	if err != nil {
		err = fmt.Errorf("get domian id from param failed: %v", err.Error())
		log.Error(err.Error())
		d.rspErrorCode(CodeInvalidParam, err.Error())
		return nil, errs.ErrInvalidParam
	}

	domain, err := d.db().GetDomain(did)
	if err != nil {
		err = fmt.Errorf("get domain [%v] failed: %v", did, err.Error())
		log.Error(err.Error())
		d.rspErrorCode(CodeDBError, err.Error())
		return nil, errs.ErrInvalidParam
	}

	return domain, nil
}

func (d DomainsCtrl) getRecodeFromParam() (*model.Recode, error) {
	rid, err := strconv.ParseInt(d.C.Param("rid"), 10, 0)
	if err != nil {
		err = fmt.Errorf("parse recode id failed: %v", err.Error())
		log.Error(err.Error())
		d.rspErrorCode(CodeInvalidParam, err.Error())
		return nil, errs.ErrInvalidParam
	}

	r, err := d.db().GetRecode(rid)
	if err != nil {
		err = fmt.Errorf("get recode [%v] failed: %v", rid, err.Error())
		log.Error(err.Error())
		d.rspErrorCode(CodeDBError, err.Error())
		return nil, errs.ErrInvalidParam
	}
	return r, nil
}

func (d DomainsCtrl) createDomainFromForm() (*model.Domain, error) {

	domain := strings.ToLower(d.C.PostForm("domain"))
	if !validateDomainName(domain) {
		err := fmt.Errorf("invalid domain [%v]", domain)
		log.Error(err.Error())
		return nil, err
	}

	_d := &model.Domain{}
	_d.DomainName = domain
	_d.Synced = false

	return _d, nil
}

func (d DomainsCtrl) createRecodeFromForm(_d *model.Domain) (*model.Recode, error) {

	r := &model.Recode{}
	r.RecordHost = d.C.PostForm("host")
	if r.RecordHost != "@" && !validateDomainName(r.RecordHost+"."+_d.DomainName) {
		err := fmt.Errorf("invalid host name [%v]", r.RecordHost)
		log.Error(err.Error())
		return nil, err
	}

	strT := d.C.PostForm("type")
	if len(strT) == 0 {
		err := errors.New("recode type is empty")
		log.Error(err.Error())
		return nil, err
	}

	t, _ := strconv.Atoi(strT)
	if t <= int(model.NONE_) || t >= int(model.LAST_) {
		err := fmt.Errorf("invalid recode type [%v]", strT)
		log.Error(err.Error())
		return nil, err
	}
	r.RecordType = model.RecodeType(t)

	r.RecodeValue = d.C.PostForm("value")
	if len(r.RecodeValue) == 0 {
		r.Dynamic = true
	} else {
		switch r.RecordType {
		case model.A:
			if !validIP4(r.RecodeValue) {
				err := fmt.Errorf("invalid A recode [%v]", r.RecodeValue)
				log.Error(err.Error())
				return nil, err
			}
		case model.AAAA:
			ip := net.ParseIP(r.RecodeValue)
			if ip == nil || ip.To4() != nil {
				err := fmt.Errorf("invalid AAAA recode [%v]", r.RecodeValue)
				log.Error(err.Error())
				return nil, err
			}
		case model.CNAME:
			s := r.RecodeValue
			l := len(s)
			if l == 0 || s[l-1] != '.' || !validateDomainName(s[0:l-1]) {
				err := fmt.Errorf("invalid CNAME recode [%v]", r.RecodeValue)
				log.Error(err.Error())
				return nil, err
			}
		}
	}

	ttl, _ := strconv.Atoi(d.C.DefaultPostForm("ttl", "600"))

	r.TTL = uint32(ttl)
	if r.TTL == 0 {
		r.TTL = 600
	}

	return r, nil
}

// GetDomains -> [GET] :/domains?r=true&offset=10&limit=0
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
func (d DomainsCtrl) GetDomains() {
	if !d.IsJsonReqest() {
		parameters := d.getTemplateParameter()
		parameters["BreadcrumbSecs"] = SectionItems{
			&SectionItem{"Domain", "/domains"},
		}

		parameters["View"] = "domains_view"
		d.HTML(http.StatusOK, parameters)
		return
	}

	/*
		l := c.DefaultQuery("limit", "10")
		o := c.DefaultQuery("offset", "0")

		limit, _ := strconv.ParseInt(l, 10, 64)
		offset, _ := strconv.ParseInt(o, 10, 64)
	*/

	userid, _ := sessions.GetUserID(d.C.Request)
	ds, err := d.db().GetDomains(userid)
	if err != nil {
		log.Error(err.Error())
		d.rspError(err)
		return
	}

	domainsJSON := []JsonMap{}
	for _, d := range ds {
		domainsJSON = append(domainsJSON, JsonMap{
			"id":     d.ID,
			"domain": d.DomainName,
		})
	}

	res := JsonMap{}
	res["code"] = CodeOK
	res["domains"] = domainsJSON

	d.JSON(http.StatusOK, res)
}

// NewDomain -> [POST] :/domains
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
func (h DomainsCtrl) NewDomain() {

	d, err := h.createDomainFromForm()
	if err != nil {
		h.JSON(http.StatusOK, JsonMap{
			"code": CodeInvalidParam,
			"msg":  err.Error(),
		})
		return
	}

	_, err = h.db().FindDomainByName(d.DomainName)
	if err == nil {
		h.JSON(http.StatusOK, JsonMap{
			"code": CodeInvalidParam,
			"msg":  "domain name [" + d.DomainName + "] already exist",
		})
		return
	}

	userid, _ := sessions.GetUserID(h.C.Request)
	id, err := h.db().NewDomain(userid, d)
	if err != nil {
		log.Error(err.Error())
		h.JSON(http.StatusOK, JsonMap{
			"code": CodeDBError,
			"msg":  "create domain [" + d.DomainName + "] failed: " + err.Error(),
		})
		return
	}

	h.JSON(http.StatusOK, JsonMap{
		"code": CodeOK,
		"id":   id,
	})
}

// GetDomain ....
func (h DomainsCtrl) GetDomain() {

}

// UpdateDomain -> [PATCH] :/domain/:did
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
func (h DomainsCtrl) UpdateDomain() {
	d, err := h.getDomainFromParam()
	if err != nil {
		return
	}

	newD, err := h.createDomainFromForm()
	if err != nil {
		h.JSON(http.StatusOK, JsonMap{
			"code": CodeInvalidParam,
			"msg":  err.Error(),
		})
		return
	}

	if newD.DomainName == d.DomainName {
		h.JSON(http.StatusOK, JsonMap{"code": CodeOK})
		return
	}

	_, err = h.db().FindDomainByName(newD.DomainName)
	if err == nil {
		h.JSON(http.StatusOK, JsonMap{
			"code": CodeInvalidParam,
			"msg":  "domain name [" + newD.DomainName + "] already exist",
		})
		return
	}

	err = h.db().UpdateDomain(d.ID, newD.DomainName)
	if err != nil {
		h.rspError(err)
	} else {
		h.rspOk()
	}
}

// DeleteDomain -> [DELETE] :/domain/:did
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
func (h DomainsCtrl) DeleteDomain() {
	d, err := h.getDomainFromParam()
	if err != nil {
		return
	}

	err = h.db().DeleteDomain(d.ID)
	if err != nil {
		h.rspError(err)
	} else {
		h.rspOk()
	}
}

// GetRecodes -> [GET] :/domain/:did/recodes
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
func (h DomainsCtrl) GetRecodes() {

	d, _ := h.getDomainFromParam()
	if !h.IsJsonReqest() {
		parameters := h.getTemplateParameter()
		parameters["BreadcrumbSecs"] = SectionItems{
			&SectionItem{"Domain", "/domains"},
			&SectionItem{"Recode", fmt.Sprintf("/domain/%v/recodes", d.ID)},
		}
		parameters["View"] = "recode_list"
		parameters["Did"] = d.ID
		h.HTML(http.StatusOK, parameters)
		return
	}

	l := h.C.DefaultQuery("limit", "10")
	o := h.C.DefaultQuery("offset", "0")

	limit, _ := strconv.ParseInt(l, 10, 64)
	offset, _ := strconv.ParseInt(o, 10, 64)

	rs, err := h.db().GetRecodes(d.ID, int(offset), int(limit))
	if err != nil {
		h.JSON(http.StatusOK, JsonMap{
			"code": CodeDBError,
			"msg":  "get recodes failed: " + err.Error(),
		})
		return
	}

	res := JsonMap{}
	res["domainID"] = d.ID
	res["domainName"] = d.DomainName

	recodesJSON := []JsonMap{}
	for _, r := range rs {
		rJSON := JsonMap{}
		rJSON["id"] = r.ID
		rJSON["host"] = r.RecordHost
		rJSON["type"] = r.RecordType
		rJSON["value"] = r.RecodeValue
		rJSON["ttl"] = r.TTL
		rJSON["dynamic"] = r.Dynamic
		if r.UpdateKey.Valid {
			rJSON["key"] = r.UpdateKey.String
		}

		recodesJSON = append(recodesJSON, rJSON)
	}
	res["recodes"] = recodesJSON
	res["code"] = CodeOK

	h.JSON(http.StatusOK, res)
}

// NewRecode -> [POST] :/domain/:did/recodes
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
func (h DomainsCtrl) NewRecode() {

	d, err := h.getDomainFromParam()
	if err != nil {
		return
	}

	recode, err := h.createRecodeFromForm(d)
	if err != nil {
		h.JSON(http.StatusOK, JsonMap{
			"code": CodeInvalidParam,
			"msg":  err.Error(),
		})
		return
	}

	_, err = h.db().FindByName(d.ID, recode.RecordHost)
	if err == nil {
		h.JSON(http.StatusOK, JsonMap{
			"code": CodeRecodeExist,
			"msg":  "recode [" + recode.RecordHost + "] exist",
		})
		return
	}

	if recode.Dynamic {
		recode.UpdateKey.Valid = true
		recode.UpdateKey.String = uuid.New().String()
	}

	_, err = h.db().NewRecode(d.ID, recode)
	if err != nil {
		h.JSON(http.StatusOK, JsonMap{
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

	h.JSON(http.StatusOK, res)
}

//GetRecode ...
func (h DomainsCtrl) GetRecode() {
	h.C.Redirect(http.StatusFound, "/html/index.html")
}

// UpdateRecode -> [PATCH] :/recode/:rid
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
func (h DomainsCtrl) UpdateRecode() {
	r, err := h.getRecodeFromParam()
	if err != nil {
		return
	}

	d, err := h.db().GetDomain(r.DomainID)
	if err != nil {
		log.Error(err.Error())
		h.JSON(http.StatusOK, JsonMap{
			"code": CodeInvalidParam,
			"msg":  err.Error(),
		})
		return
	}

	newR, err := h.createRecodeFromForm(d)
	if err != nil {
		log.Error(err.Error())
		h.JSON(http.StatusOK, JsonMap{
			"code": CodeInvalidParam,
			"msg":  err.Error(),
		})
		return
	}
	newR.ID = r.ID
	newR.DomainID = r.DomainID

	err = h.db().UpdateRecode(newR.ID, newR)
	if err != nil {
		log.Error(err.Error())
		h.rspError(err)
	} else {
		h.rspOk()
	}
}

// DeleteRecode -> [DELETE] :/recode/:id
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
func (h DomainsCtrl) DeleteRecode() {
	recode, err := h.getRecodeFromParam()
	if err != nil {
		return
	}

	err = h.db().DeleteRecode(recode.ID)
	if err != nil {
		h.rspError(err)
	} else {
		h.rspOk()
	}
}
