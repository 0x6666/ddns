package controllers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/yangsongfwd/backup/log"
	"github.com/yangsongfwd/ddns/app/sessions"
	"github.com/yangsongfwd/ddns/app/signature"
)

const (
	MwUserid    = "UserID"
	MwSecretKey = "secretKey"
)

// SignMiddleware 检查API签名
// 如果成功，会向context中写入secretkey字段
func (c ContrllerBase) SignMiddleware() {
	auth := c.C.Request.Header.Get("Authorization")
	if auth == "" {
		log.Debug("request has no authorization header")
		c.rspErrorCode(CodeNoAuthorization, "request has no authorization header")
		c.Abort()
		return
	}

	parts := strings.Split(auth, ":")
	if len(parts) != 3 {
		log.Debug("authorization header format error:" + auth)
		c.rspErrorCode(CodeAuthorizationError, "authorization header format error")
		c.Abort()
		return
	}

	log.Debug("request authorization:" + auth)

	err := signature.VerifySignature(parts[0], parts[1], parts[2], c.C.Request, c.C.Writer)
	if err != nil {
		log.Debug("verifySignature failed, %s", err)
		c.rspErrorCode(CodeVerifySignature, "verify signature")
		c.Abort()
		return
	}

	secretKey, err := signature.GetSecretKey(parts[1])
	if err != nil {
		log.Debug("get secretKey error failed: %v", err)
		c.rspErrorCode(CodeGetSecretKeyError, "get secretKey error failed: "+err.Error())
		c.Abort()
		return
	}

	c.C.Set(MwSecretKey, secretKey)
	//_c.Next()
}

func (h ContrllerBase) CookieAuthMiddleware() {
	userid, err := sessions.GetUserID(h.C.Request)
	if err != nil {
		if strings.ToLower(h.C.Request.Header.Get("DDNS-View")) == "true" {
			h.C.JSON(http.StatusUnauthorized, "")
			h.Abort()
			return
		}

		refer := h.C.Request.Referer()
		if len(refer) == 0 {
			refer = h.C.Request.RequestURI
		}
		h.C.Redirect(http.StatusTemporaryRedirect, "/login?to="+url.QueryEscape(refer))
		h.Abort()
		return
	}

	h.C.Set(MwUserid, userid)
}
