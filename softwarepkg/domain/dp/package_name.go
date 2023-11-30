package dp

type PackageName interface {
	PackageName() string
}

func NewPackageName(v string) (PackageName, error) {

	return packageName(v), nil
}

type packageName string

func (v packageName) PackageName() string {
	return string(v)
}
