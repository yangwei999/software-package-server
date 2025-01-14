package repository

import (
	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
)

type OptToFindSoftwarePkgs struct {
	Phase    dp.PackagePhase
	PkgName  dp.PackageName
	Platform dp.PackagePlatform
	Importer dp.Account

	PageNum      int
	CountPerPage int
}

type TranslatedReviewCommentIndex struct {
	PkgId     string
	CommentId string
	Language  dp.Language
}

type SoftwarePkg interface {
	HasSoftwarePkg(dp.PackageName) (bool, error)

	// AddSoftwarePkg adds a new pkg
	AddSoftwarePkg(*domain.SoftwarePkgBasicInfo) error

	SaveSoftwarePkg(pkg *domain.SoftwarePkgBasicInfo, version int) error

	FindSoftwarePkgBasicInfo(pid string) (domain.SoftwarePkgBasicInfo, int, error)

	FindSoftwarePkg(pid string) (domain.SoftwarePkg, int, error)

	FindSoftwarePkgs(OptToFindSoftwarePkgs) (r []domain.SoftwarePkgBasicInfo, total int, err error)

	AddReviewComment(pid string, comment *domain.SoftwarePkgReviewComment) error
	FindReviewComment(pid, commentId string) (domain.SoftwarePkgReviewComment, error)

	AddTranslatedReviewComment(pid string, comment *domain.SoftwarePkgTranslatedReviewComment) error
	FindTranslatedReviewComment(*TranslatedReviewCommentIndex) (domain.SoftwarePkgTranslatedReviewComment, error)

	AddOperationLog(*domain.SoftwarePkgOperationLog) error
}
