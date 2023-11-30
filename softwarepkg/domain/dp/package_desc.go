package dp

type PackageDesc interface {
	PackageDesc() string
}

func NewPackageDesc(v string) (PackageDesc, error) {

	return packageDesc(v), nil
}

type packageDesc string

func (v packageDesc) PackageDesc() string {
	return string(v)
}
