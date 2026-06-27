package tool

import (
	"golang.org/x/crypto/bcrypt"
)

func HashingPass(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func CheckHashPass(hashPass, pass []byte) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hashPass, pass)
	if err != nil {
		return false, err
	}

	return true, nil
}
