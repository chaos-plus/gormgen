package main

import (
	"fmt"
	"strings"

	generate "github.com/chaos-plus/chaos-plus-gormgen/gen"
	"github.com/robotism/flagger"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"

	sqlite "github.com/glebarez/sqlite"
	sqlite3 "gorm.io/driver/sqlite"

	"gorm.io/gorm"
)

//go:generate go run main.go --dbsrc "sqlite://migration_dlock.db" --tables chaosplus_distributed_locks --tablePrefix chaosplus

type DaoOptions struct {
	DbSource    string `mapstructure:"dbsrc" description:"database source" default:""`
	QueryDir    string `mapstructure:"queryDir" description:"generate query output directory" default:"./query"`
	ModelDir    string `mapstructure:"modelDir" description:"generate model output directory" default:"./model"`
	Tables      string `mapstructure:"tables" description:"table names" default:""`
	TablesEx    string `mapstructure:"tablesEx" description:"table names ignore" default:""`
	TablePrefix string `mapstructure:"tablePrefix" description:"table prefix" default:""`
	Clear       bool   `mapstructure:"clear" description:"clear output directory" default:"true"`
}

func main() {
	f := flagger.New()
	c := &DaoOptions{}

	f.Parse(c)
	fmt.Printf("config: %+v\n", c)

	url := strings.Split(c.DbSource, "://")
	if len(url) != 2 {
		panic("invalid db source, should be like xxxx://xxxx")
	}
	scheme := url[0]
	path := url[1]

	var db *gorm.DB
	var err error
	switch scheme {
	case "mysql":
		db, err = gorm.Open(mysql.Open(path))
	case "pgsql", "postgre", "postgresql":
		db, err = gorm.Open(postgres.Open(path))
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(path))
	case "sqlte3":
		db, err = gorm.Open(sqlite3.Open(path))
	default:
		panic("unsupported db source: " + c.DbSource)
	}
	if err != nil {
		panic(err)
	}
	generate.Generate(db, strings.Split(c.Tables, ","), strings.Split(c.TablesEx, ","), c.TablePrefix, c.QueryDir, c.ModelDir, c.Clear)
	generate.SetTestUsePureSqlite(c.QueryDir, c.ModelDir)
	generate.AddGitIgnore(c.QueryDir, c.ModelDir, "*_test.db")
}
