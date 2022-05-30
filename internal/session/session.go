// Package session 用于对session的管理
package session

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"strconv"
)

// EnableCookieSessions 创建一个用cookie做储存器的session中间件
func EnableCookieSessions() gin.HandlerFunc {
	//获取cookie前缀,参数是用于加密的密钥
	store := cookie.NewStore([]byte(viper.GetString("app.cookie_key")))
	log.Debug("创建cookie session成功")
	return sessions.Sessions("ginChatSession", store)
}

func EnableRedisSessions() gin.HandlerFunc {
	store, err := redis.NewStore(10, "tcp", "localhost:6379",
		"", []byte(viper.GetString("app.cookie_key")))
	if err != nil {
		panic(err)
	}
	log.Info("创建redisSession 成功")
	return sessions.Sessions("ginChatRedisSession", store)

}

// SaveAuthSession 用于保存session,登录和注册时,生成一个uid写入到session
func SaveAuthSession(c *gin.Context, info interface{}) {
	session := sessions.Default(c) //默认session
	session.Set("uid", info)
	session.Options(sessions.Options{MaxAge: 60}) //过期时间
	session.Save()                                //当前请求的session必须保存
}

// GetSessionUserInfo 尝试获取session,并根据信息从数据库中查询
func GetSessionUserInfo(c *gin.Context) map[string]interface{} {
	session := sessions.Default(c)

	uid := session.Get("uid") //获取uid,没有返回nil

	data := make(map[string]interface{})

	if uid != nil { //如果存在uid,即此用户session有效,则在数据库中查询其信息,返回
		//TODO:根据uid进行查询
		data["uid"] = "123"
		data["username"] = "leilei"
		data["avatar_id"] = 123

	}
	return data //没有信息返回的是空,即len=0的map
}

func AuthSessionMid() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		//尝试获取uid
		sessionValue := session.Get("uid")
		if sessionValue == nil {
			c.Redirect(http.StatusFound, "/") //没有登录或登录过期,返回主页
			return
		}
		//将uid转换为int(sessionValue是接口,必须先断言)
		uidInt, _ := strconv.Atoi(sessionValue.(string))
		if uidInt <= 0 { //id无效
			c.Redirect(http.StatusFound, "/")
			return
		}

		//有效的id,将id写入到请求中并放行(从session中拿出记录在Context中)
		c.Set("uid", sessionValue)

		c.Next()
		return
	}
}
