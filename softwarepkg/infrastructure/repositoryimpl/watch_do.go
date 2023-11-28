package repositoryimpl

import (
	"time"

	"github.com/google/uuid"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/utils"
)

type SoftwarePkgPRDO struct {
	// must set "uuid" as the name of column
	PkgId     uuid.UUID `gorm:"column:uuid;type:uuid"`
	PRNum     int       `gorm:"column:pr_num"`
	Status    string    `gorm:"column:status"`
	CreatedAt int64     `gorm:"column:created_at"`
	UpdatedAt int64     `gorm:"column:updated_at"`
}

func (s softwarePkgPR) toSoftwarePkgPRDO(p *domain.SoftwarePkg, id uuid.UUID, do *SoftwarePkgPRDO) error {
	email, err := toEmailDO(p.Importer.Email)
	if err != nil {
		return err
	}

	*do = SoftwarePkgPRDO{
		PkgId:         id,
		Num:           p.PullRequest.Num,
		CIPRNum:       p.CIPRNum,
		Status:        p.Status,
		Link:          p.PullRequest.Link,
		PkgName:       p.Name,
		PkgDesc:       p.Application.PackageDesc,
		Upstream:      p.Application.Upstream,
		Sig:           p.Application.ImportingPkgSig,
		ImporterName:  p.Importer.Name,
		ImporterEmail: email,
		SpecURL:       p.Application.SourceCode.SpecURL,
		SrcRPMURL:     p.Application.SourceCode.SrcRPMURL,
		CreatedAt:     time.Now().Unix(),
		UpdatedAt:     time.Now().Unix(),
	}

	return nil
}

func (do *SoftwarePkgPRDO) toDomainPullRequest() (pkg domain.SoftwarePkg, err error) {
	if pkg.Importer.Email, err = toEmail(do.ImporterEmail); err != nil {
		return
	}

	pkg.PullRequest.Link = do.Link
	pkg.PullRequest.Num = do.Num
	pkg.CIPRNum = do.CIPRNum
	pkg.Status = do.Status
	pkg.Name = do.PkgName
	pkg.Id = do.PkgId.String()
	pkg.Importer.Name = do.ImporterName
	pkg.Application.SourceCode.SpecURL = do.SpecURL
	pkg.Application.SourceCode.SrcRPMURL = do.SrcRPMURL
	pkg.Application.PackageDesc = do.PkgDesc
	pkg.Application.Upstream = do.Upstream
	pkg.Application.ImportingPkgSig = do.Sig

	return
}

func toEmailDO(email string) (string, error) {
	return utils.Encryption.Encrypt([]byte(email))
}

func toEmail(e string) (string, error) {
	v, err := utils.Encryption.Decrypt(e)
	if err != nil {
		return "", err
	}

	return string(v), nil
}
