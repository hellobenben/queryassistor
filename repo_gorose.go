package queryassistor

import (
	"github.com/gohouse/gorose/v2"
	"github.com/tealeg/xlsx"
	"os"
	"reflect"
	"time"
)

type goroseIml struct {
	Timezone string
}

func (*goroseIml) row(q *Query, sql string, values []interface{}) (map[string]interface{}, error) {
	var orm gorose.IOrm
	ds := q.Datasource
	if ds == "mysql" {
		orm = mysql()
	} else if ds == "clickhouse" {
		orm = clickhouse()
	}
	data, err := orm.Query(sql, values...)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	if q.DataModifier != nil {
		return q.DataModifier(data[0]), nil
	}
	return data[0], err
}

func (gi *goroseIml) Row(qb *queryBox) (map[string]interface{}, error) {
	sql := qb.ToSql()
	row, err := gi.row(qb.Query, sql, qb.BindValues)
	return row, err
}

func (*goroseIml) Rows(qb *queryBox) ([]map[string]interface{}, error) {
	sql := qb.ToSql()
	var orm gorose.IOrm
	ds := qb.Query.Datasource
	if ds == "mysql" {
		orm = mysql()
	} else if ds == "clickhouse" {
		orm = clickhouse()
	}
	data, err := orm.Query(sql, qb.BindValues...)
	if err != nil {
		return nil, err
	}
	var rows []map[string]interface{}
	dataModifier := qb.Query.DataModifier
	for _, m := range data {
		if dataModifier != nil {
			m = dataModifier(m)
		}
		rows = append(rows, m)
	}
	return rows, err
}

func (*goroseIml) Page(qb *queryBox) ([]map[string]interface{}, int64, error) {
	sql := qb.ToSql()
	qb.totalModel = true
	countSql := qb.ToSql()

	var orm gorose.IOrm
	ds := qb.Query.Datasource
	if ds == "mysql" {
		orm = mysql()
	} else if ds == "clickhouse" {
		orm = clickhouse()
	}
	totalRow, err := orm.Query(countSql, qb.BindValues...)
	if err != nil {
		return nil, 0, err
	}
	total := 0
	if len(totalRow) > 0 {
		v := reflect.ValueOf(totalRow[0]["total"])
		switch v.Kind() {
		case reflect.Int:
			total = totalRow[0]["total"].(int)
		case reflect.Uint64:
			total = int(totalRow[0]["total"].(uint64))
		case reflect.Int64:
			total = int(totalRow[0]["total"].(int64))
		}
	}
	data, err := orm.Query(sql, qb.BindValues...)
	if err != nil {
		return nil, 0, err
	}
	var rows []map[string]interface{}
	dataModifier := qb.Query.DataModifier
	for _, m := range data {
		if dataModifier != nil {
			m = dataModifier(m)
		}
		rows = append(rows, m)
	}
	return rows, int64(total), err
}

func (*goroseIml) total(qb *queryBox) (int64, error) {
	qb.totalModel = true
	countSql := qb.ToSql()

	var orm gorose.IOrm
	ds := qb.Query.Datasource
	if ds == "mysql" {
		orm = mysql()
	} else if ds == "clickhouse" {
		orm = clickhouse()
	}
	totalRow, err := orm.Query(countSql, qb.BindValues...)
	if err != nil {
		return 0, err
	}
	total := 0
	if len(totalRow) > 0 {
		v := reflect.ValueOf(totalRow[0]["total"])
		switch v.Kind() {
		case reflect.Int:
			total = totalRow[0]["total"].(int)
		case reflect.Uint64:
			total = int(totalRow[0]["total"].(uint64))
		case reflect.Int64:
			total = int(totalRow[0]["total"].(int64))
		}
	}
	return int64(total), err
}

func (gi *goroseIml) Export(qb *queryBox) (string, error) {
	exporter := newExporter(ExportOption{
		Name: qb.Query.Name,
	})
	err := gi.addSheet(exporter, qb.Query.Name, qb)
	if err != nil {
		return "", err
	}
	loc, _ := time.LoadLocation(config.Timezone)
	xlsx.DefaultDateTimeOptions = xlsx.DateTimeOptions{
		Location:        loc,
		ExcelTimeFormat: "yy-m-dd hh:mm:ss",
	}
	fileName, err := exporter.Export()

	if err != nil {
		return fileName, err
	}
	fStream, err := os.Open(fileName)
	if err != nil {
		return fileName, err
	}
	defer fStream.Close()
	return config.OSSPushFunc(fileName, fStream)
}

func (gi *goroseIml) ExportMultiQuery(cfg *ExportMultiQueryConfig) (string, error) {
	exporter := newExporter(ExportOption{
		Name: cfg.Name,
	})
	for _, sheet := range cfg.Sheets {
		err := gi.addSheet(exporter, sheet.Name, newQueryBox(sheet.Query))
		if err != nil {
			return "", err
		}
	}

	loc, _ := time.LoadLocation(config.Timezone)
	xlsx.DefaultDateTimeOptions = xlsx.DateTimeOptions{
		Location:        loc,
		ExcelTimeFormat: "yy-m-dd hh:mm:ss",
	}
	fileName, err := exporter.Export()

	if err != nil {
		return fileName, err
	}
	fStream, err := os.Open(fileName)
	if err != nil {
		return fileName, err
	}
	defer fStream.Close()
	return config.OSSPushFunc(fileName, fStream)
}

func (gi *goroseIml) addSheet(exporter *exporter, name string, qb *queryBox) error {
	qb.Query.Page = 1
	total, _ := gi.total(qb)
	if total > int64(config.MaxExport) {
		return newExportErr(config.MaxExport)
	}
	qb.Query.Limit = config.MaxExport
	data, _, err := gi.Page(qb)
	if err != nil {
		return err
	}
	c := qb.Query.ExportConfigure
	var columns []interface{}
	for _, c := range c.Columns {
		columns = append(columns, c.Title)
	}
	var list [][]interface{}
	for _, row := range data {
		var exportRow []interface{}
		for _, c := range qb.Query.ExportConfigure.Columns {
			exportRow = append(exportRow, row[c.Column])
		}
		list = append(list, exportRow)
	}

	exporter.AddSheet(&sheet{
		Name:    name,
		Columns: columns,
		Data:    list,
	})
	return nil
}
