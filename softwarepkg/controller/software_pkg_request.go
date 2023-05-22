package controller

import (
	"github.com/opensourceways/software-package-server/softwarepkg/app"
	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
)

const (
	pageNum      = 1
	countPerPage = 10
)

type softwarePkgRequest struct {
	SpecUrl         string `json:"spec_url"        binding:"required"`
	SrcRPMURL       string `json:"src_rpm_url"     binding:"required"`
	PackageName     string `json:"pkg_name"        binding:"required"`
	PackageDesc     string `json:"desc"            binding:"required"`
	PackageSig      string `json:"sig"             binding:"required"`
	PackageReason   string `json:"reason"          binding:"required"`
	PackagePlatform string `json:"platform"        binding:"required"`
	Upstream        string `json:"upstream"        binding:"required"`
}

func (s softwarePkgRequest) toCmd(importer *domain.User) (
	cmd app.CmdToApplyNewSoftwarePkg, err error,
) {
	cmd.Importer = *importer

	cmd.PkgName, err = dp.NewPackageName(s.PackageName)
	if err != nil {
		return
	}

	application := &cmd.Application

	application.SourceCode.SpecURL, err = dp.NewURL(s.SpecUrl)
	if err != nil {
		return
	}

	application.SourceCode.SrcRPMURL, err = dp.NewURL(s.SrcRPMURL)
	if err != nil {
		return
	}

	application.ImportingPkgSig, err = dp.NewImportingPkgSig(s.PackageSig)
	if err != nil {
		return
	}

	application.ReasonToImportPkg, err = dp.NewReasonToImportPkg(s.PackageReason)
	if err != nil {
		return
	}

	application.PackageDesc, err = dp.NewPackageDesc(s.PackageDesc)
	if err != nil {
		return
	}

	application.Upstream, err = dp.NewURL(s.Upstream)
	if err != nil {
		return
	}

	application.PackagePlatform, err = dp.NewPackagePlatform(s.PackagePlatform)

	return
}

type softwarePkgListQuery struct {
	Phase        string `json:"phase"          form:"phase"`
	PkgName      string `json:"pkg_name"       form:"pkg_name"`
	Importer     string `json:"importer"       form:"importer"`
	Platform     string `json:"platform"       form:"platform"`
	PageNum      int    `json:"page_num"       form:"page_num"`
	CountPerPage int    `json:"count_per_page" form:"count_per_page"`
}

func (s softwarePkgListQuery) toCmd() (pkg app.CmdToListPkgs, err error) {
	if s.Importer != "" {
		if pkg.Importer, err = dp.NewAccount(s.Importer); err != nil {
			return
		}
	}

	if s.Phase != "" {
		if pkg.Phase, err = dp.NewPackagePhase(s.Phase); err != nil {
			return
		}
	}

	if s.Platform != "" {
		if pkg.Platform, err = dp.NewPackagePlatform(s.Platform); err != nil {
			return
		}
	}

	if s.PkgName != "" {
		if pkg.PkgName, err = dp.NewPackageName(s.PkgName); err != nil {
			return
		}
	}

	if s.PageNum > 0 {
		pkg.PageNum = s.PageNum
	} else {
		pkg.PageNum = pageNum
	}

	if s.CountPerPage > 0 {
		pkg.CountPerPage = s.CountPerPage
	} else {
		pkg.CountPerPage = countPerPage
	}

	return
}

type reviewCommentRequest struct {
	Comment string `json:"comment" binding:"required"`
}

func (r reviewCommentRequest) toCmd(user *domain.User) (rc app.CmdToWriteSoftwarePkgReviewComment, err error) {
	rc.Author = user.Account

	rc.Content, err = dp.NewReviewComment(r.Comment)

	return
}

type translationCommentRequest struct {
	Language string `json:"language"`
}

func (t translationCommentRequest) toCmd(pkgId, commentId string) (cmd app.CmdToTranslateReviewComment, err error) {
	cmd.PkgId = pkgId
	cmd.CommentId = commentId
	cmd.Language, err = dp.NewLanguage(t.Language)

	return
}
