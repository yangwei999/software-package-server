package controller

import (
	"errors"

	"github.com/opensourceways/software-package-server/allerror"
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

func (req *softwarePkgRequest) toCmd(user *domain.User, ua useradapter.UserAdapter) (
	cmd app.CmdToApplyNewSoftwarePkg, err error,
) {
	if cmd.Basic, err = req.toBasic(); err != nil {
		return
	}

	cmd.Spec, err = parseSpec(req.Spec)
	if err != nil {
		return
	}

	cmd.SRPM, err = parseSRPM(req.SRPM)
	if err != nil {
		return
	}

	cmd.Sig, err = dp.NewImportingPkgSig(req.Sig)
	if err != nil {
		return
	}

	cmd.Repo, cmd.Importer, err = req.toRepo(user, ua)

	return
}

func parseSpec(url string) (dp.URL, error) {
	spec, err := dp.NewURL(url)
	if err != nil {
		return nil, err
	}

	if !dp.IsSpec(spec.FileName()) {
		err = allerror.New(allerror.ErrorCodeParamNotSpec, "not spec file")
	}

	return spec, err
}

func parseSRPM(url string) (dp.URL, error) {
	srpm, err := dp.NewURL(url)
	if err != nil {
		return nil, err
	}

	if !dp.IsSRPM(srpm.FileName()) {
		err = allerror.New(allerror.ErrorCodeParamNotSRPM, "not srpm file")
	}

	return srpm, err
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
	pkgId string, user *domain.User, ua useradapter.UserAdapter,
) (
	cmd app.CmdToUpdateSoftwarePkgApplication, err error,
) {
	cmd.PkgId = pkgId

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
		if cmd.Spec, err = parseSpec(req.Spec); err != nil {
			return
		}
	}

	if req.SRPM != "" {
		if cmd.SRPM, err = parseSRPM(req.SRPM); err != nil {
			return
		}
	}

	if req.Sig != "" {
		if cmd.Sig, err = dp.NewImportingPkgSig(req.Sig); err != nil {
			return
		}
	}

	// 1. It must pass repo link if updating committers or repo link.
	// 2. It will override the committers,
	// so it should pass committers event if it didn't not change.
	if req.RepoLink == "" {
		return
	}

	v := softwarePkgRepoRequest{
		RepoLink:   req.RepoLink,
		Committers: req.Committers,
	}
	cmd.Repo, cmd.Importer, err = v.toRepo(user, ua)

	return
}

// softwarePkgGetQuery
type softwarePkgGetQuery struct {
	Language string `json:"language"           form:"language"`
}

func (req *softwarePkgGetQuery) language() (dp.Language, error) {
	if req.Language == "" {
		return dp.NewLanguage(dp.Chinese)
	}

	return dp.NewLanguage(req.Language)
}

// softwarePkgListQuery
type softwarePkgListQuery struct {
	Phase        string `json:"phase"          form:"phase"`
	PkgName      string `json:"pkg_name"       form:"pkg_name"`
	Importer     string `json:"importer"       form:"importer"`
	Platform     string `json:"platform"       form:"platform"`
	LastId       string `json:"last_id"        form:"last_id"`
	Count        bool   `json:"count"          form:"count"`
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

// reqToClosePkg
type reqToClosePkg struct {
	Comment string `json:"comment"`
}

func (req *reqToClosePkg) toCmd(pkgId string, user *domain.User) (cmd app.CmdToClosePkg, err error) {
	if req.Comment != "" {
		if cmd.Comment, err = dp.NewReviewComment(req.Comment); err != nil {
			return
		}
	}

	cmd.PkgId = pkgId
	cmd.Reviewer = domain.Reviewer{
		Account: user.Account,
		GiteeID: user.GiteeID,
	}

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
		err = allerror.New(allerror.ErrorCodeParamMissingChekItemComment, "lack of comment")

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
	Committers []string `json:"committers"`
}

func (req *softwarePkgRepoRequest) toRepo(user *domain.User, ua useradapter.UserAdapter) (
	repo domain.SoftwarePkgRepo, importer domain.PkgCommitter, err error,
) {
	if len(req.Committers) > config.MaxNumOfCommitters {
		err = allerror.New(allerror.ErrorCodeParamTooManyCommitters, "too many committers")

		return
	}

	m := map[string]bool{}
	for _, c := range req.Committers {
		m[c] = true
	}

	if len(m) != len(req.Committers) {
		err = allerror.New(allerror.ErrorCodeParamDuplicateCommitters, "duplicate committers")

		return
	}

	// platform
	if repo.Platform, err = dp.NewPackagePlatformByRepoLink(req.RepoLink); err != nil {
		return
	}
	platform := repo.Platform.PackagePlatform()

	// importer
	importerId := user.Id(platform)
	if importerId == "" {
		err = allerror.New(allerror.ErrorCodeParamImporterMissingPlatformId, "no platform Id")

		return
	}

	importer = domain.PkgCommitter{
		Account:    user.Account,
		PlatformId: importerId,
	}

	// committers
	var r []domain.PkgCommitter

	for _, c := range req.Committers {
		if c == importerId {
			continue
		}

		if u, err1 := ua.Find(c, platform); err1 != nil {
			err = err1
		} else {
			r = append(r, domain.PkgCommitter{Account: u.Account, PlatformId: c})
		}
	}

	repo.Committers = r

	return
}

func (req *softwarePkgRepoRequest) check(user *domain.User, ua useradapter.UserAdapter) ([]string, error) {
	if len(req.Committers) > config.MaxNumOfCommitters {
		return nil, allerror.New(allerror.ErrorCodeParamTooManyCommitters, "too many committers")
	}

	// platform
	platform, err := dp.NewPackagePlatformByRepoLink(req.RepoLink)
	if err != nil {
		return nil, err
	}
	platformStr := platform.PackagePlatform()

	// committers
	var r []string

	for _, c := range req.Committers {
		if _, err1 := ua.Find(c, platformStr); err1 != nil {
			err = err1
			r = append(r, c)
		}
	}

	return r, err
}

type checkCommittersResp struct {
	InvalidCommitters []string `json:"invalid_committers"`
}
