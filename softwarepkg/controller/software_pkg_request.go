package controller

import (
	"errors"

	"github.com/opensourceways/software-package-server/softwarepkg/app"
	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/useradapter"
)

const (
	pageNum      = 1
	countPerPage = 10
)

// softwarePkgRequest
type softwarePkgRequest struct {
	Name     string `json:"pkg_name"    binding:"required"`
	Desc     string `json:"desc"        binding:"required"`
	Purpose  string `json:"reason"      binding:"required"`
	Upstream string `json:"upstream"    binding:"required"`

	Spec string `json:"spec_url"     binding:"required"`
	SRPM string `json:"src_rpm_url"  binding:"required"`

	Sig string `json:"sig"          binding:"required"`

	softwarePkgRepoRequest
}

func (req *softwarePkgRequest) toBasic() (basic domain.SoftwarePkgBasicInfo, err error) {
	basic.Name, err = dp.NewPackageName(req.Name)
	if err != nil {
		return
	}

	basic.Desc, err = dp.NewPackageDesc(req.Desc)
	if err != nil {
		return
	}

	basic.Purpose, err = dp.NewPurposeToImportPkg(req.Purpose)
	if err != nil {
		return
	}

	basic.Upstream, err = dp.NewURL(req.Upstream)

	return
}

func (req *softwarePkgRequest) toCmd(importer *domain.User, ua useradapter.UserAdapter) (
	cmd app.CmdToApplyNewSoftwarePkg, err error,
) {
	cmd.Importer = *importer

	if cmd.Basic, err = req.toBasic(); err != nil {
		return
	}

	cmd.Spec, err = dp.NewURL(req.Spec)
	if err != nil {
		return
	}

	cmd.SRPM, err = dp.NewURL(req.SRPM)
	if err != nil {
		return
	}

	cmd.Sig, err = dp.NewImportingPkgSig(req.Sig)
	if err != nil {
		return
	}

	cmd.Repo, err, _ = req.toRepo(importer, ua)

	return
}

// reqToUpdateSoftwarePkg
type reqToUpdateSoftwarePkg struct {
	Desc     string `json:"desc"`
	Purpose  string `json:"reason"`
	Upstream string `json:"upstream"`

	Spec string `json:"spec_url"`
	SRPM string `json:"src_rpm_url"`

	Sig        string   `json:"sig"`
	RepoLink   string   `json:"repo_link"`
	Committers []string `json:"committers"`
}

func (req *reqToUpdateSoftwarePkg) toCmd(
	pkgId string, importer *domain.User, ua useradapter.UserAdapter,
) (
	cmd app.CmdToUpdateSoftwarePkgApplication, err error,
) {
	cmd.PkgId = pkgId
	cmd.Importer = *importer

	if req.Desc != "" {
		if cmd.Desc, err = dp.NewPackageDesc(req.Desc); err != nil {
			return
		}
	}

	if req.Purpose != "" {
		cmd.Purpose, err = dp.NewPurposeToImportPkg(req.Purpose)
		if err != nil {
			return
		}
	}

	if req.Upstream != "" {
		if cmd.Upstream, err = dp.NewURL(req.Upstream); err != nil {
			return
		}
	}

	if req.Spec != "" {
		if cmd.Spec, err = dp.NewURL(req.Spec); err != nil {
			return
		}
	}

	if req.SRPM != "" {
		if cmd.SRPM, err = dp.NewURL(req.SRPM); err != nil {
			return
		}
	}

	if req.Sig != "" {
		if cmd.Sig, err = dp.NewImportingPkgSig(req.Sig); err != nil {
			return
		}
	}

	if req.RepoLink == "" && len(req.Committers) == 0 {
		return
	}

	if !(req.RepoLink != "" && len(req.Committers) != 0) {
		err = errors.New("repo_link and committers must be set at same time")

		return
	}

	v := softwarePkgRepoRequest{
		RepoLink:   req.RepoLink,
		Committers: req.Committers,
	}

	cmd.Repo, err, _ = v.toRepo(importer, ua)

	return
}

// softwarePkgListQuery
type softwarePkgListQuery struct {
	Phase        string `json:"phase"          form:"phase"`
	PkgName      string `json:"pkg_name"       form:"pkg_name"`
	Importer     string `json:"importer"       form:"importer"`
	Platform     string `json:"platform"       form:"platform"`
	LastId       string `json:"last_id"        form:"last_id"`
	Count        bool   `json:"count"       form:"count"`
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

	if s.LastId != "" && s.PageNum > 0 {
		err = errors.New("it can't set last_id and page_num at same time")

		return
	}

	if s.LastId == "" {
		if s.PageNum <= 0 || s.PageNum > config.MaxPageNum {
			s.PageNum = pageNum
		}
		pkg.PageNum = s.PageNum
	}

	if s.CountPerPage <= 0 || s.CountPerPage > config.MaxCountPerPage {
		s.CountPerPage = countPerPage
	}
	pkg.CountPerPage = s.CountPerPage

	return
}

// reqToAbandonPkg
type reqToAbandonPkg struct {
	Comment string `json:"comment"`
}

func (req *reqToAbandonPkg) toCmd(pkgId string, user *domain.User) (cmd app.CmdToAbandonPkg, err error) {
	if req.Comment != "" {
		if cmd.Comment, err = dp.NewReviewComment(req.Comment); err != nil {
			return
		}
	}

	cmd.PkgId = pkgId
	cmd.Importer = user.Account

	return
}

// reviewRequest
type reviewRequest struct {
	Reviews []checkItemReviewInfo `json:"reviews" binding:"required"`
}

func (req *reviewRequest) toCmd() (reviews []domain.CheckItemReviewInfo, err error) {
	reviews = make([]domain.CheckItemReviewInfo, len(req.Reviews))

	cs := make([]string, len(req.Reviews))
	for i := range req.Reviews {
		if reviews[i], err = req.Reviews[i].toInfo(); err != nil {
			return
		}

		cs[i] = req.Reviews[i].Comment
	}

	err = dp.CheckMultiComments(cs)

	return
}

type checkItemReviewInfo struct {
	Id      string `json:"id"       binding:"required"`
	Pass    bool   `json:"pass"`
	Comment string `json:"comment"`
}

func (req *checkItemReviewInfo) toInfo() (info domain.CheckItemReviewInfo, err error) {
	if !req.Pass && req.Comment == "" {
		err = errors.New("lack of comment")

		return
	}

	c, err := dp.NewCheckItemComment(req.Comment)
	if err != nil {
		return
	}

	info.Id = req.Id
	info.Pass = req.Pass
	info.Comment = c

	return
}

// softwarePkgRepoRequest
type softwarePkgRepoRequest struct {
	RepoLink   string   `json:"repo_link"    binding:"required"`
	Committers []string `json:"committers"   binding:"required"`
}

func (req *softwarePkgRepoRequest) toRepo(importer *domain.User, ua useradapter.UserAdapter) (
	repo domain.SoftwarePkgRepo, err error, invalidCommitter []string,
) {
	if len(req.Committers) > config.MaxNumOfCommitters {
		err = errors.New("too many committers")

		return
	}

	m := map[string]bool{}
	for _, c := range req.Committers {
		m[c] = true
	}

	if len(m) != len(req.Committers) {
		err = errors.New("duplicate committers")

		return
	}

	// platform
	if repo.Platform, err = dp.NewPackagePlatformByRepoLink(req.RepoLink); err != nil {
		return
	}
	platform := repo.Platform.PackagePlatform()

	// importer
	importerId := importer.Id(platform)
	if importerId == "" {
		err = errors.New("no platform Id")

		return
	}

	r := []domain.PkgCommitter{{Account: importer.Account, PlatformId: importerId}}

	// committers
	for _, c := range req.Committers {
		if c == importerId {
			continue
		}

		if u, err1 := ua.Find(c, platform); err1 != nil {
			err = err1
			invalidCommitter = append(invalidCommitter, c)
		} else {
			r = append(r, domain.PkgCommitter{Account: u.Account, PlatformId: c})
		}
	}

	repo.Committers = r

	return
}

type checkCommittersResp struct {
	InvalidCommitters []string `json:"invalid_committers"`
}
