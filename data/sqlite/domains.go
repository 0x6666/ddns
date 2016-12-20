package sqlite

import (
	"strings"

	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/data/model"
	"github.com/inimei/ddns/errs"
)

func (s *SqliteDB) NewDomain(userID int64, name string) (int64, error) {

	n := strings.ToLower(name)

	var d model.Domain
	if err := s.db.First(&d, "domain_name = ?", n).Error; err == nil {
		return -1, errs.ErrDomianExist
	}

	d.DomainName = n
	if err := s.db.Create(&d).Error; err != nil {
		return 0, err
	}
	go s.updateVersion()
	return d.ID, nil
}

func (s *SqliteDB) UpdateDomain(domainID int64, newName string) error {
	n := strings.ToLower(newName)

	var d model.Domain
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
	d := model.Domain{}
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

	err = db.Where(&model.Recode{DomainID: d.ID}).Delete(&model.Recode{}).Error
	if err != nil {
		log.Error(err.Error())
		return err
	}

	if err = db.Delete(&d).Error; err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (s *SqliteDB) GetDomains(userID int64) ([]*model.Domain, error) {
	ds := []*model.Domain{}
	err := s.db.Where(&model.Domain{UserID: userID}).Find(&ds).Error
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return ds, nil
}

func (s *SqliteDB) GetAllDomains(offset, limit int) ([]*model.Domain, error) {
	ds := []*model.Domain{}
	db := s.db.Offset(offset).Limit(limit).Find(&ds)
	if db.Error != nil {
		log.Error(db.Error.Error())
		return nil, db.Error
	}
	return ds, nil
}
