package controller

import (
	"ginChat/internal/service"
	"ginChat/internal/ws"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// Index 主页
func Index(c *gin.Context) {
	//尝试获取session并查询用户,存在跳转,否则登录
	userInfo := service.GetUserInfo(c)
	log.Infof("根据session查询到:%#v", userInfo)
	if len(userInfo) > 0 {
		c.Redirect(http.StatusFound, "/home")
		return
	}

	//未登录状态跳转至登录界面
	c.HTML(http.StatusOK, "login.html", gin.H{
		"OnlineUserCount": 10,
	})
}

// Login 处理登录
func Login(c *gin.Context) {
	service.Login(c)
}

//TODO:登出

func Logout(c *gin.Context) {
}

// Home 登录完成后展示主页的逻辑,展示聊天室入口,此处也是WS的入口
func Home(c *gin.Context) {
	userInfo := service.GetUserInfo(c) //拿到用户的信息
	rooms := []map[string]interface{}{
		{"id": 1, "num": ws.OnlineRoomUserCount(1)},
		{"id": 2, "num": ws.OnlineRoomUserCount(2)},
		{"id": 3, "num": ws.OnlineRoomUserCount(3)},
	}
	c.HTML(http.StatusOK, "index.html", gin.H{ //传给前端进行渲染
		"rooms":     rooms,
		"user_info": userInfo,
	})
}

func Room(c *gin.Context) {
	roomId := c.Param("room_id") //获取路径的查询参数
	rooms := []string{"1", "2", "3"}
	if !slice.Contain(rooms, roomId) { //默认进入1号
		c.Redirect(http.StatusFound, "/room/1")
		return
	}

	//根据session获取user的信息,以及历史消息(部分)
	userInfo := service.GetUserInfo(c)
	msgList := service.GetLimitMsg(roomId, 0)

	//返回给前端room.html页面进行处理
	c.HTML(http.StatusOK, "room.html", gin.H{
		"user_info":      userInfo,
		"msg_list":       msgList,
		"msg_list_count": len(msgList),
		"room_id":        roomId,
	})
}
