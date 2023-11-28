package main

import (
	"time"

	kafka "github.com/opensourceways/kafka-lib/agent"
	"github.com/opensourceways/server-common-lib/utils"

	"github.com/opensourceways/software-package-server/common/infrastructure/postgresql"
	"github.com/opensourceways/software-package-server/softwarepkg/infrastructure/repositoryimpl"
)

type configValidate interface {
	Validate() error
}

type configSetDefault interface {
	SetDefault()
}

type PostgresqlConfig struct {
	DB postgresql.Config `json:"db" required:"true"`
	repositoryimpl.Table
}

type Config struct {
	Kafka          kafka.Config     `json:"kafka"`
	Postgresql     PostgresqlConfig `json:"postgresql"`
	RobotToken     string           `json:"robot_token"      required:"true"`
	PkgOrg         string           `json:"pkg_org"          required:"true"`
	CommunityOrg   string           `json:"community_org"    required:"true"`
	CommunityRepo  string           `json:"community_repo"   required:"true"`
	CISuccessLabel string           `json:"ci_success_label" required:"true"`
	CIFailureLabel string           `json:"ci_failure_label" required:"true"`
	// unit second
	Interval int `json:"interval"`
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
	}
}

func (cfg *Config) SetDefault() {
	if cfg.PkgOrg == "" {
		cfg.PkgOrg = "src-openeuler"
	}

	if cfg.Interval <= 0 {
		cfg.Interval = 10
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

func (cfg *Config) IntervalDuration() time.Duration {
	return time.Second * time.Duration(cfg.Interval)
}
