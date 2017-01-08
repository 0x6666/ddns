package sessions

import (
	"net/http"

	"github.com/boj/redistore"
	"github.com/gorilla/sessions"
	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/config"
	"github.com/inimei/ddns/errs"
	"github.com/inimei/ddns/web/sessions/memstore"
)

var store sessions.Store

const (
	secretKey  = "DDNS-secret-key"
	CookieName = "ddns_sid"
)

func init() {
	o := &sessions.Options{
		Path:     "/",
		Domain:   "",
		MaxAge:   config.Data.Session.Maxage,
		Secure:   false,
		HttpOnly: false,
	}

	switch config.Data.Session.Backend {
	case "memery":
		s := memstore.NewMemStore([]byte(secretKey))
		s.Options = o
		store = s
	case "redis":
		s, err := redistore.NewRediStoreWithDB(
			10,
			"tcp",
			config.Data.Redis.Host,
			config.Data.Redis.Passwd,
			"0",
			[]byte(secretKey))

		if err != nil {
			panic(err)
		}

		s.Options = o
		store = s
	default:
		panic("invalid session backend: " + config.Data.Session.Backend)
	}
}

// Login ...
func Login(w http.ResponseWriter, r *http.Request, userid int64) error {

	session, _ := store.Get(r, CookieName)

	session.Values["userid"] = userid

	return session.Save(r, w)
}

// IsLogined ..
func IsLogined(r *http.Request) bool {
	session, _ := store.Get(r, CookieName)
	if session.IsNew {
		return false
	}
	return true
}

func Logout(w http.ResponseWriter, r *http.Request) error {
	session, _ := store.Get(r, CookieName)
	if session.IsNew {
		return nil
	}

	session.Options.MaxAge = -1
	err := session.Save(r, w)
	if err != nil {
		log.Error(err.Error())
	}
	return err
}

func GetUserID(r *http.Request) (int64, error) {
	session, _ := store.Get(r, CookieName)
	if session.IsNew {
		return 0, errs.ErrInvalidSession
	}

	userid, b := session.Values["userid"]
	if !b {
		return 0, errs.ErrInvalidSession
	}

	id, b := userid.(int64)
	if !b {
		return 0, errs.ErrInvalidSession
	}

	return id, nil
}
