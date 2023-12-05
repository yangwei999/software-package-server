package repositoryimpl

import "github.com/opensourceways/software-package-server/common/infrastructure/postgresql"

type dbClient interface {
	Insert(filter, result interface{}) error
	InsertWithNot(filter, notFilter, result interface{}) error
	Count([]postgresql.ColumnFilter) (int, error)
	GetRecords([]postgresql.ColumnFilter, interface{}, postgresql.Pagination, []postgresql.SortByColumn) error
	GetRecord(filter, result interface{}) error
	UpdateRecord(filter, update interface{}) error

	IsRowNotFound(error) bool
	IsRowExists(error) bool
}
