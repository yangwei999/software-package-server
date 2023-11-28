package repository

import "github.com/opensourceways/software-package-server/softwarepkg/domain"

type Watch interface {
	Add([]string) error
	Save(*domain.PkgWatch) error
	FindAll() ([]*domain.PkgWatch, error)
}
