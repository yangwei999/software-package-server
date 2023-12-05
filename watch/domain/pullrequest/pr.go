package pullrequest

import (
	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	watchdomain "github.com/opensourceways/software-package-server/watch/domain"
)

type PullRequest interface {
	Create(*domain.SoftwarePkg) (watchdomain.PullRequest, error)
	Update(*domain.SoftwarePkg) error
	Merge(int) error
	Close(int) error
}
