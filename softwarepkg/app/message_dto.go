package app

import (
	"fmt"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
)

// CmdToHandlePkgPRCIChecked
type CmdToHandlePkgPRCIChecked struct {
	PkgId       string
	RelevantPR  dp.URL
	PRNum       int
	FiledReason string
}

func (cmd *CmdToHandlePkgPRCIChecked) isSuccess() bool {
	return cmd.FiledReason == ""
}

func (cmd *CmdToHandlePkgPRCIChecked) logString() string {
	return fmt.Sprintf(
		"handling pkg ci checked, pkgid:%s, pr:%s",
		cmd.PkgId, cmd.RelevantPR.URL(),
	)
}

// CmdToHandlePkgRepoCreated
type CmdToHandlePkgRepoCreated struct {
	PkgId       string
	FiledReason string

	domain.RepoCreatedInfo
}

func (cmd *CmdToHandlePkgRepoCreated) isSuccess() bool {
	return cmd.FiledReason == ""
}

func (cmd *CmdToHandlePkgRepoCreated) logString() string {
	if !cmd.isSuccess() {
		return ""
	}

	return fmt.Sprintf(
		"handling pkg repo created, pkgid:%s, platform:%s, repo:%s",
		cmd.PkgId, cmd.Platform.PackagePlatform(), cmd.RepoLink.URL(),
	)
}

// CmdToHandlePkgPRClosed
type CmdToHandlePkgPRClosed struct {
	PkgId      string
	Reason     string
	RejectedBy string
}

func (cmd *CmdToHandlePkgPRClosed) logString() string {
	return fmt.Sprintf(
		"handling pkg pr closed, pkgid:%s, reason:%s, rejected by:%s",
		cmd.PkgId, cmd.Reason, cmd.RejectedBy,
	)
}

// CmdToHandlePkgPRMerged
type CmdToHandlePkgPRMerged struct {
	PkgId      string
	ApprovedBy []string
}

func (cmd *CmdToHandlePkgPRMerged) logString() string {
	return fmt.Sprintf(
		"handling pkg pr merged, pkgid:%s, approved by:%v",
		cmd.PkgId, cmd.ApprovedBy,
	)
}
