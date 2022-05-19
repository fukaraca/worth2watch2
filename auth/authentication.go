package auth

import (
	"context"
	"fmt"
	"github.com/fukaraca/worth2watch2/db"
	"github.com/fukaraca/worth2watch2/model"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gofrs/uuid"
	"log"
	"net/http"
)

//CheckCookie function checks validation of cookie. Return TRUE if it's valid
func (chc *authImp) CheckCookie(c *gin.Context, toBeChecked, userId string) bool {
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()
	cookieVal, err := chc.client.Do(ctx, "GET", toBeChecked).Result()

	switch {
	case err == redis.Nil:
		log.Println("Cookie does not exist!")
		return false
	case err != nil:
		log.Println("Get Failed:", err)
		return false
	case cookieVal == "":
		log.Println("Cookie value is empty!")
		return false
	case userId != cookieVal.(string):
		return false
	}
	return true
}

//CreateSession creates and assigns cookie for user who logged in successfully. Session-token id will be stored in cache.
func (chc *authImp) CreateSession(username string, c *gin.Context) {
	sessionToken, err := uuid.NewV4()
	if err != nil {
		log.Println("new UUID couldn't assigned error:", err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()

	if ok, err := db.IsAdmin(c, username); ok && err == nil {
		chc.client.Do(ctx, "SETEX", "admin-"+username, "3600", true)
	}

	chc.client.Do(ctx, "SETEX", sessionToken.String(), "3600", username)

	c.SetCookie("session_token", sessionToken.String(), 3600, "/", model.ServerHost, false, true)
	c.SetCookie("uid", username, 3600, "/", model.ServerHost, false, true)

}

//CheckSession function checks validation of session. If a request has no cookie or cookie is not valid then returns FALSE
func (chc *authImp) CheckSession(c *gin.Context) bool {
	toBeChecked, err := c.Cookie("session_token")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return false
	}

	toBeCheckedId, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return false
	}
	//tobeCheckedId variable is like supersecret private key
	if isCookieValid := chc.CheckCookie(c, toBeChecked, toBeCheckedId); !isCookieValid {
		log.Println("Cookie is not valid", toBeChecked)
		return false
	}

	return true
}

//DeleteSession deletes the session as named
func (chc *authImp) DeleteSession(c *gin.Context) (bool, error) {
	toBeChecked, _ := c.Cookie("session_token")
	username, _ := c.Cookie("session_token")

	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()
	res, err := chc.client.Do(ctx, "DEL", toBeChecked).Result()
	if err != nil {
		log.Println("session deletion err:", err)
		return false, fmt.Errorf("session deletion err:%s", err.Error())
	}
	if chc.CheckAdminForLoggedIn(c, username) {
		res, err = chc.client.Do(ctx, "DEL", "admin-"+username).Result()
		if err != nil {
			log.Println("session deletion err for admin:", err)
			return false, fmt.Errorf("session deletion err:%s", err.Error())
		}
	}

	log.Printf("%v item removed in order to delete session.\n", res)

	c.SetCookie("session_token", "", -1, "/", model.ServerHost, false, true)
	c.SetCookie("uid", "", -1, "/", model.ServerHost, false, true)

	return true, nil
}

//CheckAdminForLoggedIn queries the cache if given user is Admin or not
func (chc *authImp) CheckAdminForLoggedIn(c *gin.Context, username string) bool {
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()
	cookieVal, err := chc.client.Do(ctx, "GET", "admin-"+username).Result()
	if err != nil {
		//log.Println("query cache for admin register failed:", err)
		return false
	}
	if cookieVal.(string) == "1" {
		return true
	}
	return false
}
