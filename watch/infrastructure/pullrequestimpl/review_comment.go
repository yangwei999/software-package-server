package pullrequestimpl

import (
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
)

var checkItemResult = map[bool]string{
	true:  "通过",
	false: "不通过",
}

func (impl *pullRequestImpl) addReviewComment(pkg *domain.SoftwarePkg, prNum int) {
	var items localCheckItems = pkg.CheckItems()

	if err := impl.createCheckItemsComment(prNum, items); err != nil {
		logrus.Errorf("add check items comment err: %s", err.Error())
	}

	for _, v := range pkg.Reviews {
		if err := impl.createReviewDetailComment(items, &v, prNum); err != nil {
			logrus.Errorf("create review comment of %s err: %s", v.Reviewer.Account.Account(), err.Error())
		}
	}
}

func (impl *pullRequestImpl) createCheckItemsComment(prNum int, items []domain.CheckItem) error {
	body, err := impl.template.genCheckItems(&checkItemsTplData{
		CheckItems: items,
	})
	if err != nil {
		return err
	}

	return impl.comment(prNum, body)
}

func (impl *pullRequestImpl) createReviewDetailComment(
	items localCheckItems,
	review *domain.UserReview,
	prNUm int,
) error {

	var itemsTpl []*checkItemTpl
	for _, v := range review.Reviews {
		itemTpl := impl.findCheckItem(v.Id, items)
		if itemTpl == nil {
			continue
		}

		itemTpl.Result = checkItemResult[v.Pass]
		if v.Comment != nil {
			itemTpl.Comment = v.Comment.ReviewComment()
		}

		itemsTpl = append(itemsTpl, itemTpl)
	}

	body, err := impl.template.genReviewDetail(&reviewDetailTplData{
		Reviewer:   review.Account.Account(),
		CheckItems: itemsTpl,
	})
	if err != nil {
		return err
	}

	return impl.comment(prNUm, body)
}

func (impl *pullRequestImpl) findCheckItem(id string, items localCheckItems) *checkItemTpl {
	for _, v := range items {
		if v.Id == id {
			return &checkItemTpl{
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

type localCheckItems []domain.CheckItem
