package queryassistor

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"math/rand"
	"strconv"
	"time"
)

type ExportOption struct {
	Name   string
	Suffix string
}

type sheet struct {
	Name    string
	Columns []interface{}
	Data    [][]interface{}
}

type exporter struct {
	option   ExportOption
	sheets   []*sheet
	fileName string
	savePath string
}

/**
lib.ReportService.New(func(options *lib.Options) {
	options.Name = fmt.Sprintf("广告%s信息", cpx)
	options.Title = []interface{}{"A","B"}
	options.Data = [][]interface{}{{"1","2"},{"A","B"}}
	options.Suffix = string(cpx)
}).
Report().
Oss()

*/
func newExporter(opt ExportOption) *exporter {
	return &exporter{
		option:   opt,
		fileName: "",
	}
}

func (e *exporter) AddSheet(sheet *sheet) {
	e.sheets = append(e.sheets, sheet)
}

// Export
func (e *exporter) Export() (string, error) {
	var file *excelize.File
	var err error
	file = excelize.NewFile()
	sheet1Existed := false
	for _, st := range e.sheets {
		_ = file.NewSheet(st.Name)
		if st.Name == "Sheet1" {
			sheet1Existed = true
		}
		sw, _ := file.NewStreamWriter(st.Name)
		err = sw.SetRow("A1", st.Columns)
		if err != nil {
			return "", err
		}
		for i, data := range st.Data {
			err := sw.SetRow(fmt.Sprintf("A%d", i+2), data)
			if err != nil {
				return "", err
			}
		}
		if err := sw.Flush(); err != nil {
			return "", err
		}
	}

	if !sheet1Existed {
		file.DeleteSheet("Sheet1")
	}

	fileName := e.savePath + time.Now().Format("20060102030405") + strconv.Itoa(rand.Int()) + "_" + e.option.Suffix + ".xlsx"
	if err = file.SaveAs(fileName); err != nil {
		return fileName, err
	}
	return fileName, nil
}
