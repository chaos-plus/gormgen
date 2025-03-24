package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	_ "github.com/glebarez/sqlite"
	_ "github.com/pingcap/tidb/parser"
	_ "gorm.io/driver/mysql"
	_ "gorm.io/driver/postgres"
	_ "gorm.io/driver/sqlite"

	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	_ "gorm.io/plugin/dbresolver"
)

type CommonMethod struct {
	ID   int32
	Name *string
}

func Generate(db *gorm.DB, tables []string, tablesEx []string, tablePrefix string, queryDir, modelDir string, clear bool) {
	if !strings.HasSuffix(tablePrefix, "_") {
		tablePrefix = tablePrefix + "_"
	}
	fmt.Printf("tables: %v\n", tables)
	fmt.Printf("tablesEx: %v\n", tablesEx)
	fmt.Printf("tablePrefix: %s\n", tablePrefix)
	if clear {
		os.RemoveAll(queryDir)
		os.RemoveAll(modelDir)
	}
	db.NamingStrategy = schema.NamingStrategy{
		TablePrefix:   tablePrefix,
		SingularTable: true,
	}
	db.NowFunc = func() time.Time {
		return time.Now().UTC()
	}
	g := gen.NewGenerator(gen.Config{
		OutPath:      queryDir,
		ModelPkgPath: modelDir,
		Mode:         gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
		// 如果你希望为可为null的字段生成属性为指针类型, 设置 FieldNullable 为 true
		FieldNullable: true,
		// 如果你希望在 `Create` API 中为字段分配默认值, 设置 FieldCoverable 为 true,
		// 参考: https://gorm.io/docs/create.html#Default-Values
		FieldCoverable: true,
		// 如果你希望生成无符号整数类型字段, 设置 FieldSignable 为 true
		FieldSignable: true,
		// 如果你希望从数据库生成索引标记, 设置 FieldWithIndexTag 为 true
		FieldWithIndexTag: true,
		// 如果你希望从数据库生成类型标记, 设置 FieldWithTypeTag 为 true
		FieldWithTypeTag: true,
		// 如果你需要对查询代码进行单元测试, 设置 WithUnitTest 为 true
		WithUnitTest: true,
	})
	g.UseDB(db)

	g.WithTableNameStrategy(func(tableName string) (targetTableName string) {
		if len(tablesEx) > 0 && slices.Contains(tablesEx, tableName) {
			return ""
		}
		fixed := strings.TrimPrefix(tableName, tablePrefix)
		fixed = strings.TrimPrefix(fixed, "_")
		fixed = strings.TrimSuffix(fixed, "_")
		fmt.Printf("tableName=====> : %s, fixed: %s\n", tableName, fixed)
		return fixed
	})
	g.WithFileNameStrategy(func(tableName string) (targetTableName string) {
		fixed := strings.TrimPrefix(tableName, tablePrefix)
		fixed = strings.TrimPrefix(fixed, "_")
		fixed = strings.TrimSuffix(fixed, "_")
		return fixed
	})
	g.WithOpts(gen.WithMethod(gen.DefaultMethodTableWithNamer))
	if len(tables) == 0 {
		g.ApplyBasic(
			g.GenerateAllTable()...,
		)
	} else {
		models := []interface{}{}
		for _, table := range tables {
			models = append(models, g.GenerateModel(table))
		}
		g.ApplyBasic(models...)
	}
	g.Execute()
}

func AddGitIgnore(queryDir, modelDir string, ignores ...string) {
	// 在查询和模型目录下添加 .gitignore 文件
	for _, dir := range []string{queryDir, modelDir} {
		gitignorePath := filepath.Join(dir, ".gitignore")
		// 如果目录不存在则创建
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("创建目录失败 %s: %v\n", dir, err)
			continue
		}
		// 写入 .gitignore 文件内容
		content := []byte(strings.Join(ignores, "\n"))
		if err := os.WriteFile(gitignorePath, content, 0644); err != nil {
			fmt.Printf("写入 .gitignore 文件失败 %s: %v\n", gitignorePath, err)
		}
	}
}

func SetTestUsePureSqlite(queryDir, modelDir string) {
	// 遍历目录查找所有测试文件
	err := filepath.Walk(queryDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), "_test.go") {
			replaceInFile(path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("遍历查询目录出错: %v\n", err)
	}

	err = filepath.Walk(modelDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), "_test.go") {
			replaceInFile(path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("遍历模型目录出错: %v\n", err)
	}
}

// 替换文件中的导入语句
func replaceInFile(filepath string) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Printf("读取文件失败 %s: %v\n", filepath, err)
		return
	}

	// 替换导入语句
	newContent := strings.Replace(string(content),
		`"gorm.io/driver/sqlite"`,
		`"github.com/glebarez/sqlite"`, -1)

	err = os.WriteFile(filepath, []byte(newContent), 0644)
	if err != nil {
		fmt.Printf("写入文件失败 %s: %v\n", filepath, err)
	}
}
