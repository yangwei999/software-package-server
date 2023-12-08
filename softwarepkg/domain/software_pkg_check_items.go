package domain

import (
	"fmt"
	"strings"

	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
)

func (entity *SoftwarePkg) CheckItemsMap() map[string]string {
	m := map[string]string{}

	for i := range commonCheckItems {
		item := &commonCheckItems[i]

		m[item.Id] = item.Name
	}

	others := entity.otherCheckItems()
	for i := range others {
		item := &others[i]

		m[item.Id] = item.Name
	}

	return m
}

func (entity *SoftwarePkg) CheckItems() []CheckItem {
	other := entity.otherCheckItems()

	r := make([]CheckItem, 0, len(other)+len(commonCheckItems))
	r = append(r, commonCheckItems...) // don't change commonCheckItems by copy it.
	r = append(r, other...)

	return r
}

func (entity *SoftwarePkg) otherCheckItems() []CheckItem {
	sig := entity.Sig.ImportingPkgSig()
	v := []CheckItem{
		{
			Id:            entity.Sig.ImportingPkgSig(),
			Name:          "Sig",
			Desc:          fmt.Sprintf("软件包被%s Sig接纳", sig),
			NameEn:        "Sig",
			DescEn:        fmt.Sprintf("The software package is accepted by the sig of %s", sig),
			Owner:         dp.CommunityRoleSigMaintainer,
			OnlyOwner:     true,
			Modifications: []string{pkgModificationSig},
		},
	}

	cs := entity.Repo.Committers
	for i := range cs {
		c := cs[i].Account.Account()

		v = append(v, CheckItem{
			Id:            c,
			Name:          "软件包维护人",
			Desc:          fmt.Sprintf("%s 同意作为此软件包的维护人", c),
			NameEn:        "Software Package Maintainer",
			DescEn:        fmt.Sprintf("%s agrees to be the maintainer of this package", c),
			Owner:         dp.CommunityRoleCommitter,
			Keep:          true,
			OnlyOwner:     true,
			Modifications: []string{pkgModificationCommitter},
		})
	}

	return v
}

// CheckItem
type CheckItem struct {
	Id     string
	Name   string
	Desc   string
	NameEn string
	DescEn string
	Owner  dp.CommunityRole

	// If true, keep the review record of reviewer, otherwise clear all the records about
	// this item when relevant modifications happened.
	// For example, the review about the item whether the user aggreed to
	// to be committer of the pkg should be kept.
	Keep bool

	// If true, only the owner can review this item else anyone can review.
	// For example, onlye sig maintainer can determine whether the sig of pkg is correct.
	OnlyOwner bool

	// This check item should be checked again when the relevant modifications happened.
	Modifications []string
}

func (item *CheckItem) GetName(lang dp.Language) string {
	switch lang.Language() {
	case dp.Chinese:
		return item.Name

	case dp.English:
		return item.NameEn

	default:
		return ""
	}
}

func (item *CheckItem) GetDesc(lang dp.Language) string {
	switch lang.Language() {
	case dp.Chinese:
		return item.Desc

	case dp.English:
		return item.DescEn

	default:
		return ""
	}
}

func (item *CheckItem) isOwnerOfItem(roles map[string]bool) bool {
	return roles != nil && roles[item.Owner.CommunityRole()]
}

func (item *CheckItem) canReview(roles map[string]bool) bool {
	return !item.OnlyOwner || item.isOwnerOfItem(roles)
}

func (item *CheckItem) needRecheck(ms map[string]bool) bool {
	for _, v := range item.Modifications {
		if ms[v] {
			return true
		}
	}

	return false
}

func (item *CheckItem) OwnerDesc(pkg *SoftwarePkg) string {
	switch item.Owner.CommunityRole() {
	case dp.CommunityRoleTC.CommunityRole():
		return "TC members"

	case dp.CommunityRoleCommitter.CommunityRole():
		return item.Id

	case dp.CommunityRoleSigMaintainer.CommunityRole():
		return item.Id + " Sig Maintainer"

	case dp.CommunityRoleRepoMember.CommunityRole():
		return fmt.Sprintf(
			"%s Sig Maintainer or committers: %s",
			pkg.Sig.ImportingPkgSig(),
			strings.Join(pkg.Repo.CommitterIds(), ", "),
		)

	default:
		return ""
	}
}
