package app

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/email"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/pullrequest"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/repository"
)

type SoftwarePkgWatchService interface {
	AddPkgWatch(*domain.PkgWatch) error
	FindPkgWatch() ([]*domain.PkgWatch, error)
	HandleCreatePR(*domain.PkgWatch, *domain.SoftwarePkg) error
	HandleUpdatePR(*domain.SoftwarePkg) error
	HandleCI(*CmdToHandleCI) error
	HandlePRMerged(*domain.PkgWatch) error
	HandlePRClosed(*CmdToHandlePRClosed) error
	HandleDone(*domain.PkgWatch) error
}

func NewWatchService(pr pullrequest.PullRequest, r repository.Watch, e email.Email) *softwarePkgWatchService {
	return &softwarePkgWatchService{
		prCli:     pr,
		watchRepo: r,
		email:     e,
	}
}

type softwarePkgWatchService struct {
	prCli     pullrequest.PullRequest
	watchRepo repository.Watch
	email     email.Email
}

func (s *softwarePkgWatchService) AddPkgWatch(pw *domain.PkgWatch) error {
	return s.watchRepo.Add(pw)
}

func (s *softwarePkgWatchService) FindPkgWatch() ([]*domain.PkgWatch, error) {
	return s.watchRepo.FindAll()
}

func (s *softwarePkgWatchService) HandleCreatePR(watchPkg *domain.PkgWatch, pkg *domain.SoftwarePkg) error {
	pr, err := s.prCli.Create(pkg)
	if err != nil {
		return err
	}

	watchPkg.PR = pr
	watchPkg.SetPkgStatusPRCreated()

	return s.watchRepo.Save(watchPkg)
}

func (s *softwarePkgWatchService) HandleUpdatePR(pkg *domain.SoftwarePkg) error {
	return s.prCli.Update(pkg)
}

func (s *softwarePkgWatchService) HandleCI(cmd *CmdToHandleCI) error {
	if cmd.IsSuccess {
		if err := s.mergePR(cmd.PkgWatch); err != nil {
			logrus.Errorf("merge pr %d failed: %s", cmd.PR.Num, err.Error())

			return s.notifyException(cmd.PkgWatch, err.Error())
		}
	} else {
		if err := s.prCli.Close(cmd.PR.Num); err != nil {
			logrus.Errorf("close pr/%d failed: %s", cmd.PR.Num, err.Error())
		}

		return s.notifyException(cmd.PkgWatch, "ci check failed")
	}

	return nil
}

func (s *softwarePkgWatchService) mergePR(pw *domain.PkgWatch) error {
	if err := s.prCli.Merge(pw.PR.Num); err != nil {
		return fmt.Errorf("merge pr(%d) failed: %s", pw.PR.Num, err.Error())
	}

	pw.SetPkgStatusPRMerged()

	if err := s.watchRepo.Save(pw); err != nil {
		logrus.Errorf("save pr(%d) failed: %s", pw.PR.Num, err.Error())
	}

	return nil
}

func (s *softwarePkgWatchService) HandlePRMerged(pw *domain.PkgWatch) error {
	if pw.IsPkgStatusMerged() {
		return nil
	}

	pw.SetPkgStatusPRMerged()

	return s.watchRepo.Save(pw)
}

func (s *softwarePkgWatchService) HandlePRClosed(cmd *CmdToHandlePRClosed) error {
	subject := fmt.Sprintf(
		"the pr of software package was closed by: %s",
		cmd.RejectedBy,
	)
	content := s.emailContent(cmd.PR.Link)

	if err := s.email.Send(subject, content); err != nil {
		return fmt.Errorf("send email failed: %s", err.Error())
	}

	cmd.PkgWatch.SetPkgStatusException()

	return s.watchRepo.Save(cmd.PkgWatch)
}

func (s *softwarePkgWatchService) HandleDone(pw *domain.PkgWatch) error {
	pw.SetPkgStatusDone()

	return s.watchRepo.Save(pw)
}

func (s *softwarePkgWatchService) emailContent(url string) string {
	return fmt.Sprintf("th pr url is: %s", url)
}

func (s *softwarePkgWatchService) notifyException(
	pw *domain.PkgWatch, reason string,
) error {
	subject := fmt.Sprintf(
		"the ci of software package check failed: %s",
		reason,
	)
	content := s.emailContent(pw.PR.Link)

	if err := s.email.Send(subject, content); err != nil {
		return fmt.Errorf("send email failed: %s", err.Error())
	}

	pw.SetPkgStatusException()

	if err := s.watchRepo.Save(pw); err != nil {
		return fmt.Errorf("save pkg when exception error: %s", err.Error())
	}

	return nil
}
