// Package service 提供服务
package service

import (
	"ginChat/internal/db"
	"ginChat/internal/session"
	"ginChat/pkg/pwd"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func GetUserInfo(c *gin.Context) map[string]interface{} {
	return session.GetSessionUserInfo(c)
}

// Login 获取login.html中提交的登录信息,与数据库进行比对,密码正确或注册成功写入一个由ID组成的uid session
//写入成功由页面进行重定向
func Login(c *gin.Context) {
	//获取参数前端的参数由submit的postform来
	var u User
	u.Username = c.PostForm("username")
	u.Password = c.PostForm("password")
	u.AvatarId = c.PostForm("avatar_id") //默认是1
	log.Infof("POST参数:%#v", u)
	if err := c.ShouldBind(&u); err != nil { //使用参数验证
		c.JSON(http.StatusOK, gin.H{"code": 5000, "msg": err.Error()})
		return
	}
	//查询是否存在
	user := db.FindUserByField("username", u.Username)

	if user.ID > 0 {
		if err := pwd.CheckPWD(u.Password, user.Password); err != nil { //存在用户,验证密码
			c.JSON(http.StatusOK, gin.H{
				"code": 5000,
				"msg":  "密码错误",
			})
			return
		}
	} else { //不存在用户,创建新用户,将密码哈希后存入数据库
		encodedPWD, _ := pwd.HashPWD(u.Password)
		inputUser := db.User{Username: u.Username, Password: encodedPWD, AvatarId: u.AvatarId}
		ID, err := db.Adduser(inputUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 5001,
				"msg":  err})
			return
		}
		user.ID = ID
	}
	//赋予一个session,以数据库ID为key,转换为string
	if user.ID > 0 {
		session.SaveAuthSession(c, strconv.Itoa(int(user.ID)))
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": 5001,
			"msg":  "系统错误",
		})
		return
	}
}

type User struct {
	Username string `json:"username" binding:"required,max=16,min=2"` //最初16 小12
	Password string `json:"password" binding:"required,max=32,min=6"`
	AvatarId string `json:"avatar_id" binding:"required,numeric"` //必须是数字
}
