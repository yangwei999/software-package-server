package app

import (
	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/repository"
	"github.com/opensourceways/software-package-server/utils"
)

type CmdToApplyNewSoftwarePkg struct {
	Sig      dp.ImportingPkgSig
	Spec     dp.URL
	SRPM     dp.URL
	Repo     domain.SoftwarePkgRepo
	Basic    domain.SoftwarePkgBasicInfo
	Importer domain.PkgCommitter
}

type CmdToUpdateSoftwarePkgApplication struct {
	PkgId    string
	Importer domain.PkgCommitter

	domain.SoftwarePkgUpdateInfo
}

type CmdToListPkgs = repository.OptToFindSoftwarePkgs

type CmdToWriteSoftwarePkgReviewComment struct {
	PkgId   string
	Author  dp.Account
	Content dp.ReviewComment
}

type CmdToClosePkg struct {
	PkgId    string
	Comment  dp.ReviewComment
	Reviewer domain.Reviewer
}

type NewSoftwarePkgDTO struct {
	Id string `json:"id"`
}

// SoftwarePkgSummaryDTO
type SoftwarePkgSummaryDTO struct {
	Id        string `json:"id"`
	Sig       string `json:"sig"`
	Phase     string `json:"phase"`
	PkgName   string `json:"pkg_name"`
	PkgDesc   string `json:"desc"`
	Platform  string `json:"platform"`
	RepoLink  string `json:"repo_link"`
	CIStatus  string `json:"ci_status"`
	Importer  string `json:"importer"`
	AppliedAt string `json:"applied_at"`
}

func toSoftwarePkgSummaryDTO(v *repository.SoftwarePkgInfo) SoftwarePkgSummaryDTO {
	return SoftwarePkgSummaryDTO{
		Id:        v.Id,
		Sig:       v.Sig.ImportingPkgSig(),
		Phase:     v.Phase.PackagePhase(),
		PkgName:   v.PkgName.PackageName(),
		PkgDesc:   v.PkgDesc.PackageDesc(),
		Platform:  v.Platform.PackagePlatform(),
		RepoLink:  v.Platform.RepoLink(v.PkgName),
		CIStatus:  v.CIStatus.PackageCIStatus(),
		Importer:  v.Importer.Account(),
		AppliedAt: utils.ToDate(v.AppliedAt),
	}
}

func toSoftwarePkgSummaryDTOs(v []repository.SoftwarePkgInfo) []SoftwarePkgSummaryDTO {
	r := make([]SoftwarePkgSummaryDTO, len(v))

	for i := range v {
		r[i] = toSoftwarePkgSummaryDTO(&v[i])
	}

	return r
}

// SoftwarePkgSummariesDTO
type SoftwarePkgSummariesDTO struct {
	Pkgs  []SoftwarePkgSummaryDTO `json:"pkgs"`
	Total int64                   `json:"total"`
}

// SoftwarePkgDTO
type SoftwarePkgDTO struct {
	SoftwarePkgSummaryDTO

	Spec       SoftwarePkgCodeFileDTO       `json:"spec"`
	SRPM       SoftwarePkgCodeFileDTO       `json:"srpm"`
	Purpose    string                       `json:"reason"`
	Upstream   string                       `json:"upstream"`
	Committers []string                     `json:"committers"`
	Logs       []SoftwarePkgOperationLogDTO `json:"logs"`
	Reviews    SoftwarePkgReviewDTO         `json:"reviews"`
}

func toSoftwarePkgDTO(pkg *domain.SoftwarePkg, dto *SoftwarePkgDTO, lang dp.Language) {
	basic := &pkg.Basic

	*dto = SoftwarePkgDTO{
		SoftwarePkgSummaryDTO: SoftwarePkgSummaryDTO{
			Id:        pkg.Id,
			Sig:       pkg.Sig.ImportingPkgSig(),
			Phase:     pkg.Phase.PackagePhase(),
			PkgName:   basic.Name.PackageName(),
			PkgDesc:   basic.Desc.PackageDesc(),
			Platform:  pkg.Repo.Platform.PackagePlatform(),
			RepoLink:  pkg.RepoLink(),
			CIStatus:  pkg.CI.Status().PackageCIStatus(),
			Importer:  pkg.Importer.Account.Account(),
			AppliedAt: utils.ToDate(pkg.AppliedAt),
		},
		Spec:       toSoftwarePkgCodeFileDTO(&pkg.Code.Spec),
		SRPM:       toSoftwarePkgCodeFileDTO(&pkg.Code.SRPM),
		Purpose:    pkg.Basic.Purpose.PurposeToImportPkg(),
		Upstream:   pkg.Basic.Upstream.URL(),
		Committers: pkg.Repo.CommitterIds(),
		Logs:       toSoftwarePkgOperationLogDTOs(pkg.Logs),
		Reviews:    toSoftwarePkgReviewDTO(pkg, lang),
	}
}

// SoftwarePkgCodeFileDTO
type SoftwarePkgCodeFileDTO struct {
	Src          string `json:"src"`
	Name         string `json:"name"`
	DownloadAddr string `json:"download_addr"`
}

func toSoftwarePkgCodeFileDTO(v *domain.SoftwarePkgCodeFile) SoftwarePkgCodeFileDTO {
	r := SoftwarePkgCodeFileDTO{
		Src:  v.Src.URL(),
		Name: v.FileName(),
	}

	if v.DownloadAddr != nil {
		r.DownloadAddr = v.DownloadAddr.URL()
	}

	return r
}

// SoftwarePkgOperationLogDTO
type SoftwarePkgOperationLogDTO struct {
	User   string `json:"user"`
	Time   string `json:"time"`
	Action string `json:"action"`
}

func toSoftwarePkgOperationLogDTO(v *domain.SoftwarePkgOperationLog) SoftwarePkgOperationLogDTO {
	return SoftwarePkgOperationLogDTO{
		User:   v.User.Account(),
		Time:   utils.ToDateTime(v.Time),
		Action: v.Action.PackageOperationLogAction(),
	}
}

func toSoftwarePkgOperationLogDTOs(v []domain.SoftwarePkgOperationLog) (r []SoftwarePkgOperationLogDTO) {
	if n := len(v); n > 0 {
		r = make([]SoftwarePkgOperationLogDTO, n)

		for i := range v {
			r[i] = toSoftwarePkgOperationLogDTO(&v[i])
		}
	}

	return
}

type SoftwarePkgReviewDTO struct {
	Items     []CheckItemReviewDTO `json:"items"`
	Reviewers []string             `json:"reviewers"`
}

func toSoftwarePkgReviewDTO(pkg *domain.SoftwarePkg, lang dp.Language) (dto SoftwarePkgReviewDTO) {
	if rs := pkg.Reviews; len(rs) > 0 {
		v := make([]string, len(rs))

		for i := range rs {
			v[i] = rs[i].Account.Account()
		}

		dto.Reviewers = v
	}

	reviews := pkg.CheckItemReviews()

	v := make([]CheckItemReviewDTO, len(reviews))
	for i := range reviews {
		v[i] = toCheckItemReviewDTO(&reviews[i], pkg, lang)
	}

	dto.Items = v

	return
}

// CheckItemReviewDTO
type CheckItemReviewDTO struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Owner string `json:"owner"`

	// HasResult if true, the pass is the result, otherwise the pass is meaningless.
	HasResult bool                     `json:"has_result"`
	Pass      bool                     `json:"pass"`
	Agree     int                      `json:"agree"`
	Disagree  int                      `json:"disagree"`
	Reviews   []UserCheckItemReviewDTO `json:"reviews"`
}

func toCheckItemReviewDTO(r *domain.CheckItemReview, pkg *domain.SoftwarePkg, lang dp.Language) CheckItemReviewDTO {
	item := r.Item

	dto := CheckItemReviewDTO{
		Id:    item.Id,
		Name:  item.GetName(lang),
		Desc:  item.GetDesc(lang),
		Owner: r.Item.OwnerDesc(pkg),
	}

	dto.HasResult, dto.Pass = r.Result()
	dto.Agree, dto.Disagree = r.Stat()

	if n := len(r.Reviews); n > 0 {
		v := make([]UserCheckItemReviewDTO, n)

		for i := range r.Reviews {
			v[i] = toUserCheckItemReviewDTO(&r.Reviews[i])
		}

		dto.Reviews = v
	}

	return dto
}

type UserCheckItemReviewDTO struct {
	// if the reviewer is the owner of item
	Owner   bool   `json:"owner"`
	Pass    bool   `json:"pass"`
	Account string `json:"account"`
	Comment string `json:"comment"`
}

func toUserCheckItemReviewDTO(r *domain.UserCheckItemReview) UserCheckItemReviewDTO {
	dto := UserCheckItemReviewDTO{
		Owner:   r.IsOwner,
		Pass:    r.Pass,
		Account: r.Account.Account(),
	}
	if r.Comment != nil {
		dto.Comment = r.Comment.ReviewComment()
	}

	return dto
}

// CheckItemUserReviewDTO
type CheckItemUserReviewDTO struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Owner string `json:"owner"`

	// Reviewed if true, the user reviewed this item before, otherwise the pass is meaningless.
	Reviewed  bool   `json:"reviewed"`
	CanReview bool   `json:"can_review"`
	Pass      bool   `json:"pass"`
	Comment   string `json:"comment"`
}

// SoftwarePkgReviewCommentDTO
type SoftwarePkgReviewCommentDTO struct {
	Id            string `json:"id"`
	Author        string `json:"author"`
	Content       string `json:"content"`
	CreatedAt     string `json:"created_at"`
	SinceCreation int64  `json:"since_creation"`
}

func toSoftwarePkgReviewCommentDTO(v *domain.SoftwarePkgReviewComment) SoftwarePkgReviewCommentDTO {
	return SoftwarePkgReviewCommentDTO{
		Id:            v.Id,
		Author:        v.Author.Account(),
		Content:       v.Content.ReviewComment(),
		CreatedAt:     utils.ToDateTime(v.CreatedAt),
		SinceCreation: utils.Now() - v.CreatedAt,
	}
}

func toSoftwarePkgReviewCommentDTOs(v []domain.SoftwarePkgReviewComment) (r []SoftwarePkgReviewCommentDTO) {
	if n := len(v); n > 0 {
		r = make([]SoftwarePkgReviewCommentDTO, n)
		for i := range v {
			r[i] = toSoftwarePkgReviewCommentDTO(&v[i])
		}
	}

	return
}

// CmdToTranslateReviewComment
type CmdToTranslateReviewComment = repository.TranslatedReviewCommentIndex

type TranslatedReveiwCommentDTO struct {
	Content string `json:"content"`
}
