package queryassistor

import "io"

type Configure struct {
	Timezone       string
	Mysql          MysqlConfigure
	ClickHouse     ClickHouseConfigure
	OSSPushFunc    OSSPushFunc
	MaxExport      int
	ExportSavePath string
}

type MysqlConfigure struct {
	User        string
	Password    string
	Host        string
	Port        string
	DBName      string
	TablePrefix string
	Debug       bool
}

type ClickHouseConfigure struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
	Debug    bool
}

type SubQuery struct {
	Table     string
	Filter    []FilterItem
	Select    []SelectField
	Group     string
	Having    []FilterItem
	Joins     []Join
	Limit     int
	Order     string
	CountArgs string
	Id        string
}

type Query struct {
	Datasource   string
	Table        string
	Id           string
	Filter       []FilterItem
	Select       []SelectField
	Group        string
	Having       []FilterItem
	Joins        []Join
	Limit        int
	Order        string
	CountArgs    string
	Page         int
	PageSize     int
	DataModifier func(data map[string]interface{}) map[string]interface{}
	ExportConfigure
	MergeQueries []Query
	SubQuery     *SubQuery
}

func (q *Query) GetId() string {
	if len(q.Id) == 0 {
		return q.Table
	}
	return q.Id
}

type Join struct {
	Table string
	Type  string
	On    string
	*SubQuery
}

type SelectField struct {
	Column string
	Alias  string
}

type FilterItem struct {
	Field    string
	Operator string
	Value    interface{}
}

type ExportConfigure struct {
	Name    string
	Suffix  string
	Columns []ExportColumn
}

type ExportColumn struct {
	Alias  string
	Column string
	Title  string
}

//type IQuery interface {
//	Datasource() string
//	Table() string
//	Joins() []Join
//	Select() []SelectField
//	Group() string
//	GetFilter() []FilterItem
//	Page() int
//	PageSize() int
//	SetPageSize(size int)
//	Order() string
//	DataModifier() func(data map[string]interface{}) map[string]interface{}
//	toSql() string
//	AppendFilter(item FilterItem)
//}

type IReqQuery interface {
	Query() Query
}

type OSSPushFunc func(filename string, reader io.Reader) (string, error)

type ExportMultiQueryConfig struct {
	Name   string
	Sheets []Sheet
}

type Sheet struct {
	Name  string
	Query IReqQuery
}
