package repositoryimpl

import (
	"github.com/google/uuid"

	"github.com/opensourceways/software-package-server/common/infrastructure/postgresql"
	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/repository"
)

type softwarePkgPR struct {
	cli dbClient
}

func NewSoftwarePkgPR(table *Table) repository.Watch {
	return softwarePkgPR{cli: postgresql.NewDBTable(table.SoftwarePkgPR)}
}

func (s softwarePkgPR) Add(pkgIds []string) error {
	var do SoftwarePkgPRDO
	if err = s.toSoftwarePkgPRDO(p, u, &do); err != nil {
		return err
	}

	filter := SoftwarePkgPRDO{PkgId: u}

	return s.cli.Insert(&filter, &do)
}

func (s softwarePkgPR) Save(*domain.PkgWatch) error {
	u, err := uuid.Parse(p.Id)
	if err != nil {
		return err
	}
	filter := SoftwarePkgPRDO{PkgId: u}

	var do SoftwarePkgPRDO
	if err = s.toSoftwarePkgPRDO(p, u, &do); err != nil {
		return err
	}

	return s.cli.UpdateRecord(&filter, &do)
}

func (s softwarePkgPR) FindAll() ([]*domain.PkgWatch, error) {
	filter := SoftwarePkgPRDO{}

	var res []SoftwarePkgPRDO

	if err := s.cli.GetRecords(
		&filter,
		&res,
		postgresql.Pagination{},
		nil,
	); err != nil {
		return nil, err
	}

	var p = make([]domain.SoftwarePkg, len(res))

	for i := range res {
		v, err := res[i].toDomainPullRequest()
		if err != nil {
			return nil, err
		}

		p[i] = v
	}

	return p, nil
}
