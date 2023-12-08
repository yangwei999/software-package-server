package dp

import "strings"

const (
	Chinese = "chinese"
	English = "english"
)

var config Config

func Init(cfg *Config, sv SigValidator, words sensitiveWordsValidator) {
	config = *cfg
	sigValidator = sv
	sensitiveWords = words
}

func InitForMessageServer(cfg *Config) {
	config = *cfg
}

type Config struct {
	// map platform -> org address. such as gitee --> https://gitee.com/src-openeuler/
	PlatformOrgLinks              map[string]string `json:"platform_org_links"            required:"true"`
	SupportedLanguages            []string          `json:"supported_languages"`
	MaxLengthOfPackageName        int               `json:"max_length_of_pkg_name"`
	MaxLengthOfPackageDesc        int               `json:"max_length_of_pkg_desc"`
	MaxLengthOfReviewComment      int               `json:"max_length_of_review_comment"`
	MaxLengthOfCheckItemComment   int               `json:"max_length_of_check_item_comment"`
	MaxLengthOfPurposeToImportPkg int               `json:"max_length_of_purpose_to_import_pkg"`
}

func (cfg *Config) SetDefault() {
	if len(cfg.SupportedLanguages) == 0 {
		cfg.SupportedLanguages = []string{Chinese, English}
	}

	if cfg.MaxLengthOfPackageName <= 0 {
		cfg.MaxLengthOfPackageName = 60
	}

	if cfg.MaxLengthOfPackageDesc <= 0 {
		cfg.MaxLengthOfPackageDesc = 1000
	}

	if cfg.MaxLengthOfReviewComment <= 0 {
		cfg.MaxLengthOfReviewComment = 1500
	}

	if cfg.MaxLengthOfCheckItemComment <= 0 {
		cfg.MaxLengthOfCheckItemComment = 500
	}

	if cfg.MaxLengthOfPurposeToImportPkg <= 0 {
		cfg.MaxLengthOfPurposeToImportPkg = 1000
	}
}

func (cfg *Config) Validate() error {
	cfg.toLower(cfg.SupportedLanguages)

	m := map[string]string{}
	for p, addr := range cfg.PlatformOrgLinks {
		if !strings.HasSuffix(addr, "/") {
			addr += "/"
		}

		m[strings.ToLower(p)] = addr
	}

	cfg.PlatformOrgLinks = m

	return nil
}

func (cfg *Config) toLower(items []string) {
	for i, s := range items {
		items[i] = strings.ToLower(s)
	}
}

func (cfg *Config) isValidLanguage(v string) bool {
	return cfg.has(strings.ToLower(v), cfg.SupportedLanguages)
}

func (cfg *Config) isValidPlatform(v string) bool {
	return cfg.PlatformOrgLinks[v] != ""
}

func (cfg *Config) platformOfRepoLink(repoLink string) string {
	for p, orgLink := range cfg.PlatformOrgLinks {
		if strings.HasPrefix(repoLink, orgLink) {
			return p
		}
	}

	return ""
}

func (cfg *Config) orgLinkOfPlatform(v string) string {
	return cfg.PlatformOrgLinks[v]
}

func (cfg *Config) has(v string, items []string) bool {
	for _, s := range items {
		if v == s {
			return true
		}
	}

	return false
}
