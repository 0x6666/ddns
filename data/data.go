package data

import "github.com/inimei/ddns/data/model"

type IDatabase interface {
	Init() error
	Close() error

	ReadData(offset, limit int) ([]*model.Recode, error)
	CreateRecode(r *model.Recode) (int64, error)
	FindByName(name string) (*model.Recode, error)
	FindByKey(key string) (*model.Recode, error)
	GetRecode(id int64) (*model.Recode, error)
	DeleteRecode(id int64) error
	ClearRecodes() error
	UpdateRecode(r *model.Recode) error

	GetVersion() int64
	SetVersion(v int64) //only slave sync

	BeginTransaction() (IDatabase, error)
	Rollback() error
	Commit() error
}
