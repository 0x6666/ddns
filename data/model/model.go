package model

import (
	"database/sql"
	"time"
)

type Model struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
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

	Dynamic   bool           //是否未动态
	UpdateKey sql.NullString `gorm:"unique_index"`

	RecordType  int
	RecordName  string
	RecodeValue string
	TTL         int
}

func (Recode) TableName() string {
	return "recodes"
}
