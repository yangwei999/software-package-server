package pullrequestimpl

import (
	"fmt"
	"sort"

	"github.com/sirupsen/logrus"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
)

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
	sort.Sort(review.Reviews)
	var items []*checkItem
	for _, v := range review.Reviews {
		if item := impl.findCheckItem(v.Id); item != nil {
			item.Result = fmt.Sprintf("%v", v.Pass)
			if v.Comment != nil {
				item.Comment = v.Comment.ReviewComment()
			}
			items = append(items, item)
		}
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
