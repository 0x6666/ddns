package container

import (
	"sync"
	"time"

	"github.com/0x6666/ddnsx/errs"
)

type value struct {
	val    string
	expire time.Time
}

type MemContainer struct {
	sync.RWMutex

	data map[string]value
}

func newMemContainer() *MemContainer {
	return &MemContainer{data: map[string]value{}}
}

// Get ....
func (mr *MemContainer) Get(key string) (string, error) {

	mr.Lock()
	defer mr.Unlock()

	if s, ok := mr.data[key]; ok {
		if s.expire.Before(time.Now()) {
			delete(mr.data, key)
		} else {
			return s.val, nil
		}
	}
	return "", errs.ErrKeyNotFound
}

func (mr *MemContainer) Set(key string, val string, expire int64) error {
	mr.Lock()
	defer mr.Unlock()

	mr.data[key] = value{val, time.Now().Add(time.Duration(expire) * time.Second)}

	return nil
}

func (mr *MemContainer) Delete(key string) error {

	mr.Lock()
	defer mr.Unlock()

	if _, ok := mr.data[key]; !ok {
		return nil
	}

	delete(mr.data, key)
	return nil
}

func (mr *MemContainer) IsExist(key string) bool {
	mr.RLock()
	defer mr.RUnlock()

	_, ok := mr.data[key]
	return ok
}

func (mr *MemContainer) Count() (int64, error) {
	mr.RLock()
	defer mr.RUnlock()

	return int64(len(mr.data)), nil
}
