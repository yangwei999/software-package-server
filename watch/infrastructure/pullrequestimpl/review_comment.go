package pullrequestimpl

import (
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
)

var checkItemResult = map[bool]string{
	true:  "通过",
	false: "不通过",
}

func (impl *pullRequestImpl) addReviewComment(pkg *domain.SoftwarePkg, prNum int) {
	if err := impl.createCheckItemsComment(prNum); err != nil {
		logrus.Errorf("add check items comment err: %s", err.Error())
	}

	for _, v := range pkg.Reviews {
		if err := impl.createReviewDetailComment(&v, prNum); err != nil {
			logrus.Errorf("create review comment of %s err: %s", v.Reviewer.Account.Account(), err.Error())
		}
	}
}

func (impl *pullRequestImpl) createCheckItemsComment(prNum int) error {
	body, err := impl.template.genCheckItems(impl.cfg.Config)
	if err != nil {
		return err
	}

	return impl.comment(prNum, body)
}

func (impl *pullRequestImpl) createReviewDetailComment(review *domain.UserReview, prNUm int) error {
	var items []*checkItem

	var localReviews Reviews = review.Reviews
	sort.Sort(localReviews)

	for _, v := range localReviews {
		item := impl.findCheckItem(v.Id)
		if item == nil {
			continue
		}

		item.Result = checkItemResult[v.Pass]
		if v.Comment != nil {
			item.Comment = v.Comment.ReviewComment()
		}

		items = append(items, item)
	}

	body, err := impl.template.genReviewDetail(&reviewDetailTplData{
		Reviewer:   review.Account.Account(),
		CheckItems: items,
	})
	if err != nil {
		return err
	}

	return impl.comment(prNUm, body)
}

func (impl *pullRequestImpl) findCheckItem(id string) *checkItem {
	for _, v := range impl.cfg.CheckItems {
		if v.Id == id {
			return &checkItem{
				Id:   id,
				Name: v.Name,
				Desc: v.Desc,
			}
		}
	}

	return nil
}

func (impl *pullRequestImpl) comment(prNum int, content string) error {
	return impl.cli.CreatePRComment(
		impl.cfg.CommunityRobot.Org, impl.cfg.CommunityRobot.Repo,
		int32(prNum), content,
	)
}

type Reviews []domain.CheckItemReviewInfo

func (r Reviews) Len() int {
	return len(r)
}

func (r Reviews) Less(i, j int) bool {
	t := strings.Compare(r[i].Id, r[j].Id)

	return t == -1
}

func (r Reviews) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
