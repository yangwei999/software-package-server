package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	kafka "github.com/opensourceways/kafka-lib/agent"
	"github.com/opensourceways/server-common-lib/logrusutil"
	liboptions "github.com/opensourceways/server-common-lib/options"
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/software-package-server/common/infrastructure/postgresql"
	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
	"github.com/opensourceways/software-package-server/softwarepkg/infrastructure/pkgmanagerimpl"
	"github.com/opensourceways/software-package-server/softwarepkg/infrastructure/useradapterimpl"
	"github.com/opensourceways/software-package-server/watch/app"
	"github.com/opensourceways/software-package-server/watch/infrastructure/emailimpl"
	"github.com/opensourceways/software-package-server/watch/infrastructure/pullrequestimpl"
	wathcrepoimpl "github.com/opensourceways/software-package-server/watch/infrastructure/repositoryimpl"
)

type options struct {
	service liboptions.ServiceOptions
}

func (o *options) Validate() error {
	return o.service.Validate()
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options

	o.service.AddFlags(fs)

	fs.Parse(args)

	return o
}

func main() {
	logrusutil.ComponentInit("software-package-watch")
	log := logrus.NewEntry(logrus.StandardLogger())

	o := gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	if err := o.Validate(); err != nil {
		logrus.Errorf("Invalid options, err:%s", err.Error())

		return
	}

	cfg, err := loadConfig(o.service.ConfigFile)
	if err != nil {
		logrus.Errorf("load config failed, err:%s", err.Error())

		return
	}

	// postgresql
	if err = postgresql.Init(&cfg.Postgresql.DB); err != nil {
		logrus.Errorf("init db failed, err:%s", err.Error())

		return
	}

	// kafka
	if err = kafka.Init(&cfg.Kafka, log, nil, "", false); err != nil {
		logrus.Errorf("init kafka failed, err:%s", err.Error())

		return
	}

	if err := useradapterimpl.Init(&cfg.User); err != nil {
		logrus.Errorf("init maintainer failed, err:%s", err.Error())

		return
	}

	defer kafka.Exit()

	run(cfg)
}

type initServiceTest struct {
}

func (s initServiceTest) ListApprovedPkgs() ([]string, error) {
	return []string{"d0e361ee-dc00-4d71-b756-32f2dc276583"}, nil
}

func (s initServiceTest) SoftwarePkg(pkgId string) (domain.SoftwarePkg, error) {
	sig, _ := dp.NewImportingPkgSig("sig-ops")
	platform, _ := dp.NewPackagePlatform("gitee")
	account, _ := dp.NewAccount("georgecao")
	name, _ := dp.NewPackageName("caca")
	desc, _ := dp.NewPackageDesc("ok: i am desc")
	prupose, _ := dp.NewPurposeToImportPkg("i am purpose")
	upstream, _ := dp.NewURL("https://baidu.com")
	comment, _ := dp.NewReviewComment("i am comment")

	commiters := []domain.PkgCommitter{
		{
			Account:    account,
			PlatformId: "gitee",
		},
		{
			Account:    account,
			PlatformId: "gitee",
		},
	}
	return domain.SoftwarePkg{
		Id:  "d0e361ee-dc00-4d71-b756-32f2dc276583",
		Sig: sig,
		Repo: domain.SoftwarePkgRepo{
			Platform:   platform,
			Committers: commiters,
		},
		Basic: domain.SoftwarePkgBasicInfo{
			Name:     name,
			Desc:     desc,
			Purpose:  prupose,
			Upstream: upstream,
		},
		Importer: account,
		Reviews: []domain.UserReview{
			{
				Reviewer: domain.Reviewer{
					Account: account,
				},
				Reviews: []domain.CheckItemReviewInfo{
					{
						Id:      "2",
						Pass:    true,
						Comment: comment,
					},
					{
						Id:   "6",
						Pass: true,
					},
					{
						Id:   "4",
						Pass: false,
					},
				},
			},
			{
				Reviewer: domain.Reviewer{
					Account: account,
				},
				Reviews: []domain.CheckItemReviewInfo{
					{
						Id:      "7",
						Pass:    true,
						Comment: comment,
					},
					{
						Id:      "1",
						Pass:    true,
						Comment: comment,
					},
					{
						Id:   "5",
						Pass: false,
					},
				},
			},
		},
	}, nil
}

func (s initServiceTest) HandlePkgInitDone(pkgId string, pr dp.URL) error {
	return nil
}

func (s initServiceTest) HandlePkgInitStarted(pkgId string, pr dp.URL) error {
	return nil
}

func (s initServiceTest) HandlePkgAlreadyExisted(pkgId string, repoLink string) error {
	return nil
}

func (s initServiceTest) Send(subject, content string) error {
	return nil
}

func run(cfg *Config) {
	pullRequestImpl, err := pullrequestimpl.NewPullRequestImpl(&cfg.PullRequest, useradapterimpl.UserAdapter())
	if err != nil {
		logrus.Errorf("new pull request impl err:%s", err.Error())

		return
	}

	if err = pkgmanagerimpl.Init(&cfg.PkgManager); err != nil {
		logrus.Errorf("init pkg manager failed, err:%s", err.Error())

		return
	}

	initService := new(initServiceTest)
	watchService := app.NewWatchService(
		pullRequestImpl,
		wathcrepoimpl.NewSoftwarePkgPR(&cfg.Postgresql.WatchTable),
		emailimpl.NewEmailService(cfg.Email),
	)

	// watch
	w := NewWatchingImpl(&cfg.Watch, initService, watchService)
	w.Start()
	defer w.Stop()

	// wait
	wait()
}

func wait() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	defer wg.Wait()

	called := false
	ctx, done := context.WithCancel(context.Background())

	defer func() {
		if !called {
			called = true
			done()
		}
	}()

	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()

		select {
		case <-ctx.Done():
			logrus.Info("receive done. exit normally")
			return

		case <-sig:
			logrus.Info("receive exit signal")
			called = true
			done()
			return
		}
	}(ctx)

	<-ctx.Done()
}
