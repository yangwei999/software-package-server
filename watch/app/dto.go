package app

import "github.com/opensourceways/software-package-server/watch/domain"

type CmdToHandleCI struct {
	*domain.PkgWatch
	IsSuccess bool
}

type CmdToHandlePRClosed struct {
	*domain.PkgWatch
	RejectedBy string
}
