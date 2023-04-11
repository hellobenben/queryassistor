package queryassistor

import (
	"fmt"
	"reflect"
	"strings"
)

func (rq *queryBox) ToSql() string {
	rq.BindValues = []interface{}{}
	if rq.totalModel {
		oldSelect := rq.Query.Select
		oldOrder := rq.Query.Order
		oldPage := rq.Query.Page
		oldLimit := rq.Query.Limit
		//column := "COUNT(*)"
		//if len(rq.Query.CountArgs) > 0 {
		//	column = rq.Query.CountArgs
		//}
		//rq.Query.Select = []SelectField{{
		//	Column: column,
		//	Alias:  "total",
		//}}
		rq.Query.Limit = 0
		rq.Query.Order = ""
		rq.Query.Page = 1

		sql := rq.toSql(rq.Query)
		//if rq.totalModel {
		sql = fmt.Sprintf("SELECT COUNT(*) AS `total` FROM (%s) t LIMIT 1", sql)
		//}
		rq.totalModel = false
		rq.Query.Select = oldSelect
		rq.Query.Order = oldOrder
		rq.Query.Page = oldPage
		rq.Query.Limit = oldLimit
		return sql
	}
	sql := rq.toSql(rq.Query)
	return sql
}

func (rq *queryBox) toSql(q *Query) string {
	selectClause := fmt.Sprintf("SELECT %s", rq.parseSelectClause(q))
	formClause := ""
	if q.Table != "" {
		formClause = fmt.Sprintf(" FROM %s", q.Table)
	} else if q.Table == "" {
		sq := subToQuery(q.SubQuery)
		formClause = fmt.Sprintf(" FROM (%s) a", rq.toSql(sq))
	}

	joinClause := rq.parseJoin(q)
	whereClause := rq.parseWhere(q)
	groupClause := rq.parseGroupBy(q)
	havingClause := rq.parseHaving(q)
	orderClause := rq.parseOrderBy(q)
	limitClause := rq.parseLimit(q)
	offsetClause := rq.parseOffset(q)

	sql := fmt.Sprintf("%s%s%s%s%s%s%s%s%s",
		selectClause,
		formClause,
		joinClause,
		whereClause,
		groupClause,
		havingClause,
		orderClause,
		limitClause,
		offsetClause)
	return sql
}

func (rq *queryBox) mergeSql(query *Query) []string {
	var sqlArr []string
	for _, q := range rq.Query.MergeQueries {
		sqlArr = append(sqlArr, rq.toSql(&q))
	}
	return sqlArr
}

func (rq *queryBox) parseSelectClause(query *Query) string {
	var columns []string
	for _, f := range query.Select {
		if len(f.Alias) > 0 {
			columns = append(columns, fmt.Sprintf("%s AS `%s`", f.Column, f.Alias))
		} else {
			columns = append(columns, fmt.Sprintf("%s", f.Column))
		}
	}
	if len(columns) > 0 {
		return strings.Join(columns, ",")
	}
	return ""
}

func (rq *queryBox) parseGroupBy(query *Query) string {
	groupClause := ""
	if len(query.Group) > 0 {
		groupClause = " GROUP BY " + query.Group
	}
	return groupClause
}

func (rq *queryBox) parseOrderBy(query *Query) string {
	orderClause := ""
	if len(query.Order) > 0 {
		orderClause = " ORDER BY " + query.Order
	}
	return orderClause
}

func (rq *queryBox) parseLimit(query *Query) string {
	limitClause := ""
	if query.Limit > 0 {
		limitClause = fmt.Sprintf(" LIMIT %d", query.Limit)
	}
	return limitClause
}

func (rq *queryBox) parseOffset(query *Query) string {
	clause := ""
	page := query.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * query.Limit
	if offset > 0 {
		clause = fmt.Sprintf(" OFFSET %d", offset)
	}
	return clause
}

func (rq *queryBox) parseJoin(query *Query) string {
	var joins []string
	for _, j := range query.Joins {
		joinType := "JOIN"
		joinClause := ""
		switch strings.ToUpper(j.Type) {
		case "LEFT":
			joinType = "LEFT JOIN"
		case "RIGHT":
			joinType = "LEFT JOIN"
		case "INNER":
			joinType = "INNER JOIN"
		case "CROSS":
			joinType = "CROSS JOIN"
		case "FULL OUTER":
			joinType = "FULL OUTER JOIN"
		case "":
			joinType = ""
		}
		if j.SubQuery != nil {
			q := subToQuery(j.SubQuery)
			subSql := rq.toSql(q)
			joinClause = fmt.Sprintf(" %s (%s) AS %s", joinType, subSql, q.Id)
		} else {
			joinClause = fmt.Sprintf(" %s %s", joinType, j.Table)
		}
		if len(j.On) > 0 {
			joinClause = fmt.Sprintf("%s ON %s", joinClause, j.On)
		}
		joins = append(joins, joinClause)
	}
	return strings.Join(joins, " ")
}

func (rq *queryBox) parseWhere(query *Query) string {
	filters := query.Filter
	if f, ok := rq.Filters[query.GetId()]; ok {
		filters = append(filters, f...)
	}
	var where []string
	for _, f := range filters {
		where = append(where, rq.parseFilter(&f))
	}
	if len(where) == 0 {
		return ""
	}
	return fmt.Sprintf(" WHERE %s", strings.Join(where, " AND "))
}

func (rq *queryBox) parseHaving(query *Query) string {
	var having []string
	for _, h := range query.Having {
		having = append(having, rq.parseFilter(&h))
	}
	if len(having) == 0 {
		return ""
	}
	return fmt.Sprintf(" HAVING %s", strings.Join(having, " AND "))
}

func (rq *queryBox) parseFilter(f *FilterItem) string {
	operator := ""
	switch f.Operator {
	case "eq":
		operator = "="
	case "neq":
		operator = "!="
	case "lte":
		operator = "<="
	case "gte":
		operator = ">="
	case "not in":
		operator = "NOT IN"
	case "in":
		operator = "IN"
	case "gt":
		operator = ">"
	case "lt":
		operator = "<"
	case "or":
		operator = "or"
	case "like":
		operator = "like"
	case "is null":
		return fmt.Sprintf("%s IS NULL", f.Field)
	case "is not null":
		return fmt.Sprintf("%s IS NOT NULL", f.Field)
	case "":
		return fmt.Sprintf("%s", f.Field)
	}

	if len(operator) == 0 {
		return ""
	}

	kind := reflect.TypeOf(f.Value).Kind()
	switch kind {
	case reflect.Ptr:
		sq, ok := reflect.ValueOf(f.Value).Interface().(*SubQuery)
		if !ok {
			break
		}
		sub := subToQuery(sq)
		sql := rq.toSql(sub)
		return fmt.Sprintf("%s %s (%s)", f.Field, operator, sql)
	}

	if operator == "IN" || operator == "NOT IN" {
		var value []string
		switch kind {
		case reflect.String:
			value = strings.Split(f.Value.(string), ",")
		case reflect.Slice:
			value = f.Value.([]string)
		case reflect.Array:
			value = f.Value.([]string)
		}
		var str string
		for i, s := range value {
			if i > 0 {
				str += ","
			}
			str += fmt.Sprintf("'%s'", strings.Replace(s, "'", "", -1))
		}
		return fmt.Sprintf("%s %s (%s)", f.Field, operator, str)
	} else {
		rq.BindValues = append(rq.BindValues, f.Value)
	}

	return fmt.Sprintf("%s %s ?", f.Field, operator)
}
