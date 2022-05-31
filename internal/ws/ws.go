package ws

import (
	"encoding/json"
	"ginChat/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type ServeInterface interface {
	// RunWs 开启WS
	RunWs(gin *gin.Context)
	// GetOnlineUserCount 获取当前在线的用户数量
	GetOnlineUserCount() int
	// GetOnlineRoomUserCount 获取各个房间的用户数量
	GetOnlineRoomUserCount(roomId int) int
}

type Serve struct {
}

func (s Serve) RunWs(gin *gin.Context) {
	Run(gin)
}

func (s Serve) GetOnlineUserCount() int {
	return GetOnlineUserCount()
}

func (s Serve) GetOnlineRoomUserCount(roomId int) int {
	return GetOnlineRoomUserCount(roomId)
}

//客户端连接详情(每个连接入房间的客户的信息)
type wsClients struct {
	Conn       *websocket.Conn `json:"conn,omitempty"`
	RemoteAddr string          `json:"remote_addr,omitempty"`
	Uid        float64         `json:"uid,omitempty"`
	Username   string          `json:"username,omitempty"`
	RoomId     string          `json:"room_id,omitempty"`
	AvatarId   string          `json:"avatar_id,omitempty"`
	ToUser     interface{}     `json:"to_user,omitempty"`
}

//消息体 TODO:未公开结构体的公开字段?
type msg struct {
	Status int         `json:"status,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

//初始化
var (
	wsUpgrader  = websocket.Upgrader{}
	clientMsg   = msg{}
	mu          = sync.Mutex{}
	rooms       = [roomCount + 1][]wsClients{} //房间数量
	privateChat = []wsClients{}                //私聊相当于是单独的一个房间
)

//定义消息的类型
const (
	Online = iota
	Offline
	Send
	GetOnlineUser
	PrivateChat

	roomCount = 3
)

func Run(c *gin.Context) {
	//必须CheckOrigin,应该仔细验证origin,避免伪造攻击
	wsUpgrader.CheckOrigin = func(r *http.Request) bool {
		//TODO:验证origin
		return true
	}
	ws, _ := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	defer ws.Close()
	mainProcess(ws)

}

//主程序,负责循环读取客户端消息,处理消息的发送
func mainProcess(c *websocket.Conn) {
	for {
		//从连接中读取消息
		_, message, err := c.ReadMessage()
		serveMsgStr := message

		//是否是处理心跳响应(由js发起),是的话直接回应
		if string(serveMsgStr) == "heartbeat" {
			c.WriteMessage(websocket.TextMessage, []byte(`{"status":0,"data":"heartbeat ok"}`))
			continue
		}

		//正常消息解码输出
		json.Unmarshal(message, &clientMsg) //写入到clientMsg
		log.Info("来自客户端的消息", clientMsg, c.RemoteAddr())
		if clientMsg.Data == nil {
			return
		}

		if err != nil {
			log.Errorf("ReadMessage Err:%v", err)
			//断开连接
			disconnect(c) //离线通知
			//c.Close()
			return
		}

		if clientMsg.Status == Online { //进入房间,即建立连接
			handleConnClients(c)
			serveMsgStr = formatServeMsgStr(Online)
		}

		if clientMsg.Status == PrivateChat { //处理私聊
			serveMsgStr = formatServeMsgStr(PrivateChat)
			toC := findToUserCoonClient()
			if toC != nil {
				toC.(wsClients).Conn.WriteMessage(websocket.TextMessage, serveMsgStr)
			}
		}

		if clientMsg.Status == Send { //发送消息
			serveMsgStr = formatServeMsgStr(Send)

		}
		if clientMsg.Status == GetOnlineUser {
			serveMsgStr = formatServeMsgStr(GetOnlineUser)
			c.WriteMessage(websocket.TextMessage, serveMsgStr)
			continue
		}

		//发送,或者上线,都进行群发
		if clientMsg.Status == Send || clientMsg.Status == Online {
			notify(c, string(serveMsgStr))
		}

	}

}

// 获取私聊的用户连接
func findToUserCoonClient() interface{} {
	_, roomIdInt := getRoomId()

	toUserUid := clientMsg.Data.(map[string]interface{})["to_uid"].(string)

	for _, c := range rooms[roomIdInt] {
		stringUid := strconv.FormatFloat(c.Uid, 'f', -1, 64)
		if stringUid == toUserUid {
			return c
		}
	}

	return nil
}
func handleConnClients(c *websocket.Conn) {
	roomId, roomIdInt := getRoomId()

	for cKey, wcl := range rooms[roomIdInt] { //挤下线
		if wcl.Uid == clientMsg.Data.(map[string]interface{})["uid"].(float64) {
			mu.Lock()
			// 通知当前用户下线
			wcl.Conn.WriteMessage(websocket.TextMessage, []byte(`{"status":-1,"data":[]}`))
			rooms[roomIdInt] = append(rooms[roomIdInt][:cKey], rooms[roomIdInt][cKey+1:]...)
			wcl.Conn.Close()
			mu.Unlock()
		}
	}

	mu.Lock()
	rooms[roomIdInt] = append(rooms[roomIdInt], wsClients{ //上线,在一个房间中加入一个客户端
		Conn:       c,
		RemoteAddr: c.RemoteAddr().String(),
		Uid:        clientMsg.Data.(map[string]interface{})["uid"].(float64),
		Username:   clientMsg.Data.(map[string]interface{})["username"].(string),
		RoomId:     roomId,
		AvatarId:   clientMsg.Data.(map[string]interface{})["avatar_id"].(string),
	})
	mu.Unlock()

}

//处理要发送的消息,将其格式化
func formatServeMsgStr(status int) []byte {
	roomId, roomIdInt := getRoomId()
	data := map[string]interface{}{
		"username": clientMsg.Data.(map[string]interface{})["username"].(string),
		"uid":      clientMsg.Data.(map[string]interface{})["uid"].(float64),
		"room_id":  roomId,
		"time":     time.Now().UnixNano() / 1e6, // 13位  10位 => now.Unix()
	}
	if status == Send || status == PrivateChat {
		data["avatar_id"] = clientMsg.Data.(map[string]interface{})["avatar_id"].(string)
		data["content"] = clientMsg.Data.(map[string]interface{})["content"].(string)

		toUidStr := clientMsg.Data.(map[string]interface{})["to_uid"].(string)
		toUid, _ := strconv.Atoi(toUidStr)

		// 保存消息
		stringUid := strconv.FormatFloat(data["uid"].(float64), 'f', -1, 64)
		intUid, _ := strconv.Atoi(stringUid)

		if _, ok := clientMsg.Data.(map[string]interface{})["image_url"]; ok {
			// 存在图片
			db.SaveContent(map[string]interface{}{
				"user_id":    intUid,
				"to_user_id": toUid,
				"content":    data["content"],
				"room_id":    data["room_id"],
				"image_url":  clientMsg.Data.(map[string]interface{})["image_url"].(string),
			})
		} else {
			db.SaveContent(map[string]interface{}{
				"user_id":    intUid,
				"to_user_id": toUid,
				"room_id":    data["room_id"],
				"content":    data["content"],
			})
		}

	}

	if status == Online {
		data["count"] = GetOnlineRoomUserCount(roomIdInt)
		data["list"] = onLineUserList(roomIdInt)
	}

	jsonStrServeMsg := msg{
		Status: status,
		Data:   data,
	}
	serveMsgStr, _ := json.Marshal(jsonStrServeMsg)

	return serveMsgStr

}

//断开 并发送下线通知
func disconnect(c *websocket.Conn) {
	_, roomIdint := getRoomId()
	for i, con := range rooms[roomIdint] { //遍历一个房间中的所有连接
		//找到自己的连接
		if con.RemoteAddr == c.RemoteAddr().String() {
			//构建离线消息并发送
			data := map[string]interface{}{
				"username": con.Username,
				"uid":      con.Uid,
				"time":     time.Now().UnixNano() / 1e6, // 13位  10位 => now.Unix()
			}
			jsonStrServeMsg := msg{
				Status: Offline,
				Data:   data,
			}
			serveMsgStr, _ := json.Marshal(jsonStrServeMsg)
			disMsg := string(serveMsgStr)

			mu.Lock()

			rooms[roomIdint] = append(rooms[roomIdint][:i], rooms[roomIdint][i+1:]...) //将当要下线的客户端剔除
			con.Conn.Close()
			mu.Unlock()
			notify(c, disMsg)
		}

	}

}
func getRoomId() (string, int) {
	roomId := clientMsg.Data.(map[string]interface{})["room_id"].(string)

	roomIdInt, _ := strconv.Atoi(roomId)
	return roomId, roomIdInt
}

// 获取在线用户列表(wsClients)
func onLineUserList(roomId int) []wsClients {
	return rooms[roomId]
}

func notify(c *websocket.Conn, msg string) {
	_, roomIdInt := getRoomId()
	//遍历一个room中的所有用户,给每个连接发送消息(除了自己)
	for _, con := range rooms[roomIdInt] {
		if con.RemoteAddr != c.RemoteAddr().String() { //排除自己
			con.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
		}
	}
}

// GetOnlineUserCount 返回所有的在线用户数
func GetOnlineUserCount() int {
	num := 0
	for i := 1; i <= roomCount; i++ {
		num = num + GetOnlineRoomUserCount(i)
	}
	return num
}

// GetOnlineRoomUserCount 返回某个房间在线的用户
func GetOnlineRoomUserCount(roomId int) int {
	return len(rooms[roomId])
}
