package repositoryimpl

import (
	"time"

	"github.com/google/uuid"

	"github.com/opensourceways/software-package-server/watch/domain"
)

type SoftwarePkgPRDO struct {
	// must set "uuid" as the name of column
	PkgId     uuid.UUID `gorm:"column:pkg_id;type:uuid"`
	PRNum     int       `gorm:"column:pr_num"`
	PRLink    string    `gorm:"column:pr_link"`
	Status    string    `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (s *softwarePkgPR) toSoftwarePkgPRDO(pw *domain.PkgWatch, u uuid.UUID) SoftwarePkgPRDO {
	return SoftwarePkgPRDO{
		PkgId:  u,
		PRNum:  pw.PR.Num,
		PRLink: pw.PR.Link,
		Status: pw.Status,
	}
}

func (do *SoftwarePkgPRDO) toDomainPkgWatch() *domain.PkgWatch {
	return &domain.PkgWatch{
		Id: do.PkgId.String(),
		PR: domain.PullRequest{
			Num:  do.PRNum,
			Link: do.PRLink,
		},
		Status: do.Status,
	}
}
