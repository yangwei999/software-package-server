package domain

import (
	"encoding/json"

	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
)

var (
	NewSoftwarePkgRetestedEvent     = NewSoftwarePkgAppliedEvent
	NewSoftwarePkgCodeUpdatedEvent  = NewSoftwarePkgAppliedEvent
	NewSoftwarePkgCodeChangeedEvent = NewSoftwarePkgAppliedEvent
)

// softwarePkgAppliedEvent
type softwarePkgAppliedEvent struct {
	PkgId string `json:"pkg_id"`
}

func (e *softwarePkgAppliedEvent) Message() ([]byte, error) {
	return json.Marshal(e)
}

func NewSoftwarePkgAppliedEvent(pkg *SoftwarePkg) softwarePkgAppliedEvent {
	return softwarePkgAppliedEvent{
		PkgId: pkg.Id,
	}
}

func UnmarshalToSoftwarePkgAppliedEvent(data []byte) (e softwarePkgAppliedEvent, err error) {
	err = json.Unmarshal(data, &e)

	return
}

// softwarePkgInitializedEvent
type softwarePkgInitializedEvent struct {
	Importer string `json:"importer"`
	PkgId    string `json:"pkg_id"`
	PkgName  string `json:"pkg_name"`
	Platform string `json:"platform"`
}

func (e *softwarePkgInitializedEvent) Message() ([]byte, error) {
	return json.Marshal(e)
}

func NewSoftwarePkgInitializedEvent(pkg *SoftwarePkg) softwarePkgInitializedEvent {
	basic := &pkg.Basic

	return softwarePkgInitializedEvent{
		Importer: pkg.Importer.Account(),
		PkgId:    pkg.Id,
		PkgName:  basic.Name.PackageName(),
		Platform: pkg.Repo.platform(),
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
