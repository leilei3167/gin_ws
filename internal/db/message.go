package db

import (
	"gorm.io/gorm"
	"sort"
	"strconv"
)

// Message 存入db中的message字段
type Message struct {
	gorm.Model
	UserId   int    //发送方
	ToUserId int    //目标用户
	RoomId   int    //房间
	Content  string //内容
	ImageUrl string //图片内容
}

// SaveContent 将消息存入数据库
func SaveContent(value interface{}) Message {
	var m Message

	m.UserId = value.(map[string]interface{})["user_id"].(int)
	m.UserId = value.(map[string]interface{})["user_id"].(int)
	m.ToUserId = value.(map[string]interface{})["to_user_id"].(int)
	m.Content = value.(map[string]interface{})["content"].(string)

	roomIdStr := value.(map[string]interface{})["room_id"].(string)
	roomIdInt, _ := strconv.Atoi(roomIdStr)

	m.RoomId = roomIdInt
	//是否有图片
	if _, ok := value.(map[string]interface{})["image_url"]; ok {
		m.ImageUrl = value.(map[string]interface{})["image_url"].(string)
	}
	DB.Create(&m)
	return m
}

func GetLimitMsg(roomId string, offset int) []map[string]interface{} {
	var results []map[string]interface{}

	DB.Model(&Message{}).
		Select("messages.*, users.username ,users.avatar_id").
		Joins("INNER Join users on users.id = messages.user_id").
		Where("messages.room_id = " + roomId).
		Where("messages.to_user_id = 0").
		Order("messages.id desc").
		Offset(offset).
		Limit(100).
		Scan(&results)

	if offset == 0 { //排序
		sort.Slice(results, func(i, j int) bool {
			return results[i]["id"].(uint32) < results[j]["id"].(uint32)
		})
	}

	return results
}
