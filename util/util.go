package util

import (
	"golang.org/x/crypto/bcrypt"
	"strings"
)

type utl struct{}

type Utilizer interface {
	Striper(str string) *string
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) bool
}

//Striper is a function for trimming whitespaces
func (u *utl) Striper(str string) *string {
	str = strings.TrimSpace(str)
	if str == "" {
		return nil
	}
	return &str
}

//HashPassword function hashes password with bcrypt algorithm as Cost value and return hashed string value with an error
func (u *utl) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	return string(bytes), err
}

//CheckPasswordHash function checks two inputs and returns TRUE if matches
func (u *utl) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func NewUtilizer() *utl {
	return &utl{}
}

//This is non-method version
func Striper(str string) *string {
	str = strings.TrimSpace(str)
	if str == "" {
		return nil
	}
	return &str
}
