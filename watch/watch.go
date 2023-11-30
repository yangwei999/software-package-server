package main

import (
	"net/http"
	"time"

	sdk "github.com/opensourceways/go-gitee/gitee"
	"github.com/opensourceways/robot-gitee-lib/client"
	"github.com/opensourceways/server-common-lib/utils"
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/software-package-server/softwarepkg/app"
	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/repository"
)

type iClient interface {
	GetGiteePullRequest(org, repo string, number int32) (sdk.PullRequest, error)
}

func NewWatchingImpl(
	cfg *Watch,
	initService app.SoftwarePkgInitAppService,
	watchService app.SoftwarePkgWatchService,
	r repository.Watch,
) *WatchingImpl {
	cli := client.NewClient(func() []byte {
		return []byte(cfg.RobotToken)
	})

	return &WatchingImpl{
		cfg:            cfg,
		cli:            cli,
		initAppService: initService,
		watchService:   watchService,
		watchRepo:      r,
		httpCli:        utils.NewHttpClient(3),
		stop:           make(chan struct{}),
		stopped:        make(chan struct{}),
	}
}

type WatchingImpl struct {
	cfg            *Watch
	cli            iClient
	initAppService app.SoftwarePkgInitAppService
	watchService   app.SoftwarePkgWatchService
	watchRepo      repository.Watch
	httpCli        utils.HttpClient
	stop           chan struct{}
	stopped        chan struct{}
}

func (impl *WatchingImpl) Start() {
	go impl.watch()
}

func (impl *WatchingImpl) Stop() {
	close(impl.stop)

	<-impl.stopped
}

func (impl *WatchingImpl) watch() {
	interval := impl.cfg.IntervalDuration()

	needStop := func() bool {
		select {
		case <-impl.stop:
			return true
		default:
			return false
		}
	}

	var timer *time.Timer

	defer func() {
		if timer != nil {
			timer.Stop()
		}

		close(impl.stopped)
	}()

	for {
		pkgIds, err := impl.initAppService.ListApprovedPkgs()
		if err != nil {
			logrus.Errorf("list approved pkgs failed, err: %s", err.Error())
		}

		impl.AddToWatch(pkgIds)

		watchPkgs, err := impl.watchRepo.FindAll()
		if err != nil {
			logrus.Errorf("find watch pkg failed, err: %s", err.Error())
		}

		for _, v := range watchPkgs {
			impl.handle(v)

			if needStop() {
				return
			}
		}

		// time starts.
		if timer == nil {
			timer = time.NewTimer(interval)
		} else {
			timer.Reset(interval)
		}

		select {
		case <-impl.stop:
			return

		case <-timer.C:
		}
	}
}

func (impl *WatchingImpl) handle(pw *domain.PkgWatch) {
	pkg, err := impl.initAppService.SoftwarePkg(pw.Id)
	if err != nil {
		logrus.Errorf("get pkg err: %s", err.Error())

		return
	}

	pw.Pkg = pkg

	switch pw.Status {
	case domain.PkgStatusInitialized:
		if err = impl.watchService.HandleCreatePR(pw); err != nil {
			logrus.Errorf("handle create pr err: %s", err.Error())

			return
		}

		url, _ := dp.NewURL(pw.PR.Link)
		if err = impl.initAppService.HandlePkgInitStarted(pw.Id, url); err != nil {
			logrus.Errorf("handle init started err: %s", err.Error())
		}
	case domain.PkgStatusPRCreated:
		if err = impl.handlePR(pw); err != nil {
			logrus.Errorf("handle pr err: %s", err.Error())
		}
	case domain.PkgStatusPRMerged:
		url, _ := dp.NewURL(pw.PR.Link)
		if err = impl.initAppService.HandlePkgInitDone(pw.Id, url); err != nil {
			logrus.Errorf("handle init done err: %s", err.Error())
		}

		if err = impl.watchService.HandleDone(pw); err != nil {
			logrus.Errorf("handle watch done err: %s", err.Error())
		}
	}
}

func (impl *WatchingImpl) handlePR(pw *domain.PkgWatch) error {
	pr, err := impl.cli.GetGiteePullRequest(impl.cfg.CommunityOrg,
		impl.cfg.CommunityRepo, int32(pw.PR.Num))
	if err != nil {
		return err
	}

	if pr.State == sdk.StatusOpen {
		return impl.handleCILabel(pw, pr)
	}

	return impl.handlePRState(pr, pw)
}

func (impl *WatchingImpl) handleCILabel(pw *domain.PkgWatch, pr sdk.PullRequest) error {
	cmd := app.CmdToHandleCI{
		PkgWatch: pw,
	}

	for _, l := range pr.Labels {
		switch l.Name {
		case impl.cfg.CISuccessLabel:
			cmd.IsSuccess = true
			return impl.watchService.HandleCI(&cmd)

		case impl.cfg.CIFailureLabel:
			if err := impl.watchService.HandleCI(&cmd); err != nil {
				logrus.Errorf("handle ci err: %s", err.Error())
			}

			url := pw.Pkg.Repo.Platform.RepoLink(pw.Pkg.Basic.Name)
			if !impl.isRepoExist(url) {
				return nil
			}

			return impl.initAppService.HandlePkgAlreadyExisted(pw.Id, url)
		}
	}

	return nil
}

func (impl *WatchingImpl) handlePRState(pr sdk.PullRequest, pw *domain.PkgWatch) error {
	switch pr.State {
	case sdk.StatusMerged:
		return impl.watchService.HandlePRMerged(pw)

	case sdk.StatusClosed:
		cmd := app.CmdToHandlePRClosed{
			PkgWatch:   pw,
			RejectedBy: "maintainer",
		}

		return impl.watchService.HandlePRClosed(&cmd)
	}

	return nil
}

func (impl *WatchingImpl) isRepoExist(url string) bool {
	request, _ := http.NewRequest(http.MethodHead, url, nil)

	code, _ := impl.httpCli.ForwardTo(request, nil)

	return code == 0
}

func (impl *WatchingImpl) AddToWatch(pkdId []string) {
	for _, id := range pkdId {
		pw := domain.PkgWatch{
			Id:     id,
			Status: domain.PkgStatusInitialized,
		}

		if err := impl.watchRepo.Add(&pw); err != nil {
			logrus.Errorf("add pkg id %s err: %s", id, err.Error())
		}
	}
}
