package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
	"github.com/opensourceways/software-package-server/utils"
)

type Importer struct {
	Email   dp.Email
	Account dp.Account
}

type User struct {
	Importer

	GiteeID string
}

// SoftwarePkgApplication
type SoftwarePkgApplication struct {
	SourceCode        SoftwarePkgSourceCode
	PackageDesc       dp.PackageDesc
	PackagePlatform   dp.PackagePlatform
	ImportingPkgSig   dp.ImportingPkgSig
	ReasonToImportPkg dp.ReasonToImportPkg
	Upstream          dp.URL
}

type SoftwarePkgSourceCode struct {
	SpecURL   dp.URL
	SrcRPMURL dp.URL
}

// SoftwarePkgCI
type SoftwarePkgCI struct {
	PRNum  int
	Status dp.PackageCIStatus
	// TODO deal with the case that the ci is timeout
	//startTime int64
}

func (ci *SoftwarePkgCI) isSuccess() bool {
	return ci.Status != nil && ci.Status.IsCIPassed()
}

// SoftwarePkgApprover
type SoftwarePkgApprover struct {
	Account dp.Account
	IsTC    bool
}

func (approver *SoftwarePkgApprover) String() string {
	return fmt.Sprintf("%s/%v", approver.Account.Account(), approver.IsTC)
}

func StringToSoftwarePkgApprover(s string) (r SoftwarePkgApprover, err error) {
	items := strings.Split(s, "/")

	if r.Account, err = dp.NewAccount(items[0]); err == nil {
		r.IsTC, _ = strconv.ParseBool(items[1])
	}

	return
}

// SoftwarePkgBasicInfo
type SoftwarePkgBasicInfo struct {
	Id          string
	PkgName     dp.PackageName
	Importer    Importer
	RepoLink    dp.URL
	Phase       dp.PackagePhase
	CI          SoftwarePkgCI
	AppliedAt   int64
	Application SoftwarePkgApplication
	ApprovedBy  []SoftwarePkgApprover
	RejectedBy  []SoftwarePkgApprover
	RelevantPR  dp.URL
}

func (entity *SoftwarePkgBasicInfo) ReviewResult() dp.PackageReviewResult {
	if len(entity.RejectedBy) > 0 {
		return dp.PkgReviewResultRejected
	}

	if entity.hasPassedReview() {
		return dp.PkgReviewResultApproved
	}

	return nil
}

func (entity *SoftwarePkgBasicInfo) CanAddReviewComment() bool {
	return entity.Phase.IsReviewing()
}

func (entity *SoftwarePkgBasicInfo) ApproveBy(user *SoftwarePkgApprover) (bool, error) {
	if !entity.Phase.IsReviewing() || !entity.CI.isSuccess() {
		return false, errors.New("can't do this")
	}

	entity.ApprovedBy = append(entity.ApprovedBy, *user)

	b := entity.hasPassedReview()
	if b {
		entity.Phase = dp.PackagePhaseCreatingRepo
	}

	return b, nil
}

func (entity *SoftwarePkgBasicInfo) hasPassedReview() bool {
	sig := entity.Application.ImportingPkgSig.ImportingPkgSig()
	if sig == config.EcopkgSig {
		return len(entity.ApprovedBy) > 0
	}

	numApprovedByTc := 0
	numApprovedBySigMaintainer := 0

	for i := range entity.ApprovedBy {
		if entity.ApprovedBy[i].IsTC {
			numApprovedByTc++
			numApprovedBySigMaintainer++
		} else {
			numApprovedBySigMaintainer++
		}
	}

	c := numApprovedByTc >= config.MinNumApprovedByTC
	c1 := numApprovedBySigMaintainer >= config.MinNumApprovedBySigMaintainer

	return c && c1
}

func (entity *SoftwarePkgBasicInfo) RejectBy(user *SoftwarePkgApprover) error {
	if !entity.Phase.IsReviewing() {
		return errors.New("can't do this")
	}

	entity.RejectedBy = append(entity.RejectedBy, *user)

	entity.Phase = dp.PackagePhaseClosed

	return nil
}

func (entity *SoftwarePkgBasicInfo) Abandon(user *User) error {
	if !entity.Phase.IsReviewing() {
		return errors.New("can't do this")
	}

	if !dp.IsSameAccount(user.Account, entity.Importer.Account) {
		return errorNotTheImporter
	}

	entity.Phase = dp.PackagePhaseClosed

	return nil
}

func (entity *SoftwarePkgBasicInfo) RerunCI(user *User) (bool, error) {
	if !entity.Phase.IsReviewing() {
		return false, errors.New("can't do this")
	}

	if entity.CI.Status.IsCIRunning() {
		return false, errorCIIsRunning
	}

	if !dp.IsSameAccount(user.Importer.Account, entity.Importer.Account) {
		return false, errorNotTheImporter
	}

	if entity.CI.Status.IsCIWaiting() {
		return false, nil
	}

	entity.CI = SoftwarePkgCI{Status: dp.PackageCIStatusWaiting}

	return true, nil
}

func (entity *SoftwarePkgBasicInfo) UpdateApplication(cmd *SoftwarePkgApplication, user *User) error {
	if !entity.Phase.IsReviewing() {
		return errors.New("can't do this")
	}

	if !dp.IsSameAccount(user.Account, entity.Importer.Account) {
		return errorNotTheImporter
	}

	entity.Application = *cmd

	return nil
}

func (entity *SoftwarePkgBasicInfo) HandleCIChecking() error {
	b := entity.Phase.IsReviewing() && entity.CI.Status.IsCIWaiting()
	if !b {
		return errors.New("can't do this")
	}

	entity.CI.Status = dp.PackageCIStatusRunning
	//entity.CI.startTime = utils.Now()

	return nil
}

func (entity *SoftwarePkgBasicInfo) HandleCIChecked(success bool, prNum int) error {
	if !entity.Phase.IsReviewing() || entity.CI.PRNum != prNum {
		return errors.New("can't do this")
	}

	if success {
		entity.CI.PRNum = prNum
		entity.CI.Status = dp.PackageCIStatusPassed
	} else {
		entity.CI.Status = dp.PackageCIStatusFailed
	}

	return nil
}

func (entity *SoftwarePkgBasicInfo) HandlePkgInitialized(pr dp.URL) error {
	if !entity.Phase.IsCreatingRepo() {
		return errors.New("can't do this")
	}

	entity.RelevantPR = pr

	return nil
}

func (entity *SoftwarePkgBasicInfo) HandlePkgAlreadyExisted() error {
	if !entity.Phase.IsCreatingRepo() {
		return errors.New("can't do this")
	}

	entity.Phase = dp.PackagePhaseClosed

	return nil
}

type RepoCreatedInfo struct {
	Platform dp.PackagePlatform
	RepoLink dp.URL
}

func (entity *SoftwarePkgBasicInfo) HandleRepoCreated(info RepoCreatedInfo) error {
	if !entity.Phase.IsCreatingRepo() {
		return errors.New("can't do this")
	}

	if !dp.IsSamePlatform(entity.Application.PackagePlatform, info.Platform) {
		return errors.New("ignore unmached platform")
	}

	entity.RepoLink = info.RepoLink

	return nil
}

func (entity *SoftwarePkgBasicInfo) HandleCodeSaved(info RepoCreatedInfo) error {
	if err := entity.HandleRepoCreated(info); err != nil {
		return err
	}

	entity.Phase = dp.PackagePhaseImported

	return nil
}

// SoftwarePkg
type SoftwarePkg struct {
	SoftwarePkgBasicInfo

	Logs     []SoftwarePkgOperationLog
	Comments []SoftwarePkgReviewComment
}

func NewSoftwarePkg(user *User, name dp.PackageName, app *SoftwarePkgApplication) SoftwarePkgBasicInfo {
	return SoftwarePkgBasicInfo{
		PkgName:     name,
		Importer:    user.Importer,
		Phase:       dp.PackagePhaseReviewing,
		CI:          SoftwarePkgCI{Status: dp.PackageCIStatusWaiting},
		Application: *app,
		AppliedAt:   utils.Now(),
	}
}
