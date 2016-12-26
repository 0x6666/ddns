package data

import "github.com/inimei/ddns/data/model"

type IDomains interface {
	NewDomain(userID int64, name string) (int64, error)
	UpdateDomain(domainID int64, newName string) error
	DeleteDomain(domainID int64) error
	GetDomain(domainID int64) (*model.Domain, error)
	GetDomains(userID int64) ([]*model.Domain, error)
	GetAllDomains(offset, limit int) ([]*model.Domain, error)

	GetRecodes(domainID int64, offset, limit int) ([]*model.Recode, error)
	NewRecode(domainID int64, r *model.Recode) (int64, error)
	GetRecode(id int64) (*model.Recode, error)
	DeleteRecode(id int64) error
}

type IDatabase interface {
	Init() error
	Close() error

	IDomains

	ReadData(offset, limit int) ([]*model.Recode, error)
	CreateRecode(r *model.Recode) (int64, error)
	FindByName(name string) (*model.Recode, error)
	FindByKey(key string) (*model.Recode, error)
	ClearRecodes(bSynced bool) error
	UpdateRecode(r *model.Recode) error

	GetVersion() int64
	SetVersion(v int64) //only slave sync

	BeginTransaction() (IDatabase, error)
	Rollback() error
	Commit() error
}
