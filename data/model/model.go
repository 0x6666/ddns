package model

import (
	"database/sql"
	"time"
)

const CurrentVersion = "0.1"

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
	Name    string `gorm:"size:255"`
	Recodes []Recode
}

func (User) TableName() string {
	return "users"
}

type Recode struct {
	Model
	UserID int `gorm:"index"`

	Dynamic     bool           `gorm:"column:dynamic"` //是否未动态
	UpdateKey   sql.NullString `gorm:"column:key;unique_index"`
	RecordType  int            `gorm:"column:type"` // 1 ipv4
	RecordName  string         `gorm:"column:name"`
	RecodeValue string         `gorm:"column:value"`
	TTL         int            `gorm:"column:ttl"`
	Synced      bool           `gorm:"column:synced;default:'false'"`
}

func (Recode) TableName() string {
	return "recodes"
}
