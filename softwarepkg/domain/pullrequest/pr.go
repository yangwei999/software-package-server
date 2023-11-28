package pullrequest

import "github.com/opensourceways/software-package-server/softwarepkg/domain"

type PullRequest interface {
	Create(*domain.SoftwarePkg) (domain.PullRequest, error)
	Merge(int) error
	Close(int) error
	Comment(int, string) error
}
