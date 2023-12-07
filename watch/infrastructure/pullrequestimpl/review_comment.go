package pullrequestimpl

import (
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
)

var checkItemResult = map[bool]string{
	true:  "通过",
	false: "不通过",
}

func (impl *pullRequestImpl) addReviewComment(pkg *domain.SoftwarePkg, prNum int) {
	items := pkg.CheckItems()
	itemsIdMap := impl.itemsIdMap(items)

	if err := impl.createCheckItemsComment(items, itemsIdMap, prNum); err != nil {
		logrus.Errorf("add check items comment err: %s", err.Error())
	}

	for _, v := range pkg.Reviews {
		if err := impl.createReviewDetailComment(&v, itemsIdMap, prNum); err != nil {
			logrus.Errorf("create review comment of %s err: %s", v.Reviewer.Account.Account(), err.Error())
		}
	}
}

func (impl *pullRequestImpl) createCheckItemsComment(
	items []domain.CheckItem,
	itemsMap map[string]checkItemTpl,
	prNum int,
) error {
	var itemsTpl []checkItemTpl
	for _, v := range items {
		itemsTpl = append(itemsTpl, itemsMap[v.Id])
	}

	body, err := impl.template.genCheckItems(&checkItemsTplData{
		CheckItems: itemsTpl,
	})
	if err != nil {
		return err
	}

	return impl.comment(prNum, body)
}

func (impl *pullRequestImpl) createReviewDetailComment(
	review *domain.UserReview,
	itemsMap map[string]checkItemTpl,
	prNUm int,
) error {

	var itemsTpl []checkItemTpl
	for _, v := range review.Reviews {
		itemTpl, ok := itemsMap[v.Id]
		if !ok {
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

func (impl *pullRequestImpl) comment(prNum int, content string) error {
	return impl.cli.CreatePRComment(
		impl.cfg.CommunityRobot.Org, impl.cfg.CommunityRobot.Repo,
		int32(prNum), content,
	)
}

func (impl *pullRequestImpl) itemsIdMap(items []domain.CheckItem) map[string]checkItemTpl {
	m := make(map[string]checkItemTpl)
	for i, v := range items {
		m[v.Id] = checkItemTpl{
			Id:   strconv.Itoa(i + 1),
			Name: v.Name,
			Desc: v.Desc,
		}
	}

	return m
}
