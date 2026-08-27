package crypto

import "golang.org/x/crypto/bcrypt"

// BcryptPasswordHash 对密码进行 bcrypt 哈希
// 因为 bcrypt 使用了随机的盐，所以同一个密码每次生成的哈希值都不同，但都可以通过校验
func BcryptPasswordHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// BcryptPasswordCheck 对比密码哈希值
func BcryptPasswordCheck(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
