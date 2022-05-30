package route

import (
	"ginChat/internal/controller"
	"ginChat/internal/session"
	"ginChat/internal/ws"
	"ginChat/static"
	"ginChat/views"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
)

// InitRoute 初始化路由,负责注册处理器和中间件
func InitRoute() *gin.Engine {
	router := gin.Default() //使用logger和recover

	if viper.GetString(`app.debug_mod`) == "false" {
		// live 模式 打包用,会将静态资源封装
		router.StaticFS("/static", http.FS(static.EmbedStatic))
	} else {
		// dev 开发用 可动态修改静态资源不需重启服务
		router.StaticFS("/static", http.Dir("static"))
	}
	router.SetHTMLTemplate(views.GoTpl) //解析embed的模板

	//注册处理器
	root := router.Group("/", session.EnableCookieSessions())
	{
		root.GET("/", controller.Index)       //先尝试获取session中的uid,查询用户信息,如有 跳转home,否则登录
		root.POST("/login", controller.Login) //登录或注册,成功会被写入一个由数据库ID 组成的uid的sesion
		root.GET("/ws", ws.Start)             //进入房间时启动ws
		{ //直接访问此处的 都必须经过session检查的中间件
			authorized := root.Group("/", session.AuthSessionMid())
			{
				authorized.GET("/home", controller.Home)
				authorized.GET("/room/:room_id", controller.Room) //进入房间
			}
		}

	}

	return router
}
