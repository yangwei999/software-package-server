package domain

import (
	"errors"
	"strings"

	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
)

const (
	pkgModificationSig       = "sig"
	pkgModificationCode      = "code"
	pkgModificationPkgName   = "pkg_name"
	pkgModificationPkgDesc   = "pkg_desc"
	pkgModificationPurpose   = "purpose"
	pkgModificationUpstream  = "upstream"
	pkgModificationCommitter = "committer"
)

var (
	ciConfig         CIConfig
	commonCheckItems []CheckItem
)

func Init(cfg *Config, m maintainer) {
	ciConfig = cfg.CIConfig
	commonCheckItems = cfg.checkItems

	maintainerInstance = m
}

func InitForMessageServer(cfg *CIConfig, ci pkgCI) {
	ciConfig = *cfg

	ciInstance = ci
}

func pkgModification(v string) (string, error) {
	v = strings.ToLower(v)

	switch v {
	case pkgModificationSig:
	case pkgModificationCode:
	case pkgModificationPkgName:
	case pkgModificationPkgDesc:
	case pkgModificationPurpose:
	case pkgModificationUpstream:
	case pkgModificationCommitter:
	default:
		return "", errors.New("invalid pkg modification")
	}

	return v, nil
}

type CIConfig struct {
	CITimeout     int64 `json:"ci_timeout"`
	CIWaitTimeout int64 `json:"ci_wait_timeout"`
}

func (cfg *CIConfig) SetDefault() {
	if cfg.CITimeout <= 0 {
		cfg.CITimeout = 3 * 3600 // 3 hours
	}

	if cfg.CIWaitTimeout <= 0 {
		cfg.CIWaitTimeout = 600 // 10m
	}
}

type Config struct {
	CIConfig

	CheckItems []checkItemConfig `json:"check_items" required:"true"`

	checkItems []CheckItem
}

func (cfg *Config) Validate() (err error) {
	v := make([]CheckItem, len(cfg.CheckItems))

	for i := range cfg.CheckItems {
		if v[i], err = cfg.CheckItems[i].toCheckItem(); err != nil {
			return
		}
	}

	cfg.checkItems = v

	return
}

type checkItemConfig struct {
	Id     string `json:"id"        required:"true"`
	Name   string `json:"name"      required:"true"`
	Desc   string `json:"desc"      required:"true"`
	NameEn string `json:"name_en"   required:"true"`
	DescEn string `json:"desc_en"   required:"true"`
	Owner  string `json:"owner"     required:"true"`

	// If true, keep the review record of reviewer, otherwise clear all the records about
	// this item when relevant modifications happened.
	// For example, the review about the item whether the user aggreed to
	// to be committer of the pkg should be kept.
	Keep bool `json:"keep"`

	// If true, only the owner can review this item else anyone can review.
	// For example, onlye sig maintainer can determine whether the sig of pkg is correct.
	OnlyOwner bool `json:"only_owner"`

	// This check item should be checked again when the relevant modifications happened.
	Modifications []string `json:"modifications" required:"true"`
}

func (cfg *checkItemConfig) toCheckItem() (item CheckItem, err error) {
	if item.Owner, err = dp.NewCommunityRole(cfg.Owner); err != nil {
		return
	}

	ms := make([]string, len(cfg.Modifications))
	for i, v := range cfg.Modifications {
		if ms[i], err = pkgModification(v); err != nil {
			return
		}
	}
	item.Modifications = ms

	item.Id = cfg.Id
	item.Name = cfg.Name
	item.Desc = cfg.Desc
	item.NameEn = cfg.NameEn
	item.DescEn = cfg.DescEn
	item.Keep = cfg.Keep
	item.OnlyOwner = cfg.OnlyOwner

	return
}
