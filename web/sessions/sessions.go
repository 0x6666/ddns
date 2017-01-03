package sessions

import (
	"net/http"

	"github.com/inimei/ddns/errs"
	"github.com/inimei/ddns/web/sessions/memstore"
)

var store = memstore.NewMemStore([]byte("i don't know"))

const (
	CookieName = "ddns_sid"
)

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

func GetUserID(r *http.Request) (int64, error) {
	session, _ := store.Get(r, CookieName)
	if session.IsNew {
		return 0, errs.ErrInvalidSession
	}

	userid, b := session.Values["userid"]
	if b {
		return 0, errs.ErrInvalidSession
	}

	id, b := userid.(int64)
	if b {
		return 0, errs.ErrInvalidSession
	}

	return id, nil
}
