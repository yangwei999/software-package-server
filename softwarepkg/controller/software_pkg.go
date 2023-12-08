package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	commonctl "github.com/opensourceways/software-package-server/common/controller"
	"github.com/opensourceways/software-package-server/common/controller/middleware"
	"github.com/opensourceways/software-package-server/softwarepkg/app"
	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/useradapter"
)

type SoftwarePkgController struct {
	service     app.SoftwarePkgService
	userAdapter useradapter.UserAdapter
}

func AddRouteForSoftwarePkgController(
	r *gin.RouterGroup,
	s app.SoftwarePkgService,
	u useradapter.UserAdapter,
) {
	ctl := SoftwarePkgController{
		service:     s,
		userAdapter: u,
	}

	m := middleware.UserChecking().CheckUser

	r.POST("/v1/softwarepkg/committers", m, ctl.CheckCommitters)

	r.POST("/v1/softwarepkg", m, ctl.ApplyNewPkg)
	r.GET("/v1/softwarepkg", ctl.ListPkgs)
	r.GET("/v1/softwarepkg/:id", ctl.Get)
	r.PUT("/v1/softwarepkg/:id", m, ctl.Update)
	r.PUT("/v1/softwarepkg/:id/retest", m, ctl.Retest)
	r.PUT("/v1/softwarepkg/:id/close", m, ctl.Close)

	r.POST("/v1/softwarepkg/:id/review", m, ctl.Review)
	r.GET("/v1/softwarepkg/:id/review", m, ctl.GetReview)
}

// CheckCommitter
// @Summary check committer of software package
// @Description check committer of software package
// @Tags  SoftwarePkg
// @Accept json
// @Param  body  body   softwarePkgRepoRequest   true  "body of checking committers"
// @Success 201 {object} checkCommittersResp
// @Failure 400 {object} ResponseData
// @Router /v1/softwarepkg/committers [post]
func (ctl SoftwarePkgController) CheckCommitters(ctx *gin.Context) {
	var req softwarePkgRepoRequest
	if err := ctx.ShouldBindWith(&req, binding.JSON); err != nil {
		commonctl.SendBadRequestBody(ctx, err)

		return
	}

	user, err := middleware.UserChecking().FetchUser(ctx)
	if err != nil {
		commonctl.SendError(ctx, err)

		return
	}

	invalidCommitters, err := req.check(&user, ctl.userAdapter)
	if err != nil {
		if len(invalidCommitters) > 0 {
			commonctl.SendRespOfPost(ctx, checkCommittersResp{invalidCommitters})
		} else {
			commonctl.SendBadRequestParam(ctx, err)
		}
	} else {
		commonctl.SendRespOfPost(ctx, checkCommittersResp{})
	}
}

// ApplyNewPkg
// @Summary apply a new software package
// @Description apply a new software package
// @Tags  SoftwarePkg
// @Accept json
// @Param	param  body	 softwarePkgRequest	 true	"body of applying a new software package"
// @Success 201 {object} app.NewSoftwarePkgDTO
// @Failure 400 {object} ResponseData
// @Router /v1/softwarepkg [post]
func (ctl SoftwarePkgController) ApplyNewPkg(ctx *gin.Context) {
	var req softwarePkgRequest
	if err := ctx.ShouldBindWith(&req, binding.JSON); err != nil {
		commonctl.SendBadRequestBody(ctx, err)

		return
	}

	user, err := middleware.UserChecking().FetchUser(ctx)
	if err != nil {
		commonctl.SendError(ctx, err)

		return
	}

	cmd, err := req.toCmd(&user, ctl.userAdapter)
	if err != nil {
		commonctl.SendBadRequestParam(ctx, err)

		return
	}

	if r, err := ctl.service.ApplyNewPkg(&cmd); err != nil {
		commonctl.SendError(ctx, err)
	} else {
		commonctl.SendRespOfPost(ctx, r)
	}
}

// ListPkgs
// @Summary list software packages
// @Description list software packages
// @Tags  SoftwarePkg
// @Accept json
// @param    phase            query	 string   false    "phase of the softwarepkg"
// @param    pkg_name         query	 string   false    "name of the softwarepkg"
// @Param    importer         query	 string   false    "importer of the softwarePkg"
// @Param    platform         query	 string   false    "platform of the softwarePkg"
// @Param    last_id          query	 string   false    "last software pkg id of previous page"
// @Param    count            query	 bool     false    "whether count total num of the pkgs"
// @Param    page_num         query	 int      false    "page num which starts from 1"
// @Param    count_per_page   query	 int      false    "count per page"
// @Success 200 {object} app.SoftwarePkgSummariesDTO
// @Failure 400 {object} ResponseData
// @Router /v1/softwarepkg [get]
func (ctl SoftwarePkgController) ListPkgs(ctx *gin.Context) {
	var req softwarePkgListQuery
	if err := ctx.ShouldBindQuery(&req); err != nil {
		commonctl.SendBadRequestParam(ctx, err)

		return
	}

	cmd, err := req.toCmd()
	if err != nil {
		commonctl.SendBadRequestParam(ctx, err)

		return
	}

	if v, err := ctl.service.ListPkgs(&cmd); err != nil {
		commonctl.SendError(ctx, err)
	} else {
		commonctl.SendRespOfGet(ctx, v)
	}
}

// Get
// @Summary get software package
// @Description get software package
// @Tags  SoftwarePkg
// @Accept json
// @Param    id         path	string  true    "id of software package"
// @param    language   query	string  false   "language"
// @Success 200 {object} app.SoftwarePkgDTO
// @Failure 400 {object} ResponseData
// @Router /v1/softwarepkg/{id} [get]
func (ctl SoftwarePkgController) Get(ctx *gin.Context) {
	var req softwarePkgGetQuery
	if err := ctx.ShouldBindQuery(&req); err != nil {
		commonctl.SendBadRequestParam(ctx, err)

		return
	}

	lang, err := req.language()
	if err != nil {
		commonctl.SendBadRequestParam(ctx, err)

		return
	}

	if v, err := ctl.service.Get(ctx.Param("id"), lang); err != nil {
		commonctl.SendError(ctx, err)
	} else {
		commonctl.SendRespOfGet(ctx, v)
	}
}

// GetReview
// @Summary get user review on software package
// @Description get user review on software package
// @Tags  SoftwarePkg
// @Accept json
// @Param    id         path	string  true    "id of software package"
// @param    language   query	string  false   "language"
// @Success 200 {object} app.CheckItemUserReviewDTO
// @Failure 400 {object} ResponseData
// @Router /v1/softwarepkg/{id}/review [get]
func (ctl SoftwarePkgController) GetReview(ctx *gin.Context) {
	var req softwarePkgGetQuery
	if err := ctx.ShouldBindQuery(&req); err != nil {
		commonctl.SendBadRequestParam(ctx, err)

		return
	}

	lang, err := req.language()
	if err != nil {
		commonctl.SendBadRequestParam(ctx, err)

		return
	}

	user, err := middleware.UserChecking().FetchUser(ctx)
	if err != nil {
		commonctl.SendError(ctx, err)

		return
	}

	if v, err := ctl.service.GetReview(ctx.Param("id"), &user, lang); err != nil {
		commonctl.SendError(ctx, err)
	} else {
		commonctl.SendRespOfGet(ctx, v)
	}
}

// Review
// @Summary review software package
// @Description review software package
// @Tags  SoftwarePkg
// @Accept json
// @Param  id     path   string         true  "id of software package"
// @Param  body   body   reviewRequest  true  "body of reviewing a software package"
// @Success 201 {object} ResponseData
// @Failure 400 {object} ResponseData
// @Router /v1/softwarepkg/{id}/review [post]
func (ctl SoftwarePkgController) Review(ctx *gin.Context) {
	var req reviewRequest

	if err := ctx.ShouldBindWith(&req, binding.JSON); err != nil {
		commonctl.SendBadRequestBody(ctx, err)

		return
	}

	info, err := req.toCmd()
	if err != nil {
		commonctl.SendBadRequestParam(ctx, err)

		return
	}

	user, err := middleware.UserChecking().FetchUser(ctx)
	if err != nil {
		commonctl.SendError(ctx, err)

		return
	}

	reviewer := domain.Reviewer{
		Account: user.Account,
		GiteeID: user.GiteeID,
	}

	if err := ctl.service.Review(ctx.Param("id"), &reviewer, info); err != nil {
		commonctl.SendError(ctx, err)
	} else {
		commonctl.SendRespOfPut(ctx)
	}
}

// Close
// @Summary close software package
// @Description close software package
// @Tags  SoftwarePkg
// @Accept json
// @Param  id    path   string          true  "id of software package"
// @Param  body  body   reqToClosePkg   true  "comment"
// @Success 202 {object} ResponseData
// @Failure 400 {object} ResponseData
// @Router /v1/softwarepkg/{id}/close [put]
func (ctl SoftwarePkgController) Close(ctx *gin.Context) {
	user, err := middleware.UserChecking().FetchUser(ctx)
	if err != nil {
		commonctl.SendError(ctx, err)

		return
	}

	var req reqToClosePkg
	if err = ctx.ShouldBindWith(&req, binding.JSON); err != nil {
		commonctl.SendBadRequestBody(ctx, err)

		return
	}

	cmd, err := req.toCmd(ctx.Param("id"), &user)
	if err != nil {
		commonctl.SendBadRequestParam(ctx, err)

		return
	}

	if err := ctl.service.Close(&cmd); err != nil {
		commonctl.SendError(ctx, err)
	} else {
		commonctl.SendRespOfPut(ctx)
	}
}

// Update
// @Summary update application of software package
// @Description update application of software package
// @Tags  SoftwarePkg
// @Accept json
// @Param  id     path  string                  true  "id of software package"
// @Param  param  body  reqToUpdateSoftwarePkg  true  "body of updating software package application"
// @Success 202 {object} ResponseData
// @Failure 400 {object} ResponseData
// @Router /v1/softwarepkg/:id [put]
func (ctl SoftwarePkgController) Update(ctx *gin.Context) {
	var req reqToUpdateSoftwarePkg

	if err := ctx.ShouldBindWith(&req, binding.JSON); err != nil {
		commonctl.SendBadRequestBody(ctx, err)

		return
	}

	user, err := middleware.UserChecking().FetchUser(ctx)
	if err != nil {
		commonctl.SendError(ctx, err)

		return
	}

	cmd, err := req.toCmd(ctx.Param("id"), &user, ctl.userAdapter)
	if err != nil {
		commonctl.SendBadRequestParam(ctx, err)

		return
	}

	if err = ctl.service.Update(&cmd); err != nil {
		commonctl.SendError(ctx, err)
	} else {
		commonctl.SendRespOfPut(ctx)
	}
}

// ReRunCI
// @Summary rerun ci of software package
// @Description rerun ci of software package
// @Tags  SoftwarePkg
// @Accept json
// @Param	id  path	 string	 true	"id of software package"
// @Success 202 {object} ResponseData
// @Failure 400 {object} ResponseData
// @Router /v1/softwarepkg/{id}/retest [put]
func (ctl SoftwarePkgController) Retest(ctx *gin.Context) {
	user, err := middleware.UserChecking().FetchUser(ctx)
	if err != nil {
		commonctl.SendError(ctx, err)

		return
	}

	if err := ctl.service.Retest(ctx.Param("id"), &user); err != nil {
		commonctl.SendError(ctx, err)
	} else {
		commonctl.SendRespOfPut(ctx)
	}
}
