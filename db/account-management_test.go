package db_test

import (
	"errors"
	"fmt"
	"github.com/fukaraca/worth2watch2/db"
	"github.com/fukaraca/worth2watch2/model"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func testCreateNewUser(t *testing.T) {

	c := newContext()
	cases := []struct {
		name           string
		user           *model.User
		notExpectedErr error
		msg            string
	}{
		{
			name:           "case for missing info",
			user:           &model.User{Email: &sampleString1, Password: sampleString2, CreatedOn: sampleTimestamp, Username: sampleString1, Isadmin: false},
			notExpectedErr: nil,
			msg:            "it's expected to return err for not null constraint due to missing lastlogin",
		},
		{
			name:           "case for successful creation",
			user:           &model.User{Email: &adminSampleString, Password: adminSampleString, CreatedOn: sampleTimestamp, Username: adminSampleString, LastLogin: sampleTimestamp, Isadmin: true},
			notExpectedErr: errors.New("FAIL"),
			msg:            "it's expected to return nil",
		},
		{
			name:           "case for unique constraint",
			user:           &model.User{Email: &adminSampleString, Password: adminSampleString, CreatedOn: sampleTimestamp, Username: adminSampleString, LastLogin: sampleTimestamp, Isadmin: true},
			notExpectedErr: nil,
			msg:            "it's expected to return err for unique constraint",
		},
		{
			name:           "case for arbitrary non-admin insertion",
			user:           &model.User{Email: &nonAdminSampleString, Password: nonAdminSampleString, CreatedOn: sampleTimestamp, Username: nonAdminSampleString, LastLogin: sampleTimestamp, Isadmin: false},
			notExpectedErr: errors.New("FAIL"), //not equal will be used
			msg:            "it's not expected to return err",
		},
		{
			name:           "case for arbitrary non-admin pass!=username insertion",
			user:           &uniqueUser,
			notExpectedErr: errors.New("FAIL"), //not equal assertion will be used
			msg:            "it's not expected to return err",
		},
	}

	for _, s := range cases {
		err := testDB.CreateNewUser(c, s.user)
		assert.NotEqual(t, s.notExpectedErr, err, s.msg)
	}
}

func testQueryLogin(t *testing.T) {
	c := newContext()

	cases := []struct {
		name        string
		username    string
		expectedErr error
		password    string
		msg         string
	}{
		{
			name:        "happy case for username==password user that inserted just in TestCreateNewUser func",
			username:    adminSampleString,
			expectedErr: nil,
			password:    adminSampleString,
			msg:         "we didn't expect an error",
		},
		{
			name:        "not happy case for username!=password user that inserted just in TestCreateNewUser func",
			username:    uniqueUser.Username,
			expectedErr: fmt.Errorf("username not found"),
			password:    uniqueUser.Password,
			msg:         "we expect an error says username not found",
		},
	}
	for _, s := range cases {
		pass, err := testDB.QueryLogin(c, s.username)
		assert.Equal(t, s.password, pass)
		assert.Equal(t, nil, err)
	}

}

func testIsAdmin(t *testing.T) {
	c := newContext()
	gin.SetMode(gin.DebugMode)

	cases := []struct {
		name         string
		username     string
		expectedErr  error
		expectedBool bool
		msg          string
	}{
		{
			name:         "happy case for admin",
			username:     adminSampleString,
			expectedErr:  nil,
			expectedBool: true,
			msg:          "it's expected to return no err and true",
		},
		{
			name:         "happy case for non-admin",
			username:     nonAdminSampleString,
			expectedErr:  nil,
			expectedBool: false,
			msg:          "it's expected to return no err and false",
		},
		{
			name:         "non-happy case for non-existed username",
			username:     "some username",
			expectedErr:  pgx.ErrNoRows,
			expectedBool: false,
			msg:          "it's expected to return err and false",
		},
	}

	for _, s := range cases {
		ok, err := db.IsAdmin(c, s.username)
		assert.Equal(t, s.expectedBool, ok, s.msg)
		assert.Equal(t, s.expectedErr, err, s.msg)
	}
}

func testUpdateLastLogin(t *testing.T) {
	c := newContext()
	err := testDB.UpdateLastLogin(c, uniqueUser.LastLogin.Time.Add(5*time.Second), uniqueUser.Username)
	//we don't expect any error unless db fails
	assert.Equal(t, nil, err)
}

func testUpdateUserInfo(t *testing.T) {
	c := newContext()
	//we will add name and lastname to non-admin user that's created on TestNewUserCreate
	err := testDB.UpdateUserInfo(c, nonAdminSampleString+"Name", nonAdminSampleString+"Lastname", nonAdminSampleString)
	//we don't expect any error unless db fails
	assert.Equal(t, nil, err)
}

func testQueryUserInfo(t *testing.T) {
	c := newContext()
	//Now we can check users that2s created before
	cases := []struct {
		name         string
		username     string
		expectedUser *model.User
		expectedErr  error
		msg          string
	}{
		{
			name:         "check admin user's some fields",
			username:     adminSampleString,
			expectedUser: &model.User{Username: adminSampleString, Password: adminSampleString, Isadmin: true},
			expectedErr:  nil,
			msg:          "we dont expect any error",
		},
		{
			name:         "check unique user's some fields",
			username:     uniqueUser.Username,
			expectedUser: &uniqueUser,
			expectedErr:  nil,
			msg:          "we dont expect any error",
		},
		{
			name:         "check non-admin user's some fields",
			username:     nonAdminSampleString,
			expectedUser: &model.User{Username: nonAdminSampleString, Password: nonAdminSampleString, Isadmin: false},
			expectedErr:  nil,
			msg:          "we dont expect any error",
		},
		{
			name:         "check for not-existed user",
			username:     "some user",
			expectedUser: nil,
			expectedErr:  pgx.ErrNoRows,
			msg:          "we dont expect any error",
		},
	}

	for _, s := range cases {
		user, err := testDB.QueryUserInfo(c, s.username)
		if err != nil {
			assert.Equal(t, s.expectedErr, err)
			continue
		}
		assert.Equal(t, s.expectedErr, err)
		assert.Equal(t, s.expectedUser.Username, user.Username)
		assert.Equal(t, "", user.Password, "we don't expect it return password with user infos")
		assert.Equal(t, s.expectedUser.Isadmin, user.Isadmin)

		//since we time traveled unique user by 5 seconds in TestUpdateLastLogin func we can check it now
		if s.username == uniqueUser.Username {
			userLastLogin := user.LastLogin.Time.Second()
			if userLastLogin < 5 {
				userLastLogin += 60
			}
			assert.Equal(t, 5*time.Second.Seconds(), float64(userLastLogin-s.expectedUser.LastLogin.Time.Second()), "it's expected to return 5 second later of variable definition")
		}

		//since we update user's first and lastnameint TestUpdateUserInfo func, we can confirm it now
		if user.Username == nonAdminSampleString {
			assert.Equal(t, nonAdminSampleString+"Name", *user.Name)
			assert.Equal(t, nonAdminSampleString+"Lastname", *user.Lastname)
		}
	}
}
