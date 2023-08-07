package dmodel

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id        string `json:"id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	Age       int    `json:"age"`
	IsMarried bool   `json:"is_married"`
}

func (u *User) CheckPassword(password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return fmt.Errorf("password does not match")
	}
	return nil
}

func (u *User) GeneratePasswordHash() error {
	pwd, err := generatePasswordHash(u.Password)
	if err != nil {
		return err
	}
	u.Password = pwd
	return nil
}

func generatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password due to error %w", err)
	}
	return string(hash), nil
}
