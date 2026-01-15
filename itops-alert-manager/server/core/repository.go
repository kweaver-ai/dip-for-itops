package core

import (
	"database/sql"
)

type RepoError interface {
	GetError() error
	Type() string
	Error() string
}

type CtxKeyForDBTxnType string

var (
	Transaction                       = 0
	CtxKeyForDBTxn CtxKeyForDBTxnType = "DB_TXN"
)

type Repo struct {
	DB *sql.DB
}
