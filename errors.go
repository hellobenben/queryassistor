package queryassistor

import (
	"errors"
	"fmt"
)

var ErrExport = errors.New("export too data")

type ExportError struct {
	MaxExportNum int
	Err          error
}

func newExportErr(maxExportNum int) *ExportError {
	return &ExportError{
		MaxExportNum: maxExportNum,
		Err:          ErrExport,
	}
}

func (e *ExportError) Error() string {
	return fmt.Sprintf("导出失败，最大支持导出%d条记录", e.MaxExportNum)
}

func IsExportErr(err error) bool {
	if e, ok := err.(*ExportError); ok {
		if e.Err == ErrExport {
			return true
		}
	}
	return false
}
