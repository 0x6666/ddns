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
	db *gorm.DB
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
		path = config.CurDir() + "/" + path
	}

	db, err := gorm.Open("sqlite3", path)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	defer db.Close()

	return db.AutoMigrate(&model.User{}, &model.Recode{}).Error
}

func (s *SqliteDB) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SqliteDB) CreateRecode(r *model.Recode) (int64, error) {
	return 0, errs.ErrNotImplement
}

func (s *SqliteDB) ReadData(offset, limit int) (map[string]interface{}, error) {
	return nil, errs.ErrNotImplement
}
