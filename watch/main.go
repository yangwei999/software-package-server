package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	kafka "github.com/opensourceways/kafka-lib/agent"
	mongdblib "github.com/opensourceways/mongodb-lib/mongodblib"
	"github.com/opensourceways/server-common-lib/logrusutil"
	liboptions "github.com/opensourceways/server-common-lib/options"
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/software-package-server/common/infrastructure/postgresql"
	"github.com/opensourceways/software-package-server/softwarepkg/app"
	"github.com/opensourceways/software-package-server/softwarepkg/infrastructure/pkgmanagerimpl"
	"github.com/opensourceways/software-package-server/softwarepkg/infrastructure/repositoryimpl"
	"github.com/opensourceways/software-package-server/softwarepkg/infrastructure/softwarepkgadapter"
	app2 "github.com/opensourceways/software-package-server/watch/app"
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

	defer kafka.Exit()

	run(cfg)
}

func run(cfg *Config) {
	pullRequestImpl, err := pullrequestimpl.NewPullRequestImpl(&cfg.PullRequest)
	if err != nil {
		logrus.Errorf("new pull request impl err:%s", err.Error())

		return
	}

	if err = pkgmanagerimpl.Init(&cfg.PkgManager); err != nil {
		logrus.Errorf("init pkg manager failed, err:%s", err.Error())

		return
	}

	initService := app.NewSoftwarePkgInitAppService(
		softwarepkgadapter.NewsoftwarePkgAdapter(
			mongdblib.DAO(cfg.Mongo.Collections.SoftwarePkg),
		),
		pkgmanagerimpl.Instance(),
		&producer{cfg.Topics.SoftwarePkgInitialized},
		repositoryimpl.NewSoftwarePkgComment(&cfg.Postgresql.Table),
	)

	watchService := app2.NewWatchService(
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
