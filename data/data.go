package data

import "github.com/inimei/ddns/data/model"

type IDatabase interface {
	Init() error
	Close() error
	ReadData(offset, limit int) (map[string]interface{}, error)
	CreateRecode(r *model.Recode) (int64, error)
}
