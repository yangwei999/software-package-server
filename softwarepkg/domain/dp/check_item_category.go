package dp

const (
	checkItemCategorySig          = "sig"
	checkItemCategoryCode         = "code"
	checkItemCategoryPkgName      = "pkg_name"
	checkItemCategoryUpstream     = "upstream"
	checkItemCategoryCommitter    = "committer"
	checkItemCategoryDescOrReason = "desc_or_reason"
)

var (
	CheckItemCategorySig          = checkItemCategory(checkItemCategorySig)
	CheckItemCategoryCode         = checkItemCategory(checkItemCategoryCode)
	CheckItemCategoryPkgName      = checkItemCategory(checkItemCategoryPkgName)
	CheckItemCategoryUpstream     = checkItemCategory(checkItemCategoryUpstream)
	CheckItemCategoryCommitter    = checkItemCategory(checkItemCategoryCommitter)
	CheckItemCategoryDescOrReason = checkItemCategory(checkItemCategoryDescOrReason)
)

type CheckItemCategory interface {
	CheckItemCategory() string
}

type checkItemCategory string

func (v checkItemCategory) CheckItemCategory() string {
	return string(v)
}