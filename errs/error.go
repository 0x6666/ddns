package errs

import "errors"

var (
	ErrSqlitePathEmpty   = errors.New("sqlite file path empty")
	ErrNotImplement      = errors.New("not implement")
	ErrRecodeExist       = errors.New("recode already exist")
	ErrInvalidParam      = errors.New("invalid param")
	ErrDomianExist       = errors.New("domain already exist")
	ErrInvalidSession    = errors.New("invalid session")
	ErrPtrRecodeNotFound = errors.New("ptr recode not found")

	ErrKeyNotFound = errors.New("key not found")

	ErrDownloadUrlError = errors.New("download url error")
	ErrTaskAlreadyExist = errors.New("download task Already exist")
)
