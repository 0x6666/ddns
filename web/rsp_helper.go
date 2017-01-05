package web

import (
	"fmt"
	"net/http"
	"strings"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/data/model"
	"github.com/inimei/ddns/errs"
)

type JsonMap map[string]interface{}

type SectionItem struct {
	Name string
	Href string
}

type SectionItems []*SectionItem

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

func (h *handler) createDomainFromForm(c *gin.Context) *model.Domain {

	d := &model.Domain{}
	d.DomainName = strings.ToLower(c.PostForm("domain"))

	return d
}
