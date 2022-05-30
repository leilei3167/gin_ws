package main

import (
	"bytes"
	"ginChat/internal/db"
	"ginChat/internal/route"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var logger = log.New()

func init() {
	viper.SetConfigType("json") //配置文件类型为json
	if err := viper.ReadConfig(bytes.NewReader(AppJsonConfig)); err != nil {
		log.Fatal(err)
	}
	db.InitDB()
	l := viper.GetString("app.log_level")
	if l != "debug" {
		l = "trace"
	}
	level, err := log.ParseLevel(l)
	if err != nil {
		log.Fatal("初始化logger失败:", err)
	}
	logger.SetLevel(level)
}

func main() {
	gin.SetMode(viper.GetString("app.log_level"))
	port := viper.GetString("app.port")
	r := route.InitRoute()

	log.Printf("开启服务,监听:%v,日志级别:%v", port, viper.GetString("app.log_level"))
	log.Fatal(r.Run(":" + port))

}

//配置文件
var AppJsonConfig = []byte(`
{
  "app": {
    "port": "8080",
    "upload_file_path": "./temp",
    "cookie_key": "4238uihfieh49r3453kjdfg",
    "serve_type": "GoServe",
    "debug_mod": "false",
"log_level": "debug"
  },
  "mysql": {
    "dsn": "root:8888@tcp(127.0.0.1:3306)/go_gin_chat?charset=utf8mb4&parseTime=True&loc=Local"
  },
"redis": {

}
}
`)
