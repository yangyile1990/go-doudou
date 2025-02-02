package database

import (
	"github.com/iancoleman/strcase"
	"github.com/unionj-cloud/go-doudou/v2/toolkit/errorx"
	"github.com/unionj-cloud/go-doudou/v2/toolkit/executils"
	"github.com/unionj-cloud/go-doudou/v2/toolkit/gormgen"
	v3 "github.com/unionj-cloud/go-doudou/v2/toolkit/protobuf/v3"
	"github.com/unionj-cloud/go-doudou/v2/toolkit/stringutils"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"strings"
)

const (
	GormKind = "gorm"
)

func init() {
	gg := &GormGenerator{}
	ag := &AbstractBaseGenerator{
		impl:   gg,
		runner: executils.CmdRunner{},
	}
	gg.AbstractBaseGenerator = ag
	RegisterOrmGenerator(GormKind, gg)
}

var _ IOrmGenerator = (*GormGenerator)(nil)

type GormGenerator struct {
	*AbstractBaseGenerator
}

func (gg *GormGenerator) svcGo() {
	gg.g.GenerateSvcGo()
}

func (gg *GormGenerator) svcImplGo() {
	gg.g.GenerateSvcImplGo()
}

func (gg *GormGenerator) dto() {
	gg.g.GenerateDtoFile()
}

func (gg *GormGenerator) svcImplGrpc(grpcService v3.Service) {
	gg.g.GenerateSvcImplGrpc(grpcService)
}

func (gg *GormGenerator) orm() {
	gg.g.Execute()
}

func (gg *GormGenerator) fix() {
	//dir, _ := filepath.Abs(gg.Dir)
	//var files []string
	//err := filepath.Walk(dir, astutils.Visit(&files))
	//if err != nil {
	//	panic(err)
	//}
	//for _, file := range files {
	//	if filepath.Ext(file) != ".go" {
	//		continue
	//	}
	//	source, err := ioutil.ReadFile(file)
	//	if err != nil {
	//		panic(err)
	//	}
	//	fileContent := string(source)
	//	fileContent = strings.ReplaceAll(fileContent, "gorm.io/gen", "github.com/unionj-cloud/go-doudou/v2/toolkit/gormgen")
	//	ioutil.WriteFile(file, []byte(fileContent), os.ModePerm)
	//}
}

func (gg *GormGenerator) ProtoFieldNamingFn() func(string) string {
	return strcase.ToLowerCamel
}

const (
	driverMysql     = "mysql"
	driverPostgres  = "postgres"
	driverSqlite    = "sqlite"
	driverSqlserver = "sqlserver"
	driverTidb      = "tidb"
)

func (gg *GormGenerator) Initialize(conf OrmGeneratorConfig) {
	gg.Dir = conf.Dir
	gg.Driver = conf.Driver
	gg.Dsn = conf.Dsn
	gg.Client = false
	gg.Grpc = conf.Grpc
	gg.TablePrefix = strings.TrimSuffix(conf.TablePrefix, ".")
	var db *gorm.DB
	var err error
	switch gg.Driver {
	case driverMysql, driverTidb:
		conf := mysql.Config{
			DSN: gg.Dsn, // data source name
		}
		db, err = gorm.Open(mysql.New(conf))
	case driverPostgres:
		conf := postgres.Config{
			DSN: gg.Dsn,
		}
		db, err = gorm.Open(postgres.New(conf))
		if stringutils.IsNotEmpty(gg.TablePrefix) {
			db.Exec(`set search_path='` + gg.TablePrefix + `'`)
		}
	case driverSqlite:
		db, err = gorm.Open(sqlite.Open(gg.Dsn))
	case driverSqlserver:
		db, err = gorm.Open(sqlserver.Open(gg.Dsn))
	default:
		errorx.Panic("Not support driver")
	}
	if err != nil {
		errorx.Panic(err.Error())
	}
	g := gormgen.NewGenerator(gormgen.Config{
		RootDir:       gg.Dir,
		OutPath:       gg.Dir + "/query",
		Mode:          gormgen.WithDefaultQuery | gormgen.WithQueryInterface,
		FieldNullable: true,
		// if you want to assign field which has a default value in the `Create` API, set FieldCoverable true, reference: https://gorm.io/docs/create.html#Default-Values
		FieldCoverable: false,
		// if you want to generate field with unsigned integer type, set FieldSignable true
		FieldSignable: false,
		// if you want to generate index tags from database, set FieldWithIndexTag true
		FieldWithIndexTag: true,
		// if you want to generate type tags from database, set FieldWithTypeTag true
		FieldWithTypeTag: true,
		// if you need unit tests for query code, set WithUnitTest true
		WithUnitTest: false,
	})
	g.WithTableNameStrategy(func(tableName string) (targetTableName string) {
		return gg.TablePrefix + "." + tableName
	})
	g.WithJSONTagNameStrategy(func(n string) string { return n + ",omitempty" })
	g.WithImportPkgPath("github.com/unionj-cloud/go-doudou/v2/toolkit/customtypes")
	g.UseDB(db)
	g.ApplyBasic(g.GenerateAllTable(
		gormgen.FieldType(conf.Soft, "gorm.DeletedAt"),
		gormgen.FieldGenType(conf.Soft, "Time"),
	)...)
	gg.g = g
}
