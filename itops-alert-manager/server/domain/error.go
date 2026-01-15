package domain

import (
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/core"
)

type DomainError interface {
	Error() string
	GetError() core.RepoError
	Type() string
}

type domainError struct {
	err     core.RepoError
	ErrType string
}

func (e *domainError) Error() string {
	if e.err != nil {
		return e.err.Type()
	}
	return e.Type()
}

func (e *domainError) GetError() core.RepoError {
	return e.err
}

func (e *domainError) Type() string {
	return e.ErrType
}

func NewInternalError(err core.RepoError) DomainError {
	return &domainError{
		err:     err,
		ErrType: "InternalError",
	}
}
