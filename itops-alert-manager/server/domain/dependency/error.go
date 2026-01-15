package dependency

import "devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/core"

type repoError struct {
	err     error
	ErrType string
}

func (e *repoError) GetError() error {
	return e.err
}

func (e *repoError) Type() string {
	return e.ErrType
}
func (e *repoError) Error() string {
	return e.err.Error()
}

func NewRepoInternalError(err error) core.RepoError {
	return &repoError{
		err:     err,
		ErrType: "InternalError",
	}
}

func NewRepoExecuteSqlError(err error) core.RepoError {
	return &repoError{
		err:     err,
		ErrType: "ExecuteSqlError",
	}
}

type restAPIError struct {
	err     error
	ErrType string
}

func (e *restAPIError) GetError() error {
	return e.err
}

func (e *restAPIError) Type() string {
	return e.ErrType
}

func (e *restAPIError) Error() string {
	return e.err.Error()
}

func NewClientRequestError(err error) core.RestAPIError {
	return &restAPIError{
		err:     err,
		ErrType: "ClientRequestError",
	}
}
