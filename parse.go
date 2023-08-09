package queryassistor

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

func parseTag(qb *queryBox) *Query {
	q := qb.Query
	req := qb.req
	t := reflect.TypeOf(req).Elem()
	va := reflect.ValueOf(req).Elem()
	var fields []string
	for i := 0; i < t.NumField(); i++ {
		fieldType := t.Field(i)
		fields = append(fields, fieldType.Name)
		tagColumn := fieldType.Tag.Get("query")
		if len(tagColumn) > 0 {
			arr := strings.Split(tagColumn, ";")
			for _, f := range arr {
				var cmd string
				leftBracket := strings.Index(f, "(")
				rightBracket := strings.Index(f, ")")
				var params []string
				if leftBracket != -1 && rightBracket != -1 {
					cmd = f[:leftBracket]
					params = strings.Split(f[leftBracket+1:rightBracket], ",")
					for i, _ := range params {
						params[i] = strings.TrimSpace(params[i])
					}
				} else {
					cmd = f
				}
				if _, ok := commands[cmd]; ok {
					commands[cmd](qb, params, va.Field(i))
				}
			}
		}
	}
	return q
}

func commandFilter(q *queryBox, params []string, v reflect.Value) {
	var column string
	var operator string

	if len(params) >= 1 {
		column = params[0]
	}
	if len(params) >= 2 {
		operator = params[1]
	}
	if len(params) >= 3 {
		if params[2] == "omitempty" && isEmpty(v) {
			return
		}
	}
	qName := "t"
	if len(params) >= 4 {
		qName = params[3]
	}
	q.Filters[qName] = append(q.Filters[qName], FilterItem{
		Field:    column,
		Operator: operator,
		Value:    v.Interface(),
	})
}

func commandTimeFilter(q *queryBox, params []string, v reflect.Value) {
	var column string
	var operator string
	value := v.Interface()

	if len(v.String()) > 22 {
		tm, _ := time.Parse(time.RFC3339, v.String())
		value = tm.In(time.Local).Format("2006-01-02 15:04:05")
	}

	if len(params) >= 1 {
		column = params[0]
	}
	if len(params) >= 2 {
		operator = params[1]
	}
	if len(params) >= 3 {
		if params[2] == "omitempty" && isEmpty(v) {
			return
		}
	}
	if len(params) >= 4 {
		if params[3] == "end" {
			dateStr := v.String()
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				return
			}
			value = date.Add(86399 * time.Second).Format("2006-01-02 15:04:05")
		}
		if params[3] == "start" {
			dateStr := v.String()
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				return
			}
			value = date.Format("2006-01-02 15:04:05")
		}
	}
	if len(params) >= 5 {
		if params[4] == "end" {
			dateStr := v.String()
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				return
			}
			value = date.Add(86399 * time.Second).Format("2006-01-02 15:04:05")
		}
		if params[4] == "start" {
			dateStr := v.String()
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				return
			}
			value = date.Format("2006-01-02 15:04:05")
		}
	}

	qName := "t"
	if len(params) >= 5 {
		qName = params[4]
	}
	q.Filters[qName] = append(q.Filters[qName], FilterItem{
		Field:    column,
		Operator: operator,
		Value:    value,
	})
}

func commandDateFilter(q *queryBox, params []string, v reflect.Value) {
	var column string
	var operator string
	value := v.Interface()

	if len(params) >= 1 {
		column = params[0]
	}
	if len(params) >= 2 {
		operator = params[1]
	}
	if len(params) >= 3 {
		if isEmpty(v) {
			return
		}
	}
	if len(params) >= 4 {
		if params[3] == "end" {
			dateStr := v.String()
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				return
			}
			value = date.Add(86399 * time.Second).Format("2006-01-02")
		}
		if params[3] == "start" {
			dateStr := v.String()
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				return
			}
			value = date.Format("2006-01-02")
		}
	}

	qName := "t"
	if len(params) >= 5 {
		qName = params[4]
	}
	q.Filters[qName] = append(q.Filters[qName], FilterItem{
		Field:    column,
		Operator: operator,
		Value:    value,
	})
}

func commandHaving(q *queryBox, params []string, v reflect.Value) {
	var column string
	var operator string

	if len(params) >= 1 {
		column = params[0]
	}
	if len(params) >= 2 {
		operator = params[1]
	}
	if len(params) >= 3 {
		if isEmpty(v) {
			return
		}
	}

	qName := "t"
	if len(params) >= 4 {
		qName = params[3]
	}
	q.Having[qName] = append(q.Having[qName], FilterItem{
		Field:    column,
		Operator: operator,
		Value:    v.Interface(),
	})
}

func commandPage(req *queryBox, params []string, v reflect.Value) {
	page := v.Int()

	if page > 0 {
		req.Query.Page = int(page)
	} else {
		req.Query.Page = 1
	}
}

func commandLimit(req *queryBox, params []string, v reflect.Value) {
	limit := int(v.Int())

	if limit > 0 {
		req.Query.Limit = limit
		return
	}
	req.Query.Limit = 1
}

func commandPageSize(req *queryBox, params []string, v reflect.Value) {
	pageSize := int(v.Int())

	if pageSize > 0 {
		req.Query.Limit = pageSize
		return
	}

	if len(params) > 0 {
		pageSize, _ = strconv.Atoi(params[0])
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	req.Query.Limit = pageSize
}

func commandOrder(req *queryBox, params []string, v reflect.Value) {
	order := v.String()
	if len(order) > 0 {
		req.Query.Order = order
	}
}

func isEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int:
		return v.Int() == 0
	case reflect.Float64:
		return v.Float() == 0
	case reflect.Float32:
		return v.Float() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Slice:
		return v.Len() == 0
	case reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func generateFmtPlaceholder(v interface{}) string {
	ref := reflect.TypeOf(v)
	switch ref.Kind() {
	case reflect.Int, reflect.Uint, reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16, reflect.Int32, reflect.Uint32, reflect.Int64, reflect.Uint64:
		return "%d"
	case reflect.Float64, reflect.Float32:
		return "%f"
	}
	return "'%s'"
}

func commandNullFilter(q *queryBox, params []string, v reflect.Value) {
	var column string
	var operator string
	value := v.Interface()

	if len(params) >= 1 {
		column = params[0]
	}
	if len(params) >= 2 {
		operator = params[1]
	}
	if len(params) >= 3 {
		if isEmpty(v) {
			return
		}
	}
	if len(params) >= 4 {
		if params[3] == "-1" {
			dateStr := v.String()
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				return
			}
			value = date.Format("2006-01-02 15:04:05")
		}
	}

	qName := "t"
	if len(params) >= 4 {
		qName = params[3]
	}
	q.Filters[qName] = append(q.Filters[qName], FilterItem{
		Field:    column,
		Operator: operator,
		Value:    value,
	})
}
