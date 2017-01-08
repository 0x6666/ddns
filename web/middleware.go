package web

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/web/sessions"
	"github.com/inimei/ddns/web/signature"
)

const (
	MwUserid    = "UserID"
	MwSecretKey = "secretKey"
)

// SignMiddleware 检查API签名
// 如果成功，会向context中写入secretkey字段
func (h *handler) SignMiddleware(c *gin.Context) {
	auth := c.Request.Header.Get("Authorization")
	if auth == "" {
		log.Debug("request has no authorization header")
		h.rspErrorCode(c, CodeNoAuthorization, "request has no authorization header")
		c.Abort()
		return
	}

	parts := strings.Split(auth, ":")
	if len(parts) != 3 {
		log.Debug("authorization header format error:" + auth)
		h.rspErrorCode(c, CodeAuthorizationError, "authorization header format error")
		c.Abort()
		return
	}

	log.Debug("request authorization:" + auth)

	err := signature.VerifySignature(parts[0], parts[1], parts[2], c.Request, c.Writer)
	if err != nil {
		log.Debug("verifySignature failed, %s", err)
		h.rspErrorCode(c, CodeVerifySignature, "verify signature")
		c.Abort()
		return
	}

	secretKey, err := signature.GetSecretKey(parts[1])
	if err != nil {
		log.Debug("get secretKey error failed: %v", err)
		h.rspErrorCode(c, CodeGetSecretKeyError, "get secretKey error failed: "+err.Error())
		c.Abort()
		return
	}

	c.Set(MwSecretKey, secretKey)
	c.Next()
}

func (h *handler) CookieAuthMiddleware(c *gin.Context) {

	userid, err := sessions.GetUserID(c.Request)
	if err != nil {
		if strings.ToLower(c.Request.Header.Get("DDNS-View")) == "true" {
			c.JSON(http.StatusUnauthorized, "")
			c.Abort()
			return
		}

		refer := c.Request.Referer()
		if len(refer) == 0 {
			refer = c.Request.RequestURI
		}
		c.Redirect(http.StatusTemporaryRedirect, "/login?to="+url.QueryEscape(refer))
		c.Abort()
		return
	}

	c.Set(MwUserid, userid)

	c.Next()
}
