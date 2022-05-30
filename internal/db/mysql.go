package db

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := viper.GetString("mysql.dsn")
	gormconfig := gorm.Config{} //可对gorm进行配置
	DB, err = gorm.Open(mysql.Open(dsn), &gormconfig)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("连接数据库成功")
}
