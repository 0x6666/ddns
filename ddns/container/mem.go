package container

import (
	"errors"
	"sync"
)

type set map[string]bool

func (s set) toArray() []string {
	cnt := len(s)

	if cnt == 0 {
		return []string{}
	}

	ss := make([]string, 0, cnt)
	for k := range s {
		ss = append(ss, k)
	}

	return ss
}

type MemContainer struct {
	data map[string]set
	eof  bool
	mtx  *sync.Mutex
}

func newMemContainer() *MemContainer {
	return &MemContainer{map[string]set{}, false, &sync.Mutex{}}
}

//Get ....
func (mr *MemContainer) Get(key string) ([]string, error) {
	if s, ok := mr.data[key]; ok {
		return s.toArray(), nil
	}
	return []string{}, nil
}

func (mr *MemContainer) Delete(key string) error {
	if _, ok := mr.data[key]; !ok {
		return nil
	}

	mr.mtx.Lock()
	defer mr.mtx.Unlock()
	delete(mr.data, key)

	return nil
}

func (mr *MemContainer) Append(key, val string) error {
	mr.mtx.Lock()
	defer mr.mtx.Unlock()

	if _, ok := mr.data[key]; !ok {
		mr.data[key] = set{val: true}
	} else {
		mr.data[key][val] = true
	}
	return nil
}

func (mr *MemContainer) DeleteValue(key, val string) error {

	if _, ok := mr.data[key]; !ok {
		return nil
	}

	if _, ok := mr.data[key][val]; !ok {
		return nil
	}

	mr.mtx.Lock()
	defer mr.mtx.Unlock()

	delete(mr.data[key], val)
	return nil
}

func (mr *MemContainer) ValueCount(key string) (int, error) {
	if _, ok := mr.data[key]; !ok {
		return 0, nil
	}

	return len(mr.data[key]), nil
}

func (mr *MemContainer) IsExist(key string) bool {
	_, ok := mr.data[key]
	return ok
}

func (mr *MemContainer) ClearAll() error {
	mr.mtx.Lock()
	defer mr.mtx.Unlock()
	mr.data = map[string]set{}
	return nil
}

//内存的同从不大，所以不管多少一次搞定
func (mr *MemContainer) StartScan() {
	mr.eof = false
}

func (mr *MemContainer) ScanNext() ([]string, error) {

	if mr.eof {
		return nil, errors.New("Scan EOF")
	}

	mr.eof = true

	cnt := len(mr.data)
	ss := make([]string, 0, cnt)
	for k := range mr.data {
		ss = append(ss, k)
	}

	return ss, nil
}

func (mr *MemContainer) Eof() bool {
	return mr.eof
}
