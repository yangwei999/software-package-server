package app

import (
	"github.com/sirupsen/logrus"

	commonrepo "github.com/opensourceways/software-package-server/common/domain/repository"
	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/message"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/pkgmanager"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/repository"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/service"
)

type SoftwarePkgService interface {
	ApplyNewPkg(*CmdToApplyNewSoftwarePkg) (NewSoftwarePkgDTO, error)

	Close(*CmdToClosePkg) error

	Update(*CmdToUpdateSoftwarePkgApplication) error
	Retest(string, *domain.User) error

	Get(string, dp.Language) (SoftwarePkgDTO, error)
	ListPkgs(*CmdToListPkgs) (SoftwarePkgSummariesDTO, error)

	Review(pid string, user *domain.Reviewer, reviews []domain.CheckItemReviewInfo) (err error)
	GetReview(pid string, user *domain.User, lang dp.Language) ([]CheckItemUserReviewDTO, error)
}

var (
	_ SoftwarePkgService = (*softwarePkgService)(nil)

	softwarePkgRobot = "software-pkg-robot"
)

func NewSoftwarePkgService(
	repo repository.SoftwarePkg,
	manager pkgmanager.PkgManager,
	message message.SoftwarePkgMessage,
	commentRepo repository.SoftwarePkgComment,
) *softwarePkgService {
	robot, _ := dp.NewAccount(softwarePkgRobot)

	return &softwarePkgService{
		repo:        repo,
		robot:       robot,
		message:     message,
		pkgService:  service.NewPkgService(manager, message),
		commentRepo: commentRepo,
	}
}

type softwarePkgService struct {
	repo        repository.SoftwarePkg
	robot       dp.Account
	message     message.SoftwarePkgMessage
	pkgService  service.SoftwarePkgService
	commentRepo repository.SoftwarePkgComment
}

func (s *softwarePkgService) ApplyNewPkg(cmd *CmdToApplyNewSoftwarePkg) (
	dto NewSoftwarePkgDTO, err error,
) {
	v := domain.NewSoftwarePkg(
		cmd.Sig, &cmd.Repo, &cmd.Basic, cmd.Spec, cmd.SRPM, &cmd.Importer,
	)
	if s.pkgService.IsPkgExisted(cmd.Basic.Name) {
		/* TODO recover later
		err = errorSoftwarePkgExists

		return
		*/
	}

	if err = s.repo.Add(&v); err != nil {
		if commonrepo.IsErrorDuplicateCreating(err) {
			err = errorSoftwarePkgExists
		}

		return
	}

	dto.Id = v.Id

	e := domain.NewSoftwarePkgAppliedEvent(&v)
	if err1 := s.message.SendPkgAppliedEvent(&e); err1 != nil {
		logrus.Errorf(
			"failed to send pkg applied event, pkg:%s, err:%s",
			v.Id, err1.Error(),
		)
	} else {
		logrus.Debugf(
			"successfully to send pkg applied event, pkg:%s", v.Id,
		)
	}

	return
}

func (s *softwarePkgService) ListPkgs(cmd *CmdToListPkgs) (dto SoftwarePkgSummariesDTO, err error) {
	v, total, err := s.repo.FindAll(cmd)
	if err != nil {
		return
	}

	dto.Total = total

	if len(v) > 0 {
		dto.Pkgs = toSoftwarePkgSummaryDTOs(v)
	}

	return
}

func (s *softwarePkgService) Update(cmd *CmdToUpdateSoftwarePkgApplication) error {
	pkg, version, err := s.repo.Find(cmd.PkgId)
	if err != nil {
		return parseErrorForFindingPkg(err)
	}

	if err = pkg.Update(&cmd.Importer, &cmd.SoftwarePkgUpdateInfo); err != nil {
		return err
	}

	if err = s.repo.Save(&pkg, version); err != nil {
		return err
	}

	if cmd.Spec != nil || cmd.SRPM != nil {
		// it may need reload even if the spec or srpm does not change.
		e := domain.NewSoftwarePkgCodeUpdatedEvent(&pkg)

		if err1 := s.message.SendPkgCodeUpdatedEvent(&e); err1 != nil {
			logrus.Errorf(
				"failed to send pkg code updated event, pkg:%s, err:%s",
				pkg.Id, err1.Error(),
			)
		} else {
			logrus.Debugf(
				"successfully to send pkg code updated event , pkg:%s", pkg.Id,
			)
		}
	}

	return nil
}

func (s *softwarePkgService) Close(cmd *CmdToClosePkg) error {
	pkg, version, err := s.repo.FindAndIgnoreReview(cmd.PkgId)
	if err != nil {
		return parseErrorForFindingPkg(err)
	}

	if err = pkg.Close(&cmd.Reviewer); err != nil {
		return err
	}

	if err := s.repo.SaveAndIgnoreReview(&pkg, version); err != nil {
		return err
	}

	if cmd.Comment == nil {
		return nil
	}

	comment := domain.NewSoftwarePkgReviewComment(cmd.Reviewer.Account, cmd.Comment)
	if err := s.commentRepo.AddReviewComment(cmd.PkgId, &comment); err != nil {
		logrus.Errorf(
			"failed to add a comment when closing a pkg:%s, err:%s",
			cmd.PkgId, err.Error(),
		)
	}

	e := domain.NewSoftwarePkgClosedEvent(&pkg)
	if err = s.message.SendPkgClosedEvent(&e); err != nil {
		logrus.Errorf(
			"failed to send pkg closed event, pkg:%s, err:%s",
			cmd.PkgId, err.Error(),
		)
	}

	return nil
}

func (s *softwarePkgService) Retest(pid string, user *domain.User) error {
	pkg, version, err := s.repo.FindAndIgnoreReview(pid)
	if err != nil {
		return parseErrorForFindingPkg(err)
	}

	if err = pkg.Retest(user); err != nil {
		return err
	}

	if err = s.repo.SaveAndIgnoreReview(&pkg, version); err != nil {
		return err
	}

	e := domain.NewSoftwarePkgRetestedEvent(&pkg)
	if err = s.message.SendPkgRetestedEvent(&e); err != nil {
		return err
	}

	s.addCommentToRetest(pid)

	return nil
}

func (s *softwarePkgService) addCommentToRetest(pkgId string) {
	content, _ := dp.NewReviewCommentInternal("The CI will run now.")
	comment := domain.NewSoftwarePkgReviewComment(s.robot, content)

	if err := s.commentRepo.AddReviewComment(pkgId, &comment); err != nil {
		logrus.Errorf(
			"failed to add a comment when retest for pkg:%s, err:%s",
			pkgId, err.Error(),
		)
	}
}

func (s *softwarePkgService) Get(pid string, lang dp.Language) (dto SoftwarePkgDTO, err error) {
	v, _, err := s.repo.Find(pid)
	if err != nil {
		err = parseErrorForFindingPkg(err)

		return
	}

	toSoftwarePkgDTO(&v, &dto, lang)

	return
}
