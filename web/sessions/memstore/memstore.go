package memstore

import (
	"encoding/base32"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"sync"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/inimei/backup/log"
)

// NewMemStore returns a new MemStore.
func NewMemStore(keyPairs ...[]byte) *MemStore {

	return &MemStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		data: map[string]interface{}{},
	}
}

// MemStore stores sessions in memory
type MemStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options // default configuration

	data     map[string]interface{}
	dataMetx sync.RWMutex
}

// MaxLength restricts the maximum length of new sessions to l.
// If l is 0 there is no limit to the size of a session, use with caution.
// The default for a new MemcacheStore is 4096.
func (s *MemStore) MaxLength(l int) {
	for _, c := range s.Codecs {
		if codec, ok := c.(*securecookie.SecureCookie); ok {
			codec.MaxLength(l)
		}
	}
}

// Get returns a session for the given name after adding it to the registry.
//
// See CookieStore.Get().
func (s *MemStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
//
// See CookieStore.New().
func (s *MemStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true
	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			err = s.load(session)
			if err == nil {
				session.IsNew = false
			}
		}
	}
	return session, err
}

// Save adds a single session to the response.
func (s *MemStore) Save(r *http.Request, w http.ResponseWriter,
	session *sessions.Session) error {
	if session.ID == "" {
		session.ID = strings.TrimRight(
			base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32)), "=")
	}
	if err := s.save(session); err != nil {
		return err
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID,
		s.Codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

func (s *MemStore) save(session *sessions.Session) error {

	vals := make(map[string]interface{}, len(session.Values))
	for k, v := range session.Values {
		ks, ok := k.(string)
		if !ok {
			err := fmt.Errorf("Non-string key value, cannot jsonize: %v", k)
			log.Error(err.Error())
			return err
		}
		vals[ks] = v
	}

	s.dataMetx.Lock()
	defer s.dataMetx.Unlock()
	s.data[session.ID] = vals

	return nil
}

func (s *MemStore) load(session *sessions.Session) error {
	s.dataMetx.RLock()
	defer s.dataMetx.RUnlock()

	it, ok := s.data[session.ID]
	if !ok {
		return errors.New("invalid session id")
	}

	vals, ok := it.(map[string]interface{})
	if !ok {
		return errors.New("invalid session id")
	}

	for k, v := range vals {
		session.Values[k] = v
	}

	return nil
}
