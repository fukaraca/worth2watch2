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
	"time"
)

func TestCheckAdminForLoggedIn(t *testing.T) {

	fn := func(setErr error, expRet bool, msg string) {
		c := new(gin.Context)
		c.Request = &http.Request{}
		username := "furkan"
		cache := &authImp{}
		cli, mock := redismock.NewClientMock()
		defer cli.Close()

		cache.client = cli

		getStr := mock.ExpectGet("admin-" + username)
		getStr.SetErr(setErr)
		getStr.SetVal("_") //otherwise it complains
		ok := cache.CheckAdminForLoggedIn(c, username)
		assert.Equal(t, expRet, ok, msg)
	}
	fn(nil, true, "it's expected to get true")
	fn(redis.Nil, false, "it's expected to get false for user is not admin")

}

func TestCheckSession(t *testing.T) {
	fn := func(expErr, setErr error, expRet bool, msg, username, cookieKey string, sessionToken uuid.UUID) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{Header: http.Header{}}

		cookie1 := http.Cookie{Name: cookieKey, Value: sessionToken.String()}
		cookie2 := http.Cookie{Name: "uid", Value: username}
		c.Request.AddCookie(&cookie1)
		c.Request.AddCookie(&cookie2)

		cache := &authImp{}
		cli, mock := redismock.NewClientMock()
		defer cli.Close()
		mock.MatchExpectationsInOrder(true)
		cache.client = cli

		expstr := mock.ExpectGet(sessionToken.String())
		expstr.SetErr(setErr)
		expstr.SetVal(username)

		actReturn, actErr := cache.CheckSession(c)

		assert.Equal(t, expRet, actReturn, msg)
		assert.Equal(t, expErr, actErr, msg)

	}

	uid, _ := uuid.NewV4()

	fn(nil, nil, true, "it's expected to get true and nil", "someUser", "session_token", uid)
	fn(redis.Nil, redis.Nil, false, "it's expected to get false for cookie not found maybe expired", "someUser", "session_token", uid)
	fn(http.ErrNoCookie, nil, false, "it's expected to get false for errNocookie hint: cookie name is wrong .s.s", "someUser", "session_Poken", uid)

}

func TestCheckCookie(t *testing.T) {
	fn := func(setErr error, expRet bool, sessionToken uuid.UUID, msg, username, setUsername string, funcErr error) {
		c := new(gin.Context)
		c.Request = &http.Request{}

		cache := &authImp{}
		cli, mock := redismock.NewClientMock()
		defer cli.Close()

		cache.client = cli

		expstr := mock.ExpectGet(sessionToken.String())
		expstr.SetErr(setErr)
		expstr.SetVal(setUsername)

		ok, err := cache.CheckCookie(c, sessionToken.String(), username)
		assert.Equal(t, expRet, ok, msg)
		assert.Equal(t, funcErr, err)

	}
	sessTok, _ := uuid.NewV4()
	fn(redis.Nil, false, sessTok, "it's expected to return false for redis.nil err", "someUser", "someUser", redis.Nil)
	fn(errors.New("some err"), false, sessTok, "it's expected to return false for some err", "someUser", "someUser", errors.New("some err"))
	fn(nil, false, sessTok, "it's expected to return false for invalid cookie", "someUser", "", errors.New("cookie val is empty"))
	fn(nil, false, sessTok, "it's expected to return false for user-id and cookie mismatch", "someUser", "anotherUser", errors.New("invalid cookie for given user"))
	fn(nil, true, sessTok, "it's expected to return true and nil", "someUser", "someUser", nil)

}

func TestCreateSession(t *testing.T) {
	fn := func(checkSess bool, msg, username string, sessionToken uuid.UUID) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{Header: http.Header{}}

		cache := &authImp{}
		cli, mock := redismock.NewClientMock()
		defer cli.Close()
		mock.MatchExpectationsInOrder(true)
		cache.client = cli

		mock.ExpectSetEX(sessionToken.String(), username, time.Second*3600)

		cache.CreateSession(username, c)

		getUname := ""
		getSession := ""
		cookies := w.Result().Cookies()
		for _, cookie := range cookies {
			if cookie.Name == "uid" {
				getUname = cookie.Value
			}
			if cookie.Name == "session_token" {
				getSession = cookie.Value
			}
		}

		//test if cookies injected
		assert.Equal(t, 2, len(cookies))
		assert.Equal(t, username, getUname, msg)
		assert.Equal(t, checkSess, len(sessionToken.String()) == len(getSession))
	}
	uid, _ := uuid.NewV4()
	fn(true, "it's expected to get successful create", "someUser", uid)

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
