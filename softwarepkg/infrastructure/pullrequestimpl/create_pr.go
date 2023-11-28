package pullrequestimpl

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/opensourceways/server-common-lib/utils"
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/robot-gitee-software-package/softwarepkg/domain"
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
		impl.branchName(pkg.Name),
		fmt.Sprintf("sig/%s/sig-info.yaml", pkg.Application.ImportingPkgSig),
		sigInfoData,
		fmt.Sprintf(
			"sig/%s/src-openeuler/%s/%s.yaml",
			pkg.Application.ImportingPkgSig,
			strings.ToLower(pkg.Name[:1]),
			pkg.Name,
		),
		repoFile,
	}

	out, err, _ := utils.RunCmd(params...)
	if err != nil {
		logrus.Errorf(
			"run create pr shell, err=%s, out=%s, params=%v",
			err.Error(), string(out), params[:len(params)-1],
		)
	}

	return err
}

func (impl *pullRequestImpl) branchName(pkgName string) string {
	return fmt.Sprintf("software_package_%s", pkgName)
}

func (impl *pullRequestImpl) genAppendSigInfoData(pkg *domain.SoftwarePkg) (string, error) {
	return impl.template.genSigInfo(&sigInfoTplData{
		PkgName:       pkg.Name,
		ImporterEmail: pkg.Importer.Email,
		Importer:      pkg.Importer.Name,
	})

}

func (impl *pullRequestImpl) genNewRepoFile(pkg *domain.SoftwarePkg) (string, error) {
	f := filepath.Join(
		impl.cfg.ShellScript.WorkDir,
		fmt.Sprintf("%s_%s", impl.branchName(pkg.Name), pkg.Name),
	)

	err := impl.template.genRepoYaml(&repoYamlTplData{
		PkgName:     pkg.Name,
		PkgDesc:     fmt.Sprintf("'%s'", pkg.Application.PackageDesc),
		Upstream:    pkg.Application.Upstream,
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

func (impl *pullRequestImpl) createPR(pkg *domain.SoftwarePkg) (pr domain.PullRequest, err error) {
	body, err := impl.template.genPRBody(&prBodyTplData{
		PkgName: pkg.Name,
		PkgLink: impl.cfg.SoftwarePkg.Endpoint + pkg.Id,
	})
	if err != nil {
		return
	}

	v, err := impl.cli.CreatePullRequest(
		impl.cfg.CommunityRobot.Org, impl.cfg.CommunityRobot.Repo,
		fmt.Sprintf("add eco-package: %s", pkg.Name),
		body,
		fmt.Sprintf(
			"%s:%s", impl.cfg.Robot.Username, impl.branchName(pkg.Name),
		),
		"master", true,
	)
	if err == nil {
		pr.Num = int(v.Number)
		pr.Link = v.HtmlUrl
	}

	return
}
