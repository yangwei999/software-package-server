package dp

import "errors"

const (
	Gitee  = "gitee"
	Github = "github"
)

type PackagePlatform interface {
	PackagePlatform() string
	RepoLink(name PackageName) string
}

func NewPackagePlatformByRepoLink(repoLink string) (PackagePlatform, error) {
	if v := config.platformOfRepoLink(repoLink); v != "" {
		return packagePlatform(v), nil
	}

	return nil, errors.New("invalid org link")
}

func NewPackagePlatform(v string) (PackagePlatform, error) {

	return packagePlatform(v), nil
}

type packagePlatform string

func (v packagePlatform) PackagePlatform() string {
	return string(v)
}

func (v packagePlatform) RepoLink(name PackageName) string {
	return config.orgLinkOfPlatform(string(v)) + name.PackageName()
}
