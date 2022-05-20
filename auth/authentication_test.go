package auth

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckAdminForLoggedIn(t *testing.T) {

	fn := func(expErr error, expRet bool, msg string) {
		c := new(gin.Context)
		c.Request = &http.Request{}
		username := "furkan"
		cache := &authImp{}
		cli, mock := redismock.NewClientMock()
		defer cli.Close()

		cache.client = cli

		//mock.ExpectGet("admin-furkan").SetVal()
		expstr := mock.ExpectGet("admin-" + username)
		expstr.SetErr(expErr)
		ok := cache.CheckAdminForLoggedIn(c, username)
		assert.Equal(t, expRet, ok, msg)
	}
	fn(nil, true, "it's expected to get true")
	fn(redis.Nil, false, "it's expected to get false for user is not admin")

}

func TestCheckSession(t *testing.T) {

}

func TestCheckCookie(t *testing.T) {

}

func TestCreateSession(t *testing.T) {

}

func TestDeleteSession(t *testing.T) {
	fn := func(expErr error, expRet bool, msg, username string, sessionToken uuid.UUID, funcErr error) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{Header: http.Header{}}

		cookie1 := http.Cookie{Name: "session_token", Value: sessionToken.String()}
		cookie2 := http.Cookie{Name: "uid", Value: username}
		c.Request.AddCookie(&cookie1)
		c.Request.AddCookie(&cookie2)

		cache := &authImp{}
		cli, mock := redismock.NewClientMock()
		defer cli.Close()
		mock.MatchExpectationsInOrder(true)
		cache.client = cli

		retInt := mock.ExpectDel(sessionToken.String())
		retInt.SetErr(expErr)

		if expErr == nil {
			retInt.SetVal(1)
			mock.ExpectGet("admin-" + username).SetErr(nil)

		}

		ok, err := cache.DeleteSession(c)
		assert.Equal(t, expRet, ok, msg)
		assert.Equal(t, funcErr, err, msg)

	}

	uid, _ := uuid.NewV4()
	//not found user in cache
	fn(redis.Nil, false, "it's expected to get redis.nil", "test", uid, errors.New("session deletion err:"+redis.Nil.Error()))
	//successful deletion for non-admin user
	fn(nil, true, "it's expected to get true and nil", "test", uid, nil)

}
