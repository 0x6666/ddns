package model

import (
	"database/sql"
	"time"
)

const CurrentVersion = "1.0"
const DefUserID int64 = 1

type RecodeType int

const (
	AAAA RecodeType = iota
	A
	CNAME
)

type Model struct {
	ID        int64 `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SchemaVersion struct {
	Model
	Version string `gorm:"column:version"`
}

func (SchemaVersion) TableName() string {
	return "schema_version"
}

type Version struct {
	SchemaVersion string `json:"schema_version"`
	DataVersion   int64  `json:"data_version"`
}

type User struct {
	Model
	Name    string   `gorm:"size:255"`
	Domains []Domain `gorm:"ForeignKey:UserID"`
}

func (User) TableName() string {
	return "users"
}

type Domain struct {
	Model
	UserID int64 `gorm:"index"`

	DomainName string   `gorm:"column:domain_name;unique_index"`
	Recodes    []Recode `gorm:"ForeignKey:DomainID"`
	Synced     bool     `gorm:"column:synced;default:'false'"` //for slave
}

func (Domain) TableName() string {
	return "domains"
}

type Recode struct {
	Model
	DomainID int64 `gorm:"column:domain_id;index"`

	Dynamic   bool           `gorm:"column:dynamic"` //是否未动态
	UpdateKey sql.NullString `gorm:"column:key;unique_index"`
	Synced    bool           `gorm:"column:synced;default:'false'"` //for slave

	RecordType  RecodeType `gorm:"unique_index:recode_unique;column:type;type:int"` // 1 ipv4
	RecordHost  string     `gorm:"unique_index:recode_unique;column:host"`
	RecodeValue string     `gorm:"unique_index:recode_unique;column:value"`
	TTL         int        `gorm:"column:ttl"`
}

func (Recode) TableName() string {
	return "recodes"
}
