package main

import (
	"time"

	kafka "github.com/opensourceways/kafka-lib/agent"
	"github.com/opensourceways/server-common-lib/utils"

	"github.com/opensourceways/software-package-server/common/infrastructure/postgresql"
	"github.com/opensourceways/software-package-server/softwarepkg/infrastructure/pullrequestimpl"
	"github.com/opensourceways/software-package-server/softwarepkg/infrastructure/repositoryimpl"
)

type configValidate interface {
	Validate() error
}

type configSetDefault interface {
	SetDefault()
}

type PostgresqlConfig struct {
	DB    postgresql.Config    `json:"db"    required:"true"`
	Table repositoryimpl.Table `json:"table" required:"true"`
}

type Watch struct {
	RobotToken     string `json:"robot_token"      required:"true"`
	PkgOrg         string `json:"pkg_org"          required:"true"`
	CommunityOrg   string `json:"community_org"    required:"true"`
	CommunityRepo  string `json:"community_repo"   required:"true"`
	CISuccessLabel string `json:"ci_success_label" required:"true"`
	CIFailureLabel string `json:"ci_failure_label" required:"true"`
	// unit second
	Interval int `json:"interval"`
}

type Config struct {
	Kafka       kafka.Config           `json:"kafka"`
	Postgresql  PostgresqlConfig       `json:"postgresql"`
	Watch       Watch                  `json:"watch"`
	PullRequest pullrequestimpl.Config `json:"pull_request"`
}

func loadConfig(path string) (*Config, error) {
	cfg := new(Config)
	if err := utils.LoadFromYaml(path, cfg); err != nil {
		return nil, err
	}

	cfg.SetDefault()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) configItems() []interface{} {
	return []interface{}{
		&cfg.Kafka,
		&cfg.Postgresql.DB,
		&cfg.Postgresql.Table,
		&cfg.Watch,
		&cfg.PullRequest,
	}
}

func (cfg *Config) SetDefault() {
	items := cfg.configItems()
	for _, i := range items {
		if f, ok := i.(configSetDefault); ok {
			f.SetDefault()
		}
	}
}

func (w *Watch) SetDefault() {
	if w.PkgOrg == "" {
		w.PkgOrg = "src-openeuler"
	}

	if w.CommunityOrg == "" {
		w.CommunityOrg = "openeuler"
	}

	if w.CommunityRepo == "" {
		w.CommunityRepo = "community"
	}

	if w.CISuccessLabel == "" {
		w.CISuccessLabel = "ci_successful"
	}

	if w.CIFailureLabel == "" {
		w.CIFailureLabel = "ci_failed"
	}

	if w.Interval <= 0 {
		w.Interval = 10
	}
}

func (cfg *Config) Validate() error {
	if _, err := utils.BuildRequestBody(cfg, ""); err != nil {
		return err
	}

	items := cfg.configItems()
	for _, i := range items {
		if f, ok := i.(configValidate); ok {
			if err := f.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *Watch) IntervalDuration() time.Duration {
	return time.Second * time.Duration(w.Interval)
}
