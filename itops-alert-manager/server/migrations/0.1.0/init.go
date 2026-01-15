package __1_0

import (
	"fmt"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/infrastructure/db"
	_ "github.com/kweaver-ai/proton-rds-sdk-go/driver"
)

func InitDataBase() {
	// 替换为你的数据库连接信息
	db, err := db.ConnectDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect database 'itops': %v", err))
	}
	// 1. 创建数据库（如果不存在）
	createDBSQL := `CREATE DATABASE IF NOT EXISTS itops CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;`
	_, err = db.Exec(createDBSQL)
	if err != nil {
		panic(fmt.Sprintf("Failed to create database 'itops': %v", err))
	}
	fmt.Println("✅ Database 'itops' created or already exists.")

	// 2. 切换到 itops 数据库
	_, err = db.Exec("USE itops")
	if err != nil {
		panic(fmt.Sprintf("Failed to use database 'itops': %v", err))
	}

	// 3. 创建表 t_config（如果不存在）
	createTableSQL := `
CREATE TABLE IF NOT EXISTS t_config (
    f_config_key VARCHAR(255) NOT NULL PRIMARY KEY,
    f_config_value TEXT NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		panic(fmt.Sprintf("Failed to create table 't_config': %v", err))
	}
	fmt.Println("✅ Table 't_config' created or already exists.")
}
