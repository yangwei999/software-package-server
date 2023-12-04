package pullrequestimpl

import (
	"fmt"
	"strings"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
)

type Config struct {
	Robot          robotConfig          `json:"robot"`
	Template       templateConfig       `json:"template"`
	SoftwarePkg    softwarePkg          `json:"software_pkg"`
	ShellScript    shellConfig          `json:"shell_script"`
	CommunityRobot communityRobotConfig `json:"community_robot"`
	*domain.Config
}

func (cfg *Config) SetDefault() {
	cfg.Robot.setDefault()
	cfg.CommunityRobot.setDefault()
	cfg.Template.setDefault()
	cfg.ShellScript.setdefault()
}

type shellConfig struct {
	WorkDir      string `json:"work_dir"`
	BranchScript string `json:"branch_script"`
	CloneScript  string `json:"clone_script"`
}

func (cfg *shellConfig) setdefault() {
	if cfg.WorkDir == "" {
		cfg.WorkDir = "/opt/app/work_dir"
	}

	if cfg.BranchScript == "" {
		cfg.BranchScript = "/opt/app/create_branch.sh"
	}

	if cfg.CloneScript == "" {
		cfg.CloneScript = "/opt/app/clone_repo.sh"
	}
}

type robotConfig struct {
	Username      string        `json:"username"           required:"true"`
	Token         string        `json:"token"              required:"true"`
	Email         string        `json:"email"              required:"true"`
	Repo          string        `json:"repo"               required:"true"`
	RepoLink      string        `json:"link"               required:"true"`
	NewRepoBranch newRepoBranch `json:"new_repo_branch"`
}

func (cfg *robotConfig) setDefault() {
	cfg.NewRepoBranch.setDefault()
}

func (cfg *robotConfig) cloneURL() string {
	return fmt.Sprintf(
		"https://%s:%s@%s",
		cfg.Username, cfg.Token,
		strings.TrimPrefix(cfg.RepoLink, "https://"),
	)
}

type communityRobotConfig struct {
	Token    string `json:"token" required:"true"`
	Org      string `json:"org"`
	Repo     string `json:"repo"`
	RepoLink string `json:"link"  required:"true"`
}

func (cfg *communityRobotConfig) setDefault() {
	if cfg.Org == "" {
		cfg.Org = "openeuler"
	}

	if cfg.Repo == "" {
		cfg.Repo = "community"
	}
}

type newRepoBranch struct {
	Name        string `json:"name"`
	ProtectType string `json:"protect_type"`
	PublicType  string `json:"public_type"`
}

func (cfg *newRepoBranch) setDefault() {
	if cfg.Name == "" {
		cfg.Name = "master"
	}

	if cfg.ProtectType == "" {
		cfg.ProtectType = "protected"
	}

	if cfg.PublicType == "" {
		cfg.PublicType = "public"
	}
}

type templateConfig struct {
	PRBodyTpl       string `json:"pr_body_tpl"`
	SigInfoTpl      string `json:"sig_info_tpl"`
	RepoYamlTpl     string `json:"repo_yaml_tpl"`
	CheckItemsTpl   string `json:"check_items_tpl"`
	ReviewDetailTpl string `json:"review_detail_tpl"`
}

func (t *templateConfig) setDefault() {
	if t.PRBodyTpl == "" {
		t.PRBodyTpl = "/opt/app/template/pr_body.tpl"
	}

	if t.SigInfoTpl == "" {
		t.SigInfoTpl = "/opt/app/template/sig_info.tpl"
	}

	if t.RepoYamlTpl == "" {
		t.RepoYamlTpl = "/opt/app/template/repo_yaml.tpl"
	}

	if t.CheckItemsTpl == "" {
		t.CheckItemsTpl = "/opt/app/template/check_items.tpl"
	}

	if t.ReviewDetailTpl == "" {
		t.ReviewDetailTpl = "/opt/app/template/review_detail.tpl"
	}
}

type softwarePkg struct {
	Endpoint string `json:"endpoint" required:"true"`
}
