package db

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `json:"username" `
	Password string `json:"password"`
	AvatarId string `json:"avatar_id" gorm:"avatar_id"`
}

func FindUserByField(field, value string) User {
	var u User
	if field == "id" || field == "username" {
		DB.Where(field+"=?", value).First(&u) //select * from User Where field = ?
	}
	return u
}

func Adduser(u User) (uint, error) {
	result := DB.Create(&u)

	return u.ID, result.Error
}
