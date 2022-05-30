package service

import "ginChat/internal/db"

func GetLimitMsg(roomId string, offset int) []map[string]interface{} {
	return db.GetLimitMsg(roomId, offset)
}

/*func GetLimitPrivateMsg(uid, toUId string, offset int) []map[string]interface{} {
	return db.GetLimitPrivateMsg(uid, toUId, offset)
}*/
