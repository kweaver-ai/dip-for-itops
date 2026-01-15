package core

type ServiceError interface {
	Error() string
	GetError() RepoError
	Type() string
}
