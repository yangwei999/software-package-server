package repositoryimpl

import (
	"time"

	"github.com/opensourceways/software-package-server/common/infrastructure/postgresql"
	"github.com/opensourceways/software-package-server/watch/domain"
)

type softwarePkgPR struct {
	cli dbClient
}

func NewSoftwarePkgPR(table *Table) *softwarePkgPR {
	return &softwarePkgPR{cli: postgresql.NewDBTable(table.WatchCommunityPR)}
}

func (s *softwarePkgPR) Add(pw *domain.PkgWatch) error {
	filter := SoftwarePkgPRDO{PkgId: pw.Id}

	do := s.toSoftwarePkgPRDO(pw)
	now := time.Now()
	do.CreatedAt = now
	do.UpdatedAt = now

	err := s.cli.Insert(&filter, &do)
	if s.cli.IsRowExists(err) {
		return nil
	}

	return err
}

func (s *softwarePkgPR) Save(pw *domain.PkgWatch) error {
	filter := SoftwarePkgPRDO{PkgId: pw.Id}

	do := s.toSoftwarePkgPRDO(pw)
	do.UpdatedAt = time.Now()

	return s.cli.UpdateRecord(&filter, &do)
}

func (s *softwarePkgPR) FindAll() ([]*domain.PkgWatch, error) {
	var res []SoftwarePkgPRDO

	err := s.cli.DBInstance().Where(fieldStatus+" IN ?", domain.PkgStatusNeedToHandle).Find(&res).Error
	if err != nil {
		return nil, err
	}

	var p = make([]*domain.PkgWatch, len(res))
	for i := range res {
		p[i] = res[i].toDomainPkgWatch()
	}

	return p, nil
}
