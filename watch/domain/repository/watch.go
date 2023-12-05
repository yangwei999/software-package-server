package repository

import "github.com/opensourceways/software-package-server/watch/domain"

type Watch interface {
	Add(pw *domain.PkgWatch) error
	Save(*domain.PkgWatch) error
	FindAll() ([]*domain.PkgWatch, error)
}
