package service

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/message"
)

type SoftwarePkgReviewService interface {
	ApprovePkg(pkg *domain.SoftwarePkgBasicInfo, user dp.Account) error
	RejectPkg(pkg *domain.SoftwarePkgBasicInfo, user dp.Account) error
	AbandonPkg(pkg *domain.SoftwarePkgBasicInfo, user dp.Account) error
}

func NewReviewService(m message.SoftwarePkgMessage) SoftwarePkgReviewService {
	return &reviewService{message: m}
}

type reviewService struct {
	message message.SoftwarePkgMessage
}

func (s *reviewService) ApprovePkg(pkg *domain.SoftwarePkgBasicInfo, user dp.Account) error {
	if approved, err := pkg.ApproveBy(user); !approved {
		return err
	}

	e, err := domain.NewSoftwarePkgApprovedEvent(pkg)
	if err == nil {
		err = s.message.NotifyPkgApproved(&e)
	}
	s.log(pkg, "approved", err)

	return nil
}

func (s *reviewService) RejectPkg(pkg *domain.SoftwarePkgBasicInfo, user dp.Account) error {
	if rejected, err := pkg.RejectBy(user); !rejected {
		return err
	}

	e := domain.NewSoftwarePkgRejectedEvent(pkg)
	err := s.message.NotifyPkgRejected(&e)

	s.log(pkg, "rejected", err)

	return nil
}

func (s *reviewService) AbandonPkg(pkg *domain.SoftwarePkgBasicInfo, user dp.Account) error {
	if err := pkg.Abandon(user); err != nil {
		return err
	}

	e := domain.NewSoftwarePkgAbandonedEvent(pkg)
	err := s.message.NotifyPkgAbandoned(&e)

	s.log(pkg, "abandoned", err)

	return nil
}

func (s *reviewService) log(pkg *domain.SoftwarePkgBasicInfo, op string, err error) {
	msg := fmt.Sprintf(
		"notify that a pkg:%s/%s was %s", pkg.Id, pkg.PkgName.PackageName(), op,
	)

	if err == nil {
		logrus.Debugf("successfully to %s", msg)
	} else {
		logrus.Errorf("failed to %s, err:%s", msg, err.Error())
	}
}
