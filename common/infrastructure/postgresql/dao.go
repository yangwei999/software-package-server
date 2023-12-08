package postgresql

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

var (
	errRowExists   = errors.New("row exists")
	errRowNotFound = errors.New("row not found")
)

type SortByColumn struct {
	Column string
	Ascend bool
}

func (s SortByColumn) order() string {
	v := " ASC"
	if !s.Ascend {
		v = " DESC"
	}
	return s.Column + v
}

type Pagination struct {
	PageNum      int
	CountPerPage int
}

func (p Pagination) pagination() (limit, offset int) {
	limit = p.CountPerPage

	if limit > 0 && p.PageNum > 0 {
		offset = (p.PageNum - 1) * limit
	}

	return
}

type ColumnFilter struct {
	column string
	symbol string
	value  interface{}
}

func (q *ColumnFilter) condition() string {
	return fmt.Sprintf("%s %s ?", q.column, q.symbol)
}

func NewEqualFilter(column string, value interface{}) ColumnFilter {
	return ColumnFilter{
		column: column,
		symbol: "=",
		value:  value,
	}
}

func NewLikeFilter(column string, value string) ColumnFilter {
	return ColumnFilter{
		column: column,
		symbol: "ilike",
		value:  "%" + value + "%",
	}
}

type dbTable struct {
	name string
}

func NewDBTable(name string) dbTable {
	return dbTable{name: name}
}

func (t dbTable) DBInstance() *gorm.DB {
	return db.Table(t.name)
}

func (t dbTable) Insert(filter, result interface{}) error {
	query := db.Table(t.name).Where(filter).FirstOrCreate(result)

	if err := query.Error; err != nil {
		return err
	}

	if query.RowsAffected == 0 {
		return errRowExists
	}

	return nil
}

func (t dbTable) InsertWithNot(filter, notFilter, result interface{}) error {
	query := db.Table(t.name).
		Where(filter).
		Not(notFilter).
		FirstOrCreate(result)

	if err := query.Error; err != nil {
		return err
	}

	if query.RowsAffected == 0 {
		return errRowExists
	}

	return nil
}

func (t dbTable) GetRecords(
	filter []ColumnFilter, result interface{}, p Pagination,
	sort []SortByColumn,
) (err error) {
	query := db.Table(t.name)
	for i := range filter {
		query.Where(filter[i].condition(), filter[i].value)
	}

	var orders []string
	for _, v := range sort {
		orders = append(orders, v.order())
	}

	if len(orders) >= 0 {
		query.Order(strings.Join(orders, ","))
	}

	if limit, offset := p.pagination(); limit > 0 {
		query.Limit(limit).Offset(offset)
	}

	err = query.Find(result).Error

	return
}

func (t dbTable) Count(filter []ColumnFilter) (int, error) {
	var total int64
	query := db.Table(t.name)
	for i := range filter {
		query.Where(filter[i].condition(), filter[i].value)
	}

	err := query.Count(&total).Error

	return int(total), err
}

func (t dbTable) GetRecord(filter, result interface{}) error {
	err := db.Table(t.name).Where(filter).First(result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errRowNotFound
	}

	return err
}

func (t dbTable) UpdateRecord(filter, update interface{}) (err error) {
	query := db.Table(t.name).Where(filter).Updates(update)
	if err = query.Error; err != nil {
		return
	}

	if query.RowsAffected == 0 {
		err = errRowNotFound
	}

	return
}

func (t dbTable) IsRowNotFound(err error) bool {
	return errors.Is(err, errRowNotFound)
}

func (t dbTable) IsRowExists(err error) bool {
	return errors.Is(err, errRowExists)
}
