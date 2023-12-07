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
		if err := impl.createReviewDetailComment(&v, items, itemsIdMap, prNum); err != nil {
			logrus.Errorf("create review comment of %s err: %s", v.Reviewer.Account.Account(), err.Error())
		}
	}
}

func (impl *pullRequestImpl) createCheckItemsComment(
	items []domain.CheckItem,
	itemsMap map[string]string,
	prNum int,
) error {
	var itemsTpl []checkItemTpl
	for _, v := range items {
		itemsTpl = append(itemsTpl, checkItemTpl{
			Id:   itemsMap[v.Id],
			Name: v.Name,
			Desc: v.Desc,
		})
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
	items []domain.CheckItem,
	itemsMap map[string]string,
	prNUm int,
) error {

	var itemsTpl []*checkItemTpl
	for _, v := range review.Reviews {
		itemTpl := impl.findCheckItem(v.Id, items)
		if itemTpl == nil {
			continue
		}

		itemTpl.Id = itemsMap[v.Id]
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

func (impl *pullRequestImpl) findCheckItem(id string, items []domain.CheckItem) *checkItemTpl {
	for _, v := range items {
		if v.Id == id {
			return &checkItemTpl{
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

func (impl *pullRequestImpl) itemsIdMap(items []domain.CheckItem) map[string]string {
	m := make(map[string]string)
	for i, v := range items {
		m[v.Id] = strconv.Itoa(i + 1)
	}

	return m
}
