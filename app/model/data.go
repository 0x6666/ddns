package model

type OnDomainChanged func()

type IDomains interface {
	NewDomain(userID int64, domain *Domain) (int64, error)
	UpdateDomain(domainID int64, newName string) error
	DeleteDomain(domainID int64) error
	GetDomain(domainID int64) (*Domain, error)
	FindDomainByName(domain string) (*Domain, error)
	GetDomains(userID int64) ([]*Domain, error)
	GetAllDomains(offset, limit int) ([]*Domain, error)

	GetRecodes(domainID int64, offset, limit int) ([]*Recode, error)
	NewRecode(domainID int64, r *Recode) (int64, error)
	GetRecode(id int64) (*Recode, error)
	DeleteRecode(id int64) error
	UpdateRecode(id int64, r *Recode) error
	FindByName(domainID int64, name string) (*Recode, error)
}

type IDatabase interface {
	Init() error
	Close() error

	IDomains

	FindByKey(key string) (*Recode, error)

	ClearRecodes(bSynced bool) error

	GetVersion() int64
	SetVersion(v int64) //only slave sync

	BeginTransaction() (IDatabase, error)
	Rollback() error
	Commit() error

	RegListener(l OnDomainChanged)
	UnRegListener(l OnDomainChanged)
}
