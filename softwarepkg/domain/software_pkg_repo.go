package domain

import "github.com/opensourceways/software-package-server/softwarepkg/domain/dp"

type PkgCommitter struct {
	Account    dp.Account // openeuler Id
	PlatformId string     // gitee id or github id which depends on the platform of pkg repo
}

// SoftwarePkgRepo
type SoftwarePkgRepo struct {
	Platform   dp.PackagePlatform
	Committers []PkgCommitter
}

func (r *SoftwarePkgRepo) platform() string {
	return r.Platform.PackagePlatform()
}

func (r *SoftwarePkgRepo) repoLink(name dp.PackageName) string {
	return r.Platform.RepoLink(name)
}

func (r *SoftwarePkgRepo) isCommitter(u dp.Account) bool {
	for i := range r.Committers {
		if dp.IsSameAccount(r.Committers[i].Account, u) {
			return true
		}
	}

	return false
}

func (r *SoftwarePkgRepo) update(r1 *SoftwarePkgRepo) string {
	if v := r1.Platform; v != nil && v.PackagePlatform() != r.Platform.PackagePlatform() {
		r.Platform = v
	}

	if len(r1.Committers) > 0 && !r.isSameCommitters(r1) {
		r.Committers = r1.Committers

		return pkgModificationCommitter
	}

	return ""
}

func (r *SoftwarePkgRepo) isSameCommitters(r1 *SoftwarePkgRepo) bool {
	if len(r.Committers) != len(r1.Committers) {
		return false
	}

	m := r.committersMap()

	for i := range r1.Committers {
		if !m[r1.Committers[i].Account.Account()] {
			return false
		}
	}

	return true
}

func (r *SoftwarePkgRepo) committersMap() map[string]bool {
	m := map[string]bool{}

	for i := range r.Committers {
		m[r.Committers[i].Account.Account()] = true
	}

	return m
}

func (r *SoftwarePkgRepo) CommitterIds() []string {
	v := make([]string, len(r.Committers))

	for i := range r.Committers {
		v[i] = r.Committers[i].PlatformId
	}

	return v
}
