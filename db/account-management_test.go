package db_test

import (
	"errors"
	"github.com/fukaraca/worth2watch2/config"
	"github.com/fukaraca/worth2watch2/db"
	"github.com/fukaraca/worth2watch2/model"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var (
	//conn        *pgxpool.Pool
	adminSampleString    = "adminSampleString"
	nonAdminSampleString = "nonAdminSampleString"
	sampleString1        = "sampleString1"
	sampleString2        = "sampleString2"
	sampleTimestamp      = pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now(), InfinityModifier: pgtype.None}
	test_db_Host         = config.GetEnv.GetString("TEST_DB_HOST")
	test_db_Port         = config.GetEnv.GetString("TEST_DB_PORT")
	test_db_Name         = config.GetEnv.GetString("TEST_DB_NAME")
	test_db_User         = config.GetEnv.GetString("TEST_DB_USER")
	test_db_Password     = config.GetEnv.GetString("TEST_DB_PASSWORD")
	db_InitSQL_Location  = config.GetEnv.GetString("INIT_SQL_LOC")
)

var movie_list = []string{"tt0120737", "tt0111161", "tt0068646", "tt0468569", "tt0071562", "tt0050083", "tt0108052", "tt0167260"}

//var series_list = []string{"tt0108778", "tt0944947", "tt1286039", "tt2294189", "tt4635282", "tt0460091", "tt2375692", "tt0092400"} //long series
var series_list = []string{"tt11324406", "tt10234724", "tt13729648"}

type mockRepo struct {
	db.Repository
}

var testDB *mockRepo

func setupTestRepository() *mockRepo {
	var mockDB = &mockRepo{db.NewTestDB(test_db_Host, test_db_Port, test_db_User, test_db_Password, test_db_Name, db_InitSQL_Location)}
	//insertTables(mockDB)
	return mockDB
}

func fillInTheTables(mockDB *mockRepo) {
	for _, m := range movie_list {
		mockDB.AddMovieContentWithID(m)
	}
	for _, s := range series_list {
		mockDB.AddSeriesContentWithID(s)
	}
}

func newContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{}
	return c
}

func TestMain(m *testing.M) {
	testDB = setupTestRepository()
	defer testDB.CloseDB()
	run := m.Run()
	db.Truncate()
	os.Exit(run)
}

func TestCreateNewUser(t *testing.T) {
	var err error
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
	}

	for _, s := range cases {
		err = testDB.CreateNewUser(c, s.user)
		assert.NotEqual(t, s.notExpectedErr, err, s.msg)
	}
}

func TestQueryLogin(t *testing.T) {
	c := newContext()

	pass, err := testDB.QueryLogin(c, adminSampleString)
	assert.Equal(t, adminSampleString, pass)
	assert.Equal(t, nil, err)
}
