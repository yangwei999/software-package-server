package repositoryimpl

import (
	"time"

	"github.com/opensourceways/software-package-server/watch/domain"
)

const (
	fieldStatus = "status"
)

type SoftwarePkgPRDO struct {
	PkgId     string    `gorm:"column:pkg_id"`
	PRNum     int       `gorm:"column:pr_num"`
	PRLink    string    `gorm:"column:pr_link"`
	Status    string    `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (s *softwarePkgPR) toSoftwarePkgPRDO(pw *domain.PkgWatch) SoftwarePkgPRDO {
	return SoftwarePkgPRDO{
		PkgId:  pw.Id,
		PRNum:  pw.PR.Num,
		PRLink: pw.PR.Link,
		Status: pw.Status,
	}
}

func (do *SoftwarePkgPRDO) toDomainPkgWatch() *domain.PkgWatch {
	return &domain.PkgWatch{
		Id: do.PkgId,
		PR: domain.PullRequest{
			Num:  do.PRNum,
			Link: do.PRLink,
		},
		Status: do.Status,
	}
}
