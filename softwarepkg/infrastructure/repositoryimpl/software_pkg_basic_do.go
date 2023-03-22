package repositoryimpl

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/plugin/optimisticlock"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
)

const (
	fieldAppliedAt = "applied_at"
	fieldVersion   = "version"
	fieldId        = "uuid"
	frozenStatus   = 1
	unfrozenStatus = 2
)

func (s softwarePkgBasic) toSoftwarePkgBasicDO(pkg *domain.SoftwarePkgBasicInfo, do *SoftwarePkgBasicDO) {
	app := &pkg.Application

	*do = SoftwarePkgBasicDO{
		Id:              uuid.New(),
		PackageName:     pkg.PkgName.PackageName(),
		Importer:        pkg.Importer.Account(),
		Phase:           pkg.Phase.PackagePhase(),
		SourceCode:      app.SourceCode.Address.URL(),
		License:         app.SourceCode.License.License(),
		PackageDesc:     app.PackageDesc.PackageDesc(),
		PackagePlatform: app.PackagePlatform.PackagePlatform(),
		Sig:             app.ImportingPkgSig.ImportingPkgSig(),
		ReasonToImport:  app.ReasonToImportPkg.ReasonToImportPkg(),
		AppliedAt:       pkg.AppliedAt,
		UpdatedAt:       pkg.AppliedAt,
		ApprovedBy:      toStringArray(pkg.ApprovedBy),
		RejectedBy:      toStringArray(pkg.RejectedBy),
	}

	if pkg.Frozen {
		do.Frozen = frozenStatus
	} else {
		do.Frozen = unfrozenStatus
	}

	if pkg.RepoLink != nil {
		do.RepoLink = pkg.RepoLink.URL()
	}

	if pkg.RelevantPR != nil {
		do.RelevantPR = pkg.RelevantPR.URL()
	}

	do.PRNum = pkg.PRNum
}

type SoftwarePkgBasicDO struct {
	// must set "uuid" as the name of column
	Id              uuid.UUID              `gorm:"column:uuid;type:uuid"`
	PackageName     string                 `gorm:"column:package_name"`
	Importer        string                 `gorm:"column:importer"`
	RepoLink        string                 `gorm:"column:repo_link"`
	Phase           string                 `gorm:"column:phase"`
	SourceCode      string                 `gorm:"column:source_code"`
	License         string                 `gorm:"column:license"`
	PackageDesc     string                 `gorm:"column:package_desc"`
	PackagePlatform string                 `gorm:"column:package_platform"`
	RelevantPR      string                 `gorm:"column:relevant_pr"`
	PRNum           int                    `gorm:"column:pr_num"`
	Sig             string                 `gorm:"column:sig"`
	ReasonToImport  string                 `gorm:"column:reason_to_import"`
	ApprovedBy      pq.StringArray         `gorm:"column:approvedby;type:text[];default:'{}'"`
	RejectedBy      pq.StringArray         `gorm:"column:rejectedby;type:text[];default:'{}'"`
	AppliedAt       int64                  `gorm:"column:applied_at"`
	UpdatedAt       int64                  `gorm:"column:updated_at"`
	Frozen          int                    `gorm:"column:frozen"`
	Version         optimisticlock.Version `gorm:"column:version"`
}

func (do *SoftwarePkgBasicDO) toSoftwarePkgBasicInfo() (info domain.SoftwarePkgBasicInfo, err error) {
	info.Id = do.Id.String()

	if info.PkgName, err = dp.NewPackageName(do.PackageName); err != nil {
		return
	}

	if do.RepoLink != "" {
		if info.RepoLink, err = dp.NewURL(do.RepoLink); err != nil {
			return
		}
	}

	if do.RelevantPR != "" {
		if info.RelevantPR, err = dp.NewURL(do.RelevantPR); err != nil {
			return
		}
	}

	if info.Importer, err = dp.NewAccount(do.Importer); err != nil {
		return
	}

	if info.Phase, err = dp.NewPackagePhase(do.Phase); err != nil {
		return
	}

	info.AppliedAt = do.AppliedAt

	if err = do.toSoftwarePkgApplication(&info.Application); err != nil {
		return
	}

	if info.ApprovedBy, err = do.toAccounts(do.ApprovedBy); err != nil {
		return
	}

	info.RejectedBy, err = do.toAccounts(do.RejectedBy)

	info.Frozen = do.Frozen == frozenStatus

	return
}

func (do *SoftwarePkgBasicDO) toAccounts(v []string) (r []dp.Account, err error) {
	if len(v) == 0 {
		return
	}

	r = make([]dp.Account, len(v))
	for i := range v {
		if r[i], err = dp.NewAccount(v[i]); err != nil {
			return
		}
	}

	return
}

func (do *SoftwarePkgBasicDO) toSoftwarePkgApplication(app *domain.SoftwarePkgApplication) (err error) {
	if app.ReasonToImportPkg, err = dp.NewReasonToImportPkg(do.ReasonToImport); err != nil {
		return
	}

	if app.PackageDesc, err = dp.NewPackageDesc(do.PackageDesc); err != nil {
		return
	}

	if app.PackagePlatform, err = dp.NewPackagePlatform(do.PackagePlatform); err != nil {
		return
	}

	if app.ImportingPkgSig, err = dp.NewImportingPkgSig(do.Sig); err != nil {
		return
	}

	if app.SourceCode.License, err = dp.NewLicense(do.License); err != nil {
		return
	}

	app.SourceCode.Address, err = dp.NewURL(do.SourceCode)

	return
}

func toStringArray(v []dp.Account) (arr pq.StringArray) {
	arr = make(pq.StringArray, len(v))
	for i, account := range v {
		arr[i] = account.Account()
	}

	return
}
