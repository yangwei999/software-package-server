package app

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/maintainer"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/message"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/repository"
)

type SoftwarePkgMessageService interface {
	HandlePkgPRClosed(CmdToHandlePkgPRClosed) error
	HandlePkgPRMerged(CmdToHandlePkgPRMerged) error
	HandlePkgPRCIChecked(CmdToHandlePkgPRCIChecked) error
	HandlePkgRepoCreated(CmdToHandlePkgRepoCreated) error
}

type softwarePkgMessageService struct {
	repo       repository.SoftwarePkg
	message    message.SoftwarePkgIndirectMessage
	maintainer maintainer.Maintainer
}

func NewSoftwarePkgMessageService(
	repo repository.SoftwarePkg,
	message message.SoftwarePkgIndirectMessage,
	maintainer maintainer.Maintainer,
) softwarePkgMessageService {
	return softwarePkgMessageService{
		repo:       repo,
		message:    message,
		maintainer: maintainer,
	}
}

// HandlePkgPRCIChecked
func (s softwarePkgMessageService) HandlePkgPRCIChecked(cmd CmdToHandlePkgPRCIChecked) error {
	pkg, version, err := s.repo.FindSoftwarePkgBasicInfo(cmd.PkgId)
	if err != nil {
		return err
	}

	alreadyClosed, err := pkg.HandleCI(cmd.isSuccess(), cmd.RelevantPR)
	if err != nil {
		return err
	}

	if alreadyClosed {
		if cmd.isSuccess() {
			s.notifyPkgAlreadyClosed(&cmd)
		}

		return nil
	}

	if !cmd.isSuccess() {
		s.addCommentForFailedCI(&cmd)
	}

	pkg.PRNum = cmd.PRNum
	if err := s.repo.SaveSoftwarePkg(&pkg, version); err != nil {
		logrus.Errorf(
			"save pkg failed when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}

	return nil
}

func (s softwarePkgMessageService) notifyPkgAlreadyClosed(cmd *CmdToHandlePkgPRCIChecked) {
	if !cmd.isSuccess() {
		return
	}

	e := domain.NewSoftwarePkgAlreadyClosedEvent(cmd.PkgId, cmd.PRNum)

	if err := s.message.NotifyPkgAlreadyClosed(&e); err != nil {
		logrus.Errorf(
			"failed to notify the pkg is already closed when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}
}

func (s softwarePkgMessageService) addCommentForFailedCI(cmd *CmdToHandlePkgPRCIChecked) {
	author, _ := dp.NewAccount("software-pkg-robot")

	str := fmt.Sprintf(
		"I'am sorry to close this application. Because the checking failed with the reason as bellow.\n\n%s",
		cmd.FiledReason,
	)
	content, _ := dp.NewReviewComment(str)

	comment := domain.NewSoftwarePkgReviewComment(author, content)

	if err := s.repo.AddReviewComment(cmd.PkgId, &comment); err != nil {
		logrus.Errorf(
			"failed to add a comment when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}
}

// HandlePkgRepoCreated
func (s softwarePkgMessageService) HandlePkgRepoCreated(cmd CmdToHandlePkgRepoCreated) error {
	if !cmd.isSuccess() {
		logrus.Errorf(
			"failed to create repo on platform:%s for pkg:%s, err:%s",
			cmd.Platform.PackagePlatform(), cmd.PkgId, cmd.FiledReason,
		)

		return nil
	}

	pkg, version, err := s.repo.FindSoftwarePkgBasicInfo(cmd.PkgId)
	if err != nil {
		return err
	}

	if err := pkg.HandleRepoCreated(cmd.RepoCreatedInfo); err != nil {
		return err
	}

	if err := s.repo.SaveSoftwarePkg(&pkg, version); err != nil {
		logrus.Errorf(
			"save pkg failed when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}

	return nil
}

// HandlePkgPRClosed
func (s softwarePkgMessageService) HandlePkgPRClosed(cmd CmdToHandlePkgPRClosed) error {
	pkg, version, err := s.repo.FindSoftwarePkgBasicInfo(cmd.PkgId)
	if err != nil {
		return err
	}

	user, err := s.validateUser(&pkg, cmd.RejectedBy)
	if err != nil {
		logrus.Errorf(
			"validate user failed when %s, err:%s",
			cmd.logString(), err.Error(),
		)

		return err
	}

	if b, err := pkg.HandleRejectedBy(user); err != nil || b {
		return err
	}

	if err := s.repo.SaveSoftwarePkg(&pkg, version); err != nil {
		logrus.Errorf(
			"save pkg failed when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}

	return nil
}

// HandlePkgPRMerged
func (s softwarePkgMessageService) HandlePkgPRMerged(cmd CmdToHandlePkgPRMerged) error {
	pkg, version, err := s.repo.FindSoftwarePkgBasicInfo(cmd.PkgId)
	if err != nil {
		return err
	}

	users := make([]dp.Account, len(cmd.ApprovedBy))

	for i, item := range cmd.ApprovedBy {
		user, err := s.validateUser(&pkg, item)
		if err != nil {
			logrus.Errorf(
				"validate user failed when %s, err:%s",
				cmd.logString(), err.Error(),
			)

			return err
		}

		users[i] = user
	}

	if b, err := pkg.HandleApprovedBy(users); err != nil || b {
		return err
	}

	if dp.IsPkgReviewResultApproved(pkg.ReviewResult()) {
		s.notifyPkgIndirectlyApproved(&pkg, &cmd)
	}

	if err := s.repo.SaveSoftwarePkg(&pkg, version); err != nil {
		logrus.Errorf(
			"save pkg failed when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}

	return nil
}

func (s softwarePkgMessageService) notifyPkgIndirectlyApproved(
	pkg *domain.SoftwarePkgBasicInfo, cmd *CmdToHandlePkgPRMerged,
) {
	e := domain.NewSoftwarePkgIndirectlyApprovedEvent(pkg)

	if err := s.message.NotifyPkgIndirectlyApproved(&e); err != nil {
		logrus.Errorf(
			"failed to notify the pkg was approved indirectly when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}
}

func (s softwarePkgMessageService) validateUser(pkg *domain.SoftwarePkgBasicInfo, v string) (
	dp.Account, error,
) {
	user, err := s.maintainer.FindUser(v)
	if err != nil {
		return nil, err
	}

	has, err := s.maintainer.HasPermission(pkg, user)
	if err != nil {
		return nil, err
	}

	if !has {
		return nil, fmt.Errorf("no permission")
	}

	return user, nil
}
