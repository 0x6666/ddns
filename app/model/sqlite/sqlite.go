package sqlite

import (
	"reflect"
	"runtime"
	"sync"

	"github.com/yangsongfwd/backup/log"
	"github.com/yangsongfwd/ddns/app/model"
	"github.com/yangsongfwd/ddns/config"
	"github.com/yangsongfwd/ddns/errs"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type SqliteDB struct {
	db   *gorm.DB
	path string

	version  int64
	mutex    sync.RWMutex
	listener map[string]model.OnDomainChanged
}

func NewSqlite() *SqliteDB {
	return &SqliteDB{}
}

func (s *SqliteDB) Init() error {

	path := config.Data.Sqlite.Path
	size := len(path)
	if size == 0 || path[size-1] == '/' {
		return errs.ErrSqlitePathEmpty
	}

	if path[0] != '/' {
		s.path = config.CurDir() + "/" + path
	}

	db, err := gorm.Open("sqlite3", s.path)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	if !db.HasTable(&model.SchemaVersion{}) {
		d := db.AutoMigrate(&model.SchemaVersion{})
		if d.Error != nil {
			log.Error("migrate failed: %v", d.Error.Error())
			return d.Error
		}
		version := model.SchemaVersion{Version: model.CurrentVersion}
		d = d.Create(&version)
		if d.Error != nil {
			log.Error("create schema version recode failed: %v", d.Error)
			return d.Error
		}
	} else {
		//TODO: check upgrade
	}

	//defer db.Close()

	s.db = db.AutoMigrate(&model.User{}, &model.Domain{}, &model.Recode{}, &model.SchemaVersion{})

	return db.Error
}

func (s *SqliteDB) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SqliteDB) CreateRecode(r *model.Recode) (int64, error) {
	if !s.db.NewRecord(r) {
		return 0, errs.ErrRecodeExist
	}
	if err := s.db.Create(r).Error; err != nil {
		return 0, err
	}
	go s.updateVersion()
	return r.ID, nil
}

func (s *SqliteDB) ReadData(offset, limit int) ([]*model.Recode, error) {
	rc := []*model.Recode{}
	db := s.db.Offset(offset).Limit(limit).Find(&rc)
	if db.Error != nil {
		log.Error(db.Error.Error())
		return nil, db.Error
	}
	return rc, nil
}

func (s *SqliteDB) FindByKey(key string) (*model.Recode, error) {
	var recode model.Recode
	if err := s.db.First(&recode, "key = ?", key).Error; err != nil {
		return nil, err
	}
	return &recode, nil
}

func (s *SqliteDB) ClearRecodes(bSynced bool) error {
	db := s.db
	if bSynced {
		db = db.Where("synced = ?", true)
	}
	err := db.Delete(&model.Recode{}).Error
	go s.updateVersion()
	return err
}

func (s *SqliteDB) GetVersion() int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.version
}

func (s *SqliteDB) updateVersion() {
	s.mutex.Lock()
	s.version = s.version + 1
	s.mutex.Unlock()

	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for _, l := range s.listener {
		l()
	}
}

func (s *SqliteDB) SetVersion(v int64) {
	if config.Data.Server.Master == false {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		s.version = v
	} else {
		log.Error("SetVersion for slave")
	}
}

func (s *SqliteDB) BeginTransaction() (model.IDatabase, error) {
	d := s.db.Begin()
	if d.Error != nil {
		return nil, d.Error
	}

	newDb := SqliteDB{}
	newDb.db = d
	newDb.version = s.GetVersion()
	newDb.path = s.path
	return &newDb, nil
}

func (s *SqliteDB) Rollback() error {
	return s.db.Rollback().Error
}

func (s *SqliteDB) Commit() error {
	err := s.db.Commit().Error
	go s.updateVersion()
	return err
}

//RegListener ...
func (s *SqliteDB) RegListener(l model.OnDomainChanged) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if l == nil {
		log.Warn("listener is nil")
		return
	}

	if s.listener == nil {
		s.listener = map[string]model.OnDomainChanged{}
	}

	funcName := runtime.FuncForPC(reflect.ValueOf(l).Pointer()).Name()

	for n := range s.listener {
		if funcName == n {
			return
		}
	}
	log.Debug("reg listener [%v]", funcName)

	s.listener[funcName] = l
}

//UnRegListener ...
func (s *SqliteDB) UnRegListener(l model.OnDomainChanged) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if l == nil {
		log.Warn("listener is nil")
		return
	}

	if s.listener == nil {
		return
	}

	funcName := runtime.FuncForPC(reflect.ValueOf(l).Pointer()).Name()

	for n := range s.listener {
		if funcName == n {
			log.Debug("unreg listener [%v]", funcName)
			delete(s.listener, funcName)
			return
		}
	}
}
