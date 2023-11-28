package repositoryimpl

type Table struct {
	ReviewComment      string `json:"review_comment"        required:"true"`
	TranslationComment string `json:"translation_comment"   required:"true"`
	SoftwarePkgPR      string `json:"software_pkg_pr"       required:"true"`
}
