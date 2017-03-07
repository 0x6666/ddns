package web

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"strconv"

	"net"

	"github.com/gin-gonic/gin"
	"github.com/yangsongfwd/backup/log"
	"github.com/yangsongfwd/ddns/data/model"
	"github.com/yangsongfwd/ddns/errs"
)

type JsonMap map[string]interface{}

type SectionItem struct {
	Name string
	Href string
}

type SectionItems []*SectionItem

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

	val, exist := c.Get(MwUserid)
	if exist {
		//TODO: impl user
		parameters["CurUser"] = JsonMap{"UserID": val, "UserName": "YangSong"}
	}

	return parameters
}

func (h *handler) HTML(c *gin.Context, code int, params map[string]interface{}) {
	c.HTML(code, fmt.Sprintf("%v", params["Layout"]), params)
}

func (h *handler) getDomainFromParam(c *gin.Context) (*model.Domain, error) {
	did, err := strconv.ParseInt(c.Param("did"), 10, 64)
	if err != nil {
		err = fmt.Errorf("get domian id from param failed: %v", err.Error())
		log.Error(err.Error())
		h.rspErrorCode(c, CodeInvalidParam, err.Error())
		return nil, errs.ErrInvalidParam
	}

	domain, err := h.ws.db.GetDomain(did)
	if err != nil {
		err = fmt.Errorf("get domain [%v] failed: %v", did, err.Error())
		log.Error(err.Error())
		h.rspErrorCode(c, CodeDBError, err.Error())
		return nil, errs.ErrInvalidParam
	}

	return domain, nil
}

func (h *handler) getRecodeFromParam(c *gin.Context) (*model.Recode, error) {
	rid, err := strconv.ParseInt(c.Param("rid"), 10, 0)
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
	return r, nil
}

func (h *handler) createDomainFromForm(c *gin.Context) (*model.Domain, error) {

	domain := strings.ToLower(c.PostForm("domain"))
	if !validateDomainName(domain) {
		err := fmt.Errorf("invalid domain [%v]", domain)
		log.Error(err.Error())
		return nil, err
	}

	d := &model.Domain{}
	d.DomainName = domain
	d.Synced = false

	return d, nil
}

func (h *handler) createRecodeFromForm(c *gin.Context, d *model.Domain) (*model.Recode, error) {

	r := &model.Recode{}
	r.RecordHost = c.PostForm("host")
	if r.RecordHost != "@" && !validateDomainName(r.RecordHost+"."+d.DomainName) {
		err := fmt.Errorf("invalid host name [%v]", r.RecordHost)
		log.Error(err.Error())
		return nil, err
	}

	strT := c.PostForm("type")
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

	r.RecodeValue = c.PostForm("value")
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

	ttl, _ := strconv.Atoi(c.DefaultPostForm("ttl", "600"))

	r.TTL = uint32(ttl)
	if r.TTL == 0 {
		r.TTL = 600
	}

	return r, nil
}
