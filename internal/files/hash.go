package files

import "golang.org/x/crypto/bcrypt"

func CheckPassword(provided_secret, database_hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(database_hash), []byte(provided_secret))
	if err != nil {
		return false
	}
	return true
}

func GenerateHash(secret string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
