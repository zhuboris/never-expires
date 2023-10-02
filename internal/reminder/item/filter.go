package item

import (
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Filter func() (sql string, param any)

func ByDateBefore(date time.Time) Filter {
	return func() (string, any) {
		return `AND expiration_date <= $`, date
	}
}

func ByStorageID(id pgtype.UUID) Filter {
	return func() (string, any) {
		return `AND i.storage_id = $`, id
	}
}

func ByName(name string) Filter {
	return func() (string, any) {
		return `AND ii.name ILIKE $`, name
	}
}

func ByOpenedStatus(isOpened bool) Filter {
	return func() (string, any) {
		return `AND ii.is_opened = $`, isOpened
	}
}

func makeQueryFilteringPath(nextParamIndex int, filters []Filter) (sql string, params []any) {
	var sqlBuilder strings.Builder
	params = make([]any, 0, len(filters))

	for _, filter := range filters {
		sql, param := filter()
		sqlBuilder.WriteString(sql)
		sqlBuilder.WriteString(strconv.Itoa(nextParamIndex))
		params = append(params, param)
		nextParamIndex++
	}

	return sqlBuilder.String(), params
}
