package db

import (
	"fmt"
	"github.com/fukaraca/worth2watch2/model"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"golang.org/x/net/context"
	"log"
	"time"
)

//QueryLogin queries the password for given username and returns hashed-password or error depending on the result
func (dbi *dbImp) QueryLogin(c *gin.Context, username string) (string, error) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()
	result, err := conn.Query(ctx, "SELECT password FROM users WHERE username LIKE $1;", username)
	defer result.Close()
	if err != nil {
		log.Println("login query for password failed:", err)
	}

	password := ""
	for result.Next() {
		if err := result.Scan(&password); err == pgx.ErrNoRows {
			return "", fmt.Errorf("username not found")
		} else if err == nil {
			return password, nil
		}
	}
	return "", err
}

//IsAdmin checks DBService for given user whether he/she is admin or not
func IsAdmin(c *gin.Context, username string) (bool, error) {
	if gin.Mode() == gin.TestMode {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()
	isAdmin := false
	result := conn.QueryRow(ctx, "SELECT isadmin FROM users WHERE username = $1;", username)
	err := result.Scan(&isAdmin)
	if err != nil {
		log.Println("login query for password failed:", err)
		return false, err
	}

	if isAdmin {
		return true, nil
	}
	return false, nil
}

//CreateNewUser simply inserts new user contents to DBService
func (dbi *dbImp) CreateNewUser(c *gin.Context, newUser *model.User) error {
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()

	_, err := conn.Exec(ctx, "INSERT INTO users (user_id,username,password,email,name,lastname,createdon,lastlogin,isadmin)  VALUES (nextval('users_user_id_seq'),$1,$2,$3,$4,$5,$6,$7,$8);", newUser.Username, newUser.Password, newUser.Email, newUser.Name, newUser.Lastname, newUser.CreatedOn, newUser.LastLogin, newUser.Isadmin)

	if err != nil {
		return fmt.Errorf("user infos for register was failed to insert to DB:%v", err)
	}
	return nil
}

//UpdateLastLogin updates login time
func (dbi *dbImp) UpdateLastLogin(c *gin.Context, lastLoginTime time.Time, logUsername string) error {
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()
	_, err := conn.Exec(ctx, "UPDATE users SET lastlogin = $1 WHERE username = $2;", lastLoginTime, logUsername)
	if err != nil {
		return err
	}
	return nil
}

//UpdateUserInfo updates given user info
func (dbi *dbImp) UpdateUserInfo(c *gin.Context, firstname, lastname, username string) error {
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()
	_, err := conn.Exec(ctx, "UPDATE users SET name = $1,lastname=$2 WHERE username = $3;", firstname, lastname, username)
	if err != nil {
		return err
	}
	return nil
}

//QueryUserInfo returns user info from db except password
func (dbi *dbImp) QueryUserInfo(c *gin.Context, username string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()

	tempUser := new(model.User)
	row := conn.QueryRow(ctx, "SELECT * FROM users WHERE username = $1;", username)
	err := row.Scan(&tempUser.UserID, &tempUser.Username, &tempUser.Password, &tempUser.Email, &tempUser.Name, &tempUser.Lastname, &tempUser.CreatedOn, &tempUser.LastLogin, &tempUser.Isadmin)
	if err != nil {
		return nil, fmt.Errorf("scanning the user infos from DB was failed:%v", err)
	}
	tempUser.Password = ""
	return tempUser, nil
}
