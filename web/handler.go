package web

import (
	"net/http"
	"strings"

	"strconv"

	"github.com/gin-gonic/gin"
	"wpsgit.kingsoft.net/wpsep/util/log/log"
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
	CodeOK          = "OK"
	CodeDBError     = "DBError"
	CodeUnknowError = "UnknowError"
)

func requestType(r *http.Request) string {
	accept := r.Header["Accept"]
	if len(accept) == 0 {
		return MIMEHTML
	}

	accept = strings.Split(accept[0], ",")

	return accept[0]
}

type handler struct {
	ws *WebServer
}

func (h *handler) getTemplateParameter(c *gin.Context) map[string]interface{} {
	return map[string]interface{}{}
}

func (h *handler) root(c *gin.Context) {
	//c.Redirect(http.StatusFound, "/html/index.html")
	c.Redirect(http.StatusFound, pRecodes)
}

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

	res := map[string]interface{}{}
	res["result"] = CodeOK
	res["data"] = data

	c.JSON(http.StatusOK, res)
}

func (h *handler) newRecode(c *gin.Context) {
	c.Redirect(http.StatusFound, "/html/index.html")
}

func (h *handler) getRecode(c *gin.Context) {
	c.Redirect(http.StatusFound, "/html/index.html")
}

func (h *handler) rspError(c *gin.Context, err error) {
	c.JSON(http.StatusOK, gin.H{
		"result": CodeUnknowError,
		"msg":    err.Error(),
	})
}
