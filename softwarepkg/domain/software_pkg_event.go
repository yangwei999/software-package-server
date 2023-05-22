package domain

import (
	"encoding/json"

	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
)

var (
	NewSoftwarePkgAppUpdatedEvent  = NewSoftwarePkgAppliedEvent
	NewSoftwarePkgInitializedEvent = NewSoftwarePkgApprovedEvent
)

// softwarePkgAppliedEvent
type softwarePkgAppliedEvent struct {
	PkgId string `json:"pkg_id"`
}

func (e *softwarePkgAppliedEvent) Message() ([]byte, error) {
	return json.Marshal(e)
}

func NewSoftwarePkgAppliedEvent(pkg *SoftwarePkgBasicInfo) softwarePkgAppliedEvent {
	return softwarePkgAppliedEvent{
		PkgId: pkg.Id,
	}
}

func UnmarshalToSoftwarePkgAppliedEvent(data []byte) (e softwarePkgAppliedEvent, err error) {
	err = json.Unmarshal(data, &e)

	return
}

// softwarePkgApprovedEvent
type softwarePkgApprovedEvent struct {
	Importer          string `json:"importer"`
	ImporterEmail     string `json:"importer_email"`
	PkgId             string `json:"pkg_id"`
	PkgName           string `json:"pkg_name"`
	PkgDesc           string `json:"pkg_desc"`
	Upstream          string `json:"upstream"`
	SpecURL           string `json:"spec_url"`
	SrcRPMURL         string `json:"src_rpm_url"`
	Platform          string `json:"platform"`
	ImportingPkgSig   string `json:"sig"`
	ReasonToImportPkg string `json:"reason_to_import"`
	CIPRNum           int    `json:"ci_pr_num"`
}

func (e *softwarePkgApprovedEvent) Message() ([]byte, error) {
	return json.Marshal(e)
}

func NewSoftwarePkgApprovedEvent(pkg *SoftwarePkgBasicInfo) softwarePkgApprovedEvent {
	app := &pkg.Application

	return softwarePkgApprovedEvent{
		Importer:          pkg.Importer.Account.Account(),
		ImporterEmail:     pkg.Importer.Email.Email(),
		PkgId:             pkg.Id,
		PkgName:           pkg.PkgName.PackageName(),
		PkgDesc:           app.PackageDesc.PackageDesc(),
		Upstream:          app.Upstream.URL(),
		SpecURL:           app.SourceCode.SpecURL.URL(),
		SrcRPMURL:         app.SourceCode.SrcRPMURL.URL(),
		CIPRNum:           pkg.CI.PRNum,
		Platform:          app.PackagePlatform.PackagePlatform(),
		ImportingPkgSig:   app.ImportingPkgSig.ImportingPkgSig(),
		ReasonToImportPkg: app.ReasonToImportPkg.ReasonToImportPkg(),
	}
}

// softwarePkgAlreadyExistedEvent
type softwarePkgAlreadyExistedEvent struct {
	PkgName string `json:"pkg_name"`
}

func (e *softwarePkgAlreadyExistedEvent) Message() ([]byte, error) {
	return json.Marshal(e)
}

func NewSoftwarePkgAlreadyExistEvent(pkg dp.PackageName) softwarePkgAlreadyExistedEvent {
	return softwarePkgAlreadyExistedEvent{
		PkgName: pkg.PackageName(),
	}
}

func UnmarshalToSoftwarePkgAlreadyExistEvent(data []byte) (
	e softwarePkgAlreadyExistedEvent, err error,
) {
	err = json.Unmarshal(data, &e)

	return
}
