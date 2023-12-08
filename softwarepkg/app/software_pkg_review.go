package app

import (
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
)

func (s *softwarePkgService) GetReview(pid string, user *domain.User, lang dp.Language) (
	[]CheckItemUserReviewDTO, error,
) {
	pkg, _, err := s.repo.Find(pid)
	if err != nil {
		return nil, parseErrorForFindingPkg(err)
	}

	userReview := pkg.UserReview(user)

	items := pkg.CheckItems()
	r := make([]CheckItemUserReviewDTO, len(items))

	for i := range items {
		item := &items[i]

		canReview, info := userReview.CheckItemReview(item)

		r[i] = CheckItemUserReviewDTO{
			Id:        item.Id,
			Name:      item.GetName(lang),
			Desc:      item.GetDesc(lang),
			Owner:     item.OwnerDesc(&pkg),
			CanReview: canReview,
		}

		if canReview && info != nil {
			v := &r[i]

			v.Pass = info.Pass
			v.Reviewed = true
			if info.Comment != nil {
				v.Comment = info.Comment.ReviewComment()
			}
		}
	}

	return r, nil
}

func (s *softwarePkgService) Review(
	pid string, user *domain.Reviewer,
	reviews []domain.CheckItemReviewInfo,
) error {
	pkg, version, err := s.repo.Find(pid)
	if err != nil {
		return parseErrorForFindingPkg(err)
	}

	err = pkg.AddReview(&domain.UserReview{
		Reviewer: *user,
		Reviews:  reviews,
	})
	if err != nil {
		return err
	}

	if err := s.repo.Save(&pkg, version); err != nil {
		return err
	}

	s.addCommentOfReview(&pkg, user, reviews)

	return nil
}

func (s *softwarePkgService) addCommentOfReview(
	pkg *domain.SoftwarePkg,
	user *domain.Reviewer,
	reviews []domain.CheckItemReviewInfo,
) {
	itemMap := pkg.CheckItemsMap()

	items := make([]string, len(reviews))
	for i := range reviews {
		items[i] = reviews[i].String(itemMap[reviews[i].Id])
	}

	content, _ := dp.NewReviewCommentInternal(strings.Join(items, "\n"))
	comment := domain.NewSoftwarePkgReviewComment(user.Account, content)

	if err := s.commentRepo.AddReviewComment(pkg.Id, &comment); err != nil {
		logrus.Errorf(
			"failed to add a comment when review for pkg:%s, err:%s",
			pkg.Id, err.Error(),
		)
	}
}
