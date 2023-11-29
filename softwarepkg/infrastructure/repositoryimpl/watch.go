package repositoryimpl

import (
	"time"

	"github.com/google/uuid"

	"github.com/opensourceways/software-package-server/common/infrastructure/postgresql"
	"github.com/opensourceways/software-package-server/softwarepkg/domain"
)

type softwarePkgPR struct {
	cli dbClient
}

func NewSoftwarePkgPR(table *Table) *softwarePkgPR {
	return &softwarePkgPR{cli: postgresql.NewDBTable(table.WatchCommunityPR)}
}

func (s *softwarePkgPR) Add(pw *domain.PkgWatch) error {
	u, err := uuid.Parse(pw.Id)
	if err != nil {
		return err
	}

	filter := SoftwarePkgPRDO{PkgId: u}

	do := s.toSoftwarePkgPRDO(pw, u)
	now := time.Now()
	do.CreatedAt = now
	do.UpdatedAt = now

	err = s.cli.Insert(&filter, &do)
	if s.cli.IsRowExists(err) {
		return nil
	}

	return err
}

func (s *softwarePkgPR) Save(pw *domain.PkgWatch) error {
	u, err := uuid.Parse(pw.Id)
	if err != nil {
		return err
	}
	filter := SoftwarePkgPRDO{PkgId: u}

	do := s.toSoftwarePkgPRDO(pw, u)
	do.UpdatedAt = time.Now()

	return s.cli.UpdateRecord(&filter, &do)
}

func (s *softwarePkgPR) FindAll() ([]*domain.PkgWatch, error) {
	var res []SoftwarePkgPRDO
	err := s.cli.GetRecords(
		[]postgresql.ColumnFilter{},
		&res,
		postgresql.Pagination{},
		nil,
	)
	if err != nil {
		return nil, err
	}

	var p = make([]*domain.PkgWatch, len(res))
	for i := range res {
		p[i] = res[i].toDomainPkgWatch()
	}

	return p, nil
}
