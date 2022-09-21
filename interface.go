package queryassistor

func Row(query IReqQuery) (map[string]interface{}, error) {
	qb := newQueryBox(query)
	row, err := iml.Row(qb)
	if err != nil {
		return nil, err
	}

	for _, mq := range qb.Query.MergeQueries {
		qb.BindValues = nil
		sql := qb.toSql(&mq)
		mRow, err := iml.row(&mq, sql, qb.BindValues)
		if err != nil {
			return nil, err
		}
		for k, v := range mRow {
			row[k] = v
		}
	}
	return row, nil
}

func Rows(query IReqQuery) ([]map[string]interface{}, error) {
	qb := newQueryBox(query)
	return iml.Rows(qb)
}

func Page(query IReqQuery) ([]map[string]interface{}, int64, error) {
	qb := newQueryBox(query)
	return iml.Page(qb)
}

func Export(query IReqQuery) (string, error) {
	qb := newQueryBox(query)
	return iml.Export(qb)
}

func ExportMultiQuery(exportConfig *ExportMultiQueryConfig) (string, error) {
	return iml.ExportMultiQuery(exportConfig)
}
