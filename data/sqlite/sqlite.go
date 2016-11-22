package sqlite

import (
	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/config"
	"github.com/inimei/ddns/data/model"
	"github.com/inimei/ddns/errs"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type SqliteDB struct {
	db   *gorm.DB
	path string
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

	//defer db.Close()

	s.db = db.AutoMigrate(&model.User{}, &model.Recode{})

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

func (s *SqliteDB) FindByName(name string) (*model.Recode, error) {
	var recode model.Recode
	if err := s.db.First(&recode, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &recode, nil
}

func (s *SqliteDB) GetRecode(id int64) (*model.Recode, error) {
	var r model.Recode
	err := s.db.First(&r, id).Error
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return &r, nil
}

func (s *SqliteDB) DeleteRecode(id int64) error {
	r := &model.Recode{}
	r.ID = id
	return s.db.Delete(r).Error
}
