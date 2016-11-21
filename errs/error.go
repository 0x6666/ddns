package errs

import "errors"

var (
	ErrSqlitePathEmpty = errors.New("sqlite file path empty")
	ErrNotImplement    = errors.New("not implement")
	ErrRecodeExist     = errors.New("recode already exist")
	ErrInvalidParam    = errors.New("invalid param")
)
