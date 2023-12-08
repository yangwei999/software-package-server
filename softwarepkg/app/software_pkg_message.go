package app

import (
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/message"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/pkgcodeadapter"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/pkgmanager"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/repository"
)

type SoftwarePkgMessageService interface {
	DownloadPkgCode(cmd CmdToDownloadPkgCode) error
	StartCI(cmd CmdToStartCI) error
	HandlePkgCIDone(CmdToHandlePkgCIDone) error

	HandlePkgClosed(*domain.SoftwarePkgClosedEvent) error
	HandlePkgRepoCodePushed(CmdToHandlePkgRepoCodePushed) error

	ImportPkg(CmdToHandlePkgAlreadyExisted) error
}

func NewSoftwarePkgMessageService(
	code pkgcodeadapter.PkgCodeAdapter,
	repo repository.SoftwarePkg,
	manager pkgmanager.PkgManager,
	message message.SoftwarePkgIndirectMessage,
	commentRepo repository.SoftwarePkgComment,
) softwarePkgMessageService {
	robot, _ := dp.NewAccount(softwarePkgRobot)

	return softwarePkgMessageService{
		code:        code,
		repo:        repo,
		robot:       robot,
		manager:     manager,
		message:     message,
		commentRepo: commentRepo,
	}
}

type softwarePkgMessageService struct {
	code        pkgcodeadapter.PkgCodeAdapter
	repo        repository.SoftwarePkg
	robot       dp.Account
	manager     pkgmanager.PkgManager
	message     message.SoftwarePkgIndirectMessage
	commentRepo repository.SoftwarePkgComment
}

// DownloadPkgCode
func (s softwarePkgMessageService) DownloadPkgCode(cmd CmdToDownloadPkgCode) error {
	pkg, version, err := s.repo.FindAndIgnoreReview(cmd.PkgId)
	if err != nil {
		return err
	}

	files := pkg.FilesToDownload()
	if len(files) == 0 {
		return nil
	}

	changed, err := s.code.Download(files, pkg.Basic.Name)
	if err != nil {
		return err
	}

	// update
	pkg1, version, err := s.repo.FindAndIgnoreReview(cmd.PkgId)
	if err != nil {
		return err
	}

	if !pkg1.SaveDownloadedFiles(files) {
		return nil
	}

	if err = s.repo.SaveAndIgnoreReview(&pkg1, version); err != nil {
		logrus.Errorf(
			"save pkg failed when %s, err:%s",
			cmd.logString(), err.Error(),
		)

		return err
	}

	if changed {
		// CI can be started manually. So don't notify everytime.
		s.notifyPkgCodeChanged(&pkg1)
	}

	return nil
}

func (s softwarePkgMessageService) notifyPkgCodeChanged(pkg *domain.SoftwarePkg) {
	e := domain.NewSoftwarePkgCodeChangeedEvent(pkg)

	if err := s.message.SendPkgCodeChangedEvent(&e); err != nil {
		logrus.Errorf(
			"failed to send pkg code changed event, pkg:%s, err:%s",
			pkg.Id, err.Error(),
		)
	}
}

// StartCI
func (s softwarePkgMessageService) StartCI(cmd CmdToStartCI) error {
	pkg, version, err := s.repo.FindAndIgnoreReview(cmd.PkgId)
	if err != nil {
		return err
	}

	if err = pkg.StartCI(); err != nil {
		return err
	}

	if err = s.repo.SaveAndIgnoreReview(&pkg, version); err != nil {
		logrus.Errorf(
			"save pkg failed when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}

	return nil
}

// HandlePkgCIDone
func (s softwarePkgMessageService) HandlePkgCIDone(cmd CmdToHandlePkgCIDone) error {
	pkg, version, err := s.repo.FindAndIgnoreReview(cmd.PkgId)
	if err != nil {
		return err
	}

	if err := pkg.HandleCIDone(cmd.PRNumber, cmd.Success); err != nil {
		return err
	}

	if err = s.repo.SaveAndIgnoreReview(&pkg, version); err != nil {
		logrus.Errorf(
			"save pkg failed when %s, err:%s",
			cmd.logString(), err.Error(),
		)

		return err
	}

	s.addCIComment(&cmd)

	return nil
}

func (s softwarePkgMessageService) addCIComment(cmd *CmdToHandlePkgCIDone) {
	content, _ := dp.NewReviewCommentInternal(cmd.Detail)
	comment := domain.NewSoftwarePkgReviewComment(s.robot, content)

	if err := s.commentRepo.AddReviewComment(cmd.PkgId, &comment); err != nil {
		logrus.Errorf(
			"failed to add a comment when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}
}

// HandlePkgRepoCodePushed
func (s softwarePkgMessageService) HandlePkgRepoCodePushed(cmd CmdToHandlePkgRepoCodePushed) error {
	pkg, version, err := s.repo.FindAndIgnoreReview(cmd.PkgId)
	if err != nil {
		return err
	}

	if err := pkg.HandleRepoCodePushed(); err != nil {
		return err
	}

	if err = s.repo.SaveAndIgnoreReview(&pkg, version); err != nil {
		logrus.Errorf(
			"save pkg failed when %s, err:%s",
			cmd.logString(), err.Error(),
		)

		return err
	}

	if err := s.code.Clear(pkg.CIId(), pkg.PackageName()); err != nil {
		logrus.Errorf("failed to clear pkg, pkg:%s, err:%s", pkg.Id, err.Error())
	}

	return nil
}

// HandlePkgClosed
func (s softwarePkgMessageService) HandlePkgClosed(event *domain.SoftwarePkgClosedEvent) error {
	name, err := dp.NewPackageName(event.PackageName)
	if err != nil {
		return nil
	}

	if err := s.code.Clear(event.CIId, name); err != nil {
		logrus.Errorf("failed to clear pkg, pkg:%s, err:%s", event.PkgId, err.Error())
	}

	return nil
}

// ImportPkg
func (s softwarePkgMessageService) ImportPkg(cmd CmdToHandlePkgAlreadyExisted) error {
	return nil
	/* TODO refactor later
	v, err := s.manager.GetPkg(cmd.PkgName)
	if err != nil {
		logrus.Errorf(
			"failed to get pkg info when %s, err:%s",
			cmd.logString(), err.Error(),
		)

		return err
	}

	if err = s.repo.Add(&v); err != nil {
		if commonrepo.IsErrorDuplicateCreating(err) {
			return nil
		}

		logrus.Errorf(
			"failed to add a software pkg when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}

	return err
	*/
}
