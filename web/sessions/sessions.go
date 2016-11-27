package sessions

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("i don't know"))

const (
	CookieName = "ddns_sid"
)

//TODO: impl

func Login(w http.ResponseWriter, r *http.Request, name string) error {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := store.Get(r, CookieName)

	// Set some session values.
	session.Values["name"] = name
	session.Values[42] = 43
	// Save it before we write to the response/return from the handler.
	session.Save(r, w)

	return nil
}

func IsLogined(r *http.Request) bool {
	session, _ := store.Get(r, CookieName)
	if session.IsNew {
		return false
	}
	return true
}
