package sqlite

import (
	"strings"

	"github.com/yangsongfwd/backup/log"
	m "github.com/yangsongfwd/ddns/data/model"
	"github.com/yangsongfwd/ddns/errs"
)

func (s *SqliteDB) NewDomain(userID int64, d *m.Domain) (int64, error) {

	d.DomainName = strings.ToLower(d.DomainName)

	if err := s.db.First(&d, "domain_name = ?", d.DomainName).Error; err == nil {
		return -1, errs.ErrDomianExist
	}

	d.UserID = userID
	if err := s.db.Create(&d).Error; err != nil {
		return 0, err
	}
	go s.updateVersion()
	return d.ID, nil
}

func (s *SqliteDB) UpdateDomain(domainID int64, newName string) error {
	n := strings.ToLower(newName)

	var d m.Domain
	if err := s.db.First(&d, "domain_name = ?", n).Error; err == nil {
		//exist
		if d.ID == domainID {
			return nil
		}

		return errs.ErrDomianExist
	}

	if err := s.db.First(&d, domainID).Error; err != nil {
		log.Error(err.Error())
		return err
	}

	d.ID = domainID
	if err := s.db.Set("gorm:save_associations", false).Model(&d).Update("domain_name", n).Error; err != nil {
		log.Error(err.Error())
		return err
	}

	go s.updateVersion()
	return nil
}

func (s *SqliteDB) DeleteDomain(domainID int64) error {
	d := m.Domain{}
	var err error
	if err = s.db.First(&d, domainID).Error; err != nil {
		log.Error(err.Error())
		return err
	}

	db := s.db.Begin()
	defer func() {
		if err != nil {
			db.Rollback()
		} else {
			db.Commit()
		}
	}()

	err = db.Where(&m.Recode{DomainID: d.ID}).Delete(&m.Recode{}).Error
	if err != nil {
		log.Error(err.Error())
		return err
	}

	if err = db.Delete(&d).Error; err != nil {
		log.Error(err.Error())
		return err
	}
	go s.updateVersion()
	return nil
}

func (s *SqliteDB) GetDomain(domainID int64) (*m.Domain, error) {
	d := m.Domain{}
	if err := s.db.First(&d, domainID).Error; err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return &d, nil
}

func (s *SqliteDB) FindDomainByName(domain string) (*m.Domain, error) {
	d := m.Domain{}
	if err := s.db.First(&d, "domain_name = ?", domain).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *SqliteDB) GetDomains(userID int64) ([]*m.Domain, error) {
	ds := []*m.Domain{}
	err := s.db.Where(&m.Domain{UserID: userID}).Find(&ds).Error
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return ds, nil
}

func (s *SqliteDB) GetAllDomains(offset, limit int) ([]*m.Domain, error) {
	ds := []*m.Domain{}
	db := s.db.Set("gorm:save_associations", false).Offset(offset).Limit(limit).Find(&ds)
	if db.Error != nil {
		log.Error(db.Error.Error())
		return nil, db.Error
	}
	return ds, nil
}

func (s *SqliteDB) GetRecodes(domainID int64, offset, limit int) ([]*m.Recode, error) {

	d := &m.Domain{}
	d.ID = domainID
	rs := []*m.Recode{}
	db := s.db.Model(d).Offset(offset).Limit(limit).Related(&rs)
	if db.Error != nil {
		log.Error(db.Error.Error())
		return nil, db.Error
	}
	return rs, nil
}

func (s *SqliteDB) NewRecode(domainID int64, r *m.Recode) (int64, error) {

	if r == nil || domainID == 0 {
		return 0, errs.ErrInvalidParam
	}

	d := m.Domain{}
	if err := s.db.First(&d, domainID).Error; err != nil {
		log.Error(err.Error())
		return 0, err
	}

	r.DomainID = d.ID
	if err := s.db.Create(r).Error; err != nil {
		return 0, err
	}
	go s.updateVersion()
	return 0, nil
}

func (s *SqliteDB) GetRecode(id int64) (*m.Recode, error) {
	var r m.Recode
	err := s.db.First(&r, id).Error
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return &r, nil
}

func (s *SqliteDB) DeleteRecode(id int64) error {
	r := &m.Recode{}
	r.ID = id
	d := s.db.Delete(r)
	go s.updateVersion()
	return d.Error
}

func (s *SqliteDB) UpdateRecode(id int64, r *m.Recode) error {

	r.ID = id
	db := s.db.Save(r)
	if db.Error != nil {
		log.Error(db.Error.Error())
	}

	go s.updateVersion()
	return db.Error
}

func (s *SqliteDB) FindByName(domainID int64, name string) (*m.Recode, error) {
	d := &m.Domain{}
	d.ID = domainID

	var recode m.Recode
	if err := s.db.Model(d).First(&recode, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &recode, nil
}
