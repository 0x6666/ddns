package data

import "github.com/inimei/ddns/data/model"

type OnDomainChanged func()

type IDomains interface {
	NewDomain(userID int64, domain *model.Domain) (int64, error)
	UpdateDomain(domainID int64, newName string) error
	DeleteDomain(domainID int64) error
	GetDomain(domainID int64) (*model.Domain, error)
	FindDomainByName(domain string) (*model.Domain, error)
	GetDomains(userID int64) ([]*model.Domain, error)
	GetAllDomains(offset, limit int) ([]*model.Domain, error)

	GetRecodes(domainID int64, offset, limit int) ([]*model.Recode, error)
	NewRecode(domainID int64, r *model.Recode) (int64, error)
	GetRecode(id int64) (*model.Recode, error)
	DeleteRecode(id int64) error
	UpdateRecode(id int64, r *model.Recode) error
	FindByName(domainID int64, name string) (*model.Recode, error)
}

type IDatabase interface {
	Init() error
	Close() error

	IDomains

	FindByKey(key string) (*model.Recode, error)

	ClearRecodes(bSynced bool) error

	GetVersion() int64
	SetVersion(v int64) //only slave sync

	BeginTransaction() (IDatabase, error)
	Rollback() error
	Commit() error

	RegListener(l OnDomainChanged)
	UnRegListener(l OnDomainChanged)
}
