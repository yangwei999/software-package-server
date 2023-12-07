package pullrequestimpl

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/opensourceways/server-common-lib/utils"
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	watchdomain "github.com/opensourceways/software-package-server/watch/domain"
)

func (impl *pullRequestImpl) createBranch(pkg *domain.SoftwarePkg) error {
	sigInfoData, err := impl.genAppendSigInfoData(pkg)
	if err != nil {
		return err
	}

	repoFile, err := impl.genNewRepoFile(pkg)
	if err != nil {
		return err
	}

	cfg := &impl.cfg
	params := []string{
		cfg.ShellScript.BranchScript,
		impl.localRepoDir,
		impl.branchName(pkg.Basic.Name.PackageName()),
		fmt.Sprintf("sig/%s/sig-info.yaml", pkg.Sig.ImportingPkgSig()),
		sigInfoData,
		fmt.Sprintf(
			"sig/%s/src-openeuler/%s/%s.yaml",
			pkg.Sig.ImportingPkgSig(),
			strings.ToLower(pkg.Basic.Name.PackageName()[:1]),
			pkg.Basic.Name.PackageName(),
		),
		repoFile,
	}

	out, err, _ := utils.RunCmd(params...)
	if err != nil {
		logrus.Errorf(
			"run create pr shell, err=%s, out=%s, params=%v",
			err.Error(), string(out), params,
		)
	}

	return err
}

func (impl *pullRequestImpl) branchName(pkgName string) string {
	return fmt.Sprintf("software_package_%s", pkgName)
}

func (impl *pullRequestImpl) genAppendSigInfoData(pkg *domain.SoftwarePkg) (string, error) {
	data := sigInfoTplData{
		PkgName: pkg.Basic.Name.PackageName(),
	}

	for _, v := range pkg.Repo.Committers {
		user, err := impl.ua.Find(v.Account.Account(), pkg.Repo.Platform.PackagePlatform())
		if err != nil {
			logrus.Errorf("get email of %s %s error:%s",
				v.Account.Account(),
				pkg.Repo.Platform.PackagePlatform(),
				err.Error(),
			)
			continue
		}

		data.Committers = append(data.Committers, committer{
			OpeneulerId: v.Account.Account(),
			Name:        v.Account.Account(),
			Email:       user.Email.Email(),
		})
	}

	return impl.template.genSigInfo(&data)
}

func (impl *pullRequestImpl) genNewRepoFile(pkg *domain.SoftwarePkg) (string, error) {
	pkgName := pkg.Basic.Name.PackageName()
	f := filepath.Join(
		impl.cfg.ShellScript.WorkDir,
		fmt.Sprintf("%s_%s", impl.branchName(pkgName), pkgName),
	)

	err := impl.template.genRepoYaml(&repoYamlTplData{
		PkgName:     pkgName,
		PkgDesc:     fmt.Sprintf("'%s'", pkg.Basic.Desc.PackageDesc()),
		Upstream:    pkg.Basic.Upstream.URL(),
		Platform:    pkg.Repo.Platform.PackagePlatform(),
		BranchName:  impl.cfg.Robot.NewRepoBranch.Name,
		ProtectType: impl.cfg.Robot.NewRepoBranch.ProtectType,
		PublicType:  impl.cfg.Robot.NewRepoBranch.PublicType,
	}, f)

	return f, err
}

func (impl *pullRequestImpl) genTemplate(fileName string, data interface{}) (string, error) {
	tmpl, err := template.ParseFiles(fileName)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (impl *pullRequestImpl) createPR(pkg *domain.SoftwarePkg) (pr watchdomain.PullRequest, err error) {
	pkgName := pkg.Basic.Name.PackageName()

	body, err := impl.template.genPRBody(&prBodyTplData{
		PkgName: pkgName,
		PkgLink: impl.cfg.SoftwarePkg.Endpoint + pkg.Id,
	})
	if err != nil {
		return
	}

	v, err := impl.cli.CreatePullRequest(
		impl.cfg.CommunityRobot.Org, impl.cfg.CommunityRobot.Repo,
		fmt.Sprintf("add eco-package: %s", pkgName),
		body,
		fmt.Sprintf(
			"%s:%s", impl.cfg.Robot.Username, impl.branchName(pkgName),
		),
		"master", true,
	)
	if err == nil {
		pr.Num = int(v.Number)
		pr.Link = v.HtmlUrl
	}

	return
}
