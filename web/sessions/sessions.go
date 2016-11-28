package sessions

import (
	"net/http"

	"github.com/inimei/ddns/web/sessions/memstore"
)

var store = memstore.NewMemStore([]byte("i don't know"))

const (
	CookieName = "ddns_sid"
)

// Login ...
func Login(w http.ResponseWriter, r *http.Request, name string) error {

	session, _ := store.Get(r, CookieName)

	session.Values["name"] = name

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
