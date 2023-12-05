package domain

import (
	"errors"
	"strings"

	"github.com/opensourceways/software-package-server/allerror"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
)

type Reviewer struct {
	Account dp.Account
	GiteeID string
}

type maintainer interface {
	Roles(*SoftwarePkg, *Reviewer) (tc, sigMaitainer bool)
}

var maintainerInstance maintainer

func (pkg *SoftwarePkg) UserReview(user *User) UserReview {
	account := user.Account.Account()

	for i := range pkg.Reviews {
		if item := &pkg.Reviews[i]; item.Account.Account() == account {
			item.initRole(pkg)

			return *item
		}
	}

	v := UserReview{
		Reviewer: Reviewer{
			Account: user.Account,
			GiteeID: user.GiteeID,
		},
	}

	v.initRole(pkg)

	return v
}

func (pkg *SoftwarePkg) addReview(ur *UserReview, items []CheckItem) error {
	if ur.isEmpty() {
		b, i := pkg.oldReviewer(&ur.Reviewer)
		if !b {
			return errors.New("invalid review")
		}

		v := pkg.Reviews
		n := len(v) - 1
		if i != n {
			v[i] = v[n]
		}
		pkg.Reviews = v[:n]

		return nil
	}

	ur.initRole(pkg)

	if err := ur.validate(items); err != nil {
		return err
	}

	if b, i := pkg.oldReviewer(&ur.Reviewer); b {
		pkg.Reviews[i] = *ur
	} else {
		pkg.Reviews = append(pkg.Reviews, *ur)
	}

	return nil
}

func (pkg *SoftwarePkg) oldReviewer(reviewer *Reviewer) (bool, int) {
	for i := range pkg.Reviews {
		if dp.IsSameAccount(pkg.Reviews[i].Reviewer.Account, reviewer.Account) {
			return true, i
		}
	}

	return false, 0
}

func (pkg *SoftwarePkg) doesPassReview(items []CheckItem) bool {
	for i := range pkg.Reviews {
		pkg.Reviews[i].initRole(pkg)
	}

	for i := range items {
		rf := checkItemReview(&items[i], pkg.Reviews)

		if hasResult, pass := rf.Result(); !hasResult || !pass {
			return false
		}
	}

	return true
}

func (pkg *SoftwarePkg) clearReview(pkgms []string, items []CheckItem) {
	ms := map[string]bool{}
	for i := range pkgms {
		ms[pkgms[i]] = true
	}

	for i := range items {
		if v := &items[i]; !v.Keep && v.needRecheck(ms) {
			for j := range pkg.Reviews {
				pkg.Reviews[j].clearOn(v)
			}
		}
	}
}

func (pkg *SoftwarePkg) CheckItemReviews() []CheckItemReview {
	items := pkg.CheckItems()

	for i := range pkg.Reviews {
		pkg.Reviews[i].initRole(pkg)
	}

	r := make([]CheckItemReview, len(items))
	for i := range items {
		r[i] = checkItemReview(&items[i], pkg.Reviews)

	}

	return r
}

func checkItemReview(item *CheckItem, reviews []UserReview) (rf CheckItemReview) {
	rf.Item = item

	if len(reviews) == 0 {
		return
	}

	rs := make([]UserCheckItemReview, 0, len(reviews))

	for i := range reviews {
		if v, exist := reviews[i].checkItemReview(item); exist {
			rs = append(rs, v)
		}
	}

	rf.Reviews = rs

	return
}

// UserReview
type UserReview struct {
	Reviewer

	Reviews []CheckItemReviewInfo

	roles map[string]bool
}

func (r *UserReview) isEmpty() bool {
	return len(r.Reviews) == 0
}

func (r *UserReview) clearOn(item *CheckItem) {
	for i := range r.Reviews {
		if v := &r.Reviews[i]; v.Id == item.Id {
			n := len(r.Reviews) - 1
			if i != n {
				r.Reviews[i] = r.Reviews[n]
			}
			r.Reviews = r.Reviews[:n]

			return
		}
	}
}

func (r *UserReview) initRole(pkg *SoftwarePkg) {
	tc, sigMaitainer := maintainerInstance.Roles(pkg, &r.Reviewer)

	roles := map[string]bool{}

	if tc {
		roles[dp.CommunityRoleTC.CommunityRole()] = true
	}

	if sigMaitainer {
		roles[dp.CommunityRoleSigMaintainer.CommunityRole()] = true
		roles[dp.CommunityRoleRepoMember.CommunityRole()] = true
	}

	if pkg.isCommitter(r.Account) {
		roles[dp.CommunityRoleCommitter.CommunityRole()] = true

		v := r.checkItemReviewInfo(&CheckItem{Id: r.Account.Account()})

		if v != nil && v.Pass {
			roles[dp.CommunityRoleRepoMember.CommunityRole()] = true
		}
	}

	r.roles = roles
}

func (r *UserReview) checkItemReviewInfo(item *CheckItem) *CheckItemReviewInfo {
	for i := range r.Reviews {
		if v := &r.Reviews[i]; v.Id == item.Id {
			return v
		}
	}

	return nil
}

func (r *UserReview) validate(items []CheckItem) error {
	for i := range items {
		info := r.checkItemReviewInfo(&items[i])

		if info != nil && !items[i].canReview(r.roles) {
			return allerror.NewNoPermission("not the owner")
		}
	}

	return nil
}

func (r *UserReview) CheckItemReview(item *CheckItem) (bool, *CheckItemReviewInfo) {
	if v := item.canReview(r.roles); !v {
		return false, nil
	}

	return true, r.checkItemReviewInfo(item)
}

func (r *UserReview) checkItemReview(item *CheckItem) (UserCheckItemReview, bool) {
	info := r.checkItemReviewInfo(item)
	if info == nil {
		return UserCheckItemReview{}, false
	}

	return UserCheckItemReview{
		Reviewer:            &r.Reviewer,
		IsOwner:             item.isOwnerOfItem(r.roles),
		CheckItemReviewInfo: info,
	}, true
}

// CheckItemReview
type CheckItemReview struct {
	Item    *CheckItem
	Reviews []UserCheckItemReview
}

// return has result, pass
func (r *CheckItemReview) Result() (bool, bool) {
	if len(r.Reviews) == 0 {
		return false, false
	}

	pass := false

	for i := range r.Reviews {
		if v := &r.Reviews[i]; v.IsOwner {
			if !v.Pass {
				return true, false
			}

			pass = true
		}
	}

	return pass, pass
}

func (r *CheckItemReview) Stat() (agree, disagree int) {
	for i := range r.Reviews {
		if r.Reviews[i].Pass {
			agree++
		} else {
			disagree++
		}
	}

	return
}

// UserCheckItemReview
type UserCheckItemReview struct {
	*Reviewer
	*CheckItemReviewInfo

	IsOwner bool
}

// CheckItemReviewInfo
type CheckItemReviewInfo struct {
	Id      string
	Pass    bool
	Comment dp.ReviewComment
}

// CheckItem
type CheckItem struct {
	Id    string
	Name  string
	Desc  string
	Owner dp.CommunityRole

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
		return item.Id + " Sig Maintainer or committers: " + strings.Join(pkg.Repo.CommitterIds(), ", ")

	default:
		return ""
	}
}
