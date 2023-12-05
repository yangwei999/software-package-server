package pullrequestimpl

import (
	"bytes"
	"io/ioutil"
	"text/template"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
)

type sigInfoTplData struct {
	PkgName    string
	Committers []committer
}

type committer struct {
	OpeneulerId string
	Name        string
	Email       string
}

type repoYamlTplData struct {
	PkgName     string
	PkgDesc     string
	Upstream    string
	Platform    string
	BranchName  string
	ProtectType string
	PublicType  string
}

type prBodyTplData struct {
	PkgName string
	PkgLink string
}

type reviewDetailTplData struct {
	Reviewer   string
	CheckItems []*checkItem
}

type checkItem struct {
	Id      string
	Name    string
	Desc    string
	Result  string
	Comment string
}

func newTemplateImpl(cfg *templateConfig) (templateImpl, error) {
	r := templateImpl{}

	// pr body
	tmpl, err := template.ParseFiles(cfg.PRBodyTpl)
	if err != nil {
		return r, err
	}
	r.prBodyTpl = tmpl

	// repo yaml
	tmpl, err = template.ParseFiles(cfg.RepoYamlTpl)
	if err != nil {
		return r, err
	}
	r.repoYamlTpl = tmpl

	// sig info
	tmpl, err = template.ParseFiles(cfg.SigInfoTpl)
	if err != nil {
		return r, err
	}
	r.sigInfoTpl = tmpl

	// check items
	tmpl, err = template.ParseFiles(cfg.CheckItemsTpl)
	if err != nil {
		return r, err
	}
	r.checkItemsTpl = tmpl

	// review detail
	tmpl, err = template.ParseFiles(cfg.ReviewDetailTpl)
	if err != nil {
		return r, nil
	}
	r.reviewDetailTpl = tmpl

	return r, nil
}

type templateImpl struct {
	prBodyTpl       *template.Template
	sigInfoTpl      *template.Template
	repoYamlTpl     *template.Template
	checkItemsTpl   *template.Template
	reviewDetailTpl *template.Template
}

func (impl *templateImpl) genPRBody(data *prBodyTplData) (string, error) {
	return impl.gen(impl.prBodyTpl, data)
}

func (impl *templateImpl) genSigInfo(data *sigInfoTplData) (string, error) {
	return impl.gen(impl.sigInfoTpl, data)
}

func (impl *templateImpl) genRepoYaml(data *repoYamlTplData, f string) error {
	buf := new(bytes.Buffer)

	if err := impl.repoYamlTpl.Execute(buf, data); err != nil {
		return err
	}

	return ioutil.WriteFile(f, buf.Bytes(), 0644)
}

func (impl *templateImpl) genCheckItems(data *domain.Config) (string, error) {
	return impl.gen(impl.checkItemsTpl, data)
}

func (impl *templateImpl) genReviewDetail(data *reviewDetailTplData) (string, error) {
	return impl.gen(impl.reviewDetailTpl, data)
}

func (impl *templateImpl) gen(tpl *template.Template, data interface{}) (string, error) {
	buf := new(bytes.Buffer)

	if err := tpl.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
