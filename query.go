package queryassistor

func subToQuery(sb *SubQuery) *Query {
	q := &Query{
		Table:     sb.Table,
		Id:        sb.Id,
		Filter:    sb.Filter,
		Select:    sb.Select,
		Group:     sb.Group,
		Having:    sb.Having,
		Joins:     sb.Joins,
		Limit:     sb.Limit,
		Order:     sb.Order,
		CountArgs: sb.CountArgs,
	}
	if len(q.Id) == 0 {
		q.Id = q.Table
	}
	return q
}
