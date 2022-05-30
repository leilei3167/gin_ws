package pwd

import "golang.org/x/crypto/bcrypt"

// HashPWD 将裸密码加密
func HashPWD(pwd string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// CheckPWD 检查裸密码是否和哈希后的密码一致,一致返回nil,否则返回错误
func CheckPWD(password, hashed string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}
