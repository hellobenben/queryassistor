package queryassistor

import (
	"fmt"
	_ "github.com/ClickHouse/clickhouse-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gohouse/gorose/v2"
	"reflect"
	"sync"
)

var config Configure
var mysqlEngin *gorose.Engin
var mysqlGRoseOnce sync.Once
var iml *goroseIml

var clickhouseEngin *gorose.Engin
var clickhouseOnce sync.Once
var commands map[string]func(q *queryBox, params []string, v reflect.Value)

func Initialize(configure Configure) {
	config = configure
	if config.MaxExport <= 0 {
		config.MaxExport = 10000
	}
	commands = make(map[string]func(q *queryBox, params []string, v reflect.Value))
	commands["filter"] = commandFilter
	commands["page"] = commandPage
	commands["pagesize"] = commandPageSize
	commands["timeFilter"] = commandTimeFilter
	commands["dateFilter"] = commandDateFilter
	commands["having"] = commandHaving
	commands["nullFilter"] = commandNullFilter
	commands["order"] = commandOrder
	iml = &goroseIml{}
}

func mysql() gorose.IOrm {
	mysqlGRoseOnce.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
			config.Mysql.User, config.Mysql.Password, config.Mysql.Host, config.Mysql.Port, config.Mysql.DBName)
		var err error
		mysqlEngin, err = gorose.Open(&gorose.Config{
			Driver: "mysql",
			Dsn:    dsn,
		})
		if err != nil {
			panic(err)
		}
	})
	return mysqlEngin.NewOrm()
}

func clickhouse() gorose.IOrm {
	clickhouseOnce.Do(func() {
		dsn := fmt.Sprintf("tcp://%s:%s/?username=%s&password=%s&database=%s&debug=%t",
			config.ClickHouse.Host,
			config.ClickHouse.Port,
			config.ClickHouse.User,
			config.ClickHouse.Password,
			config.ClickHouse.DBName,
			config.ClickHouse.Debug,
		)
		var err error
		clickhouseEngin, err = gorose.Open(&gorose.Config{
			Driver: "clickhouse",
			Dsn:    dsn,
		})
		if err != nil {
			panic(err)
		}
	})

	return clickhouseEngin.NewOrm()
}
