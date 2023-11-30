package dp

type PurposeToImportPkg interface {
	PurposeToImportPkg() string
}

func NewPurposeToImportPkg(v string) (PurposeToImportPkg, error) {

	return purposeToImportPkg(v), nil
}

type purposeToImportPkg string

func (v purposeToImportPkg) PurposeToImportPkg() string {
	return string(v)
}
