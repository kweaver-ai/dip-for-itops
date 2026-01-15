package service

import "devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/core"

type serviceError struct {
	err     core.RepoError
	ErrType string
}

func (e *serviceError) Error() string {
	if e.err != nil {
		return e.err.Type()
	}
	return e.Type()
}

func (e *serviceError) GetError() core.RepoError {
	return e.err
}

func (e *serviceError) Type() string {
	return e.ErrType
}

func NewSvcInternalError(err core.RepoError) core.ServiceError {
	return &serviceError{
		err:     err,
		ErrType: "InternalError",
	}
}

func NewSvcNotFoundError(err core.RepoError) core.ServiceError {
	return &serviceError{
		err:     err,
		ErrType: "NotFound",
	}
}

func NewSvcNotFoundTemplate(err core.RepoError) core.ServiceError {
	return &serviceError{
		err:     err,
		ErrType: "NotFoundTemplate",
	}
}

func NewSvcNameSameError(err core.RepoError) core.ServiceError {
	return &serviceError{
		err:     err,
		ErrType: "NameExisted",
	}
}

func NewSvUnauthorizedError(err core.RepoError) core.ServiceError {
	return &serviceError{
		err:     err,
		ErrType: "Unauthorized",
	}
}

func NewSvcGenerateIDFailedError(err core.RepoError) core.ServiceError {
	return &serviceError{
		err:     err,
		ErrType: "GenerateIDFailed",
	}
}
