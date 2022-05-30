package ws

import "github.com/gin-gonic/gin"

func Create() ServeInterface {
	return &Serve{}
}

func Start(c *gin.Context) {
	Create().RunWs(c)
}

func OnlineUserCount() int {
	return Create().GetOnlineUserCount()
}

func OnlineRoomUserCount(roomId int) int {
	return Create().GetOnlineRoomUserCount(roomId)
}
