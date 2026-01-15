package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

const (
	DataBaseName = "itops"
	DriverName   = "mysql"
)

var (
	dbOnce sync.Once
	db     *sql.DB
)

// 配置db的客户端参数
func NewDBAccess() *sql.DB {
	dbOnce.Do(func() {
		_, _ = NewDB()
	})
	return db
}

func NewDB() (*sql.DB, error) {
	conf := config.Get().Mysql
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&loc=Local",
		conf.Username, conf.Password, conf.Host, conf.Port, DataBaseName)
	var err error
	// 打开连接失败
	db, err = sql.Open(DriverName, dsn)
	if err != nil {
		return nil, errors.Wrapf(err, "new db err:")
	}
	if err := db.Ping(); err != nil {
		return nil, errors.Wrapf(err, "db ping err")
	}
	// 最大连接数
	db.SetMaxOpenConns(100)
	// 闲置连接数
	db.SetMaxIdleConns(20)
	// 最大连接周期
	db.SetConnMaxLifetime(100 * time.Second)

	return db, nil
}

func ConnectDB() (*sql.DB, error) {
	conf := config.Get().Mysql

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", conf.Username, conf.Password, conf.Host, conf.Port)
	fmt.Printf("dsn %v\n", dsn)
	var err error
	db, err = sql.Open(DriverName, dsn)
	if err != nil {
		return nil, errors.Wrapf(err, "new db err:")
	}
	if err := db.Ping(); err != nil {
		return nil, errors.Wrapf(err, "db ping err")
	}
	return db, nil
}
