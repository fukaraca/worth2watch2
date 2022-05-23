package db_test

import (
	"context"
	"fmt"
	"github.com/fukaraca/worth2watch2/config"
	"github.com/fukaraca/worth2watch2/db"
	"github.com/fukaraca/worth2watch2/model"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var (
	adminSampleString    = "adminSampleString"
	nonAdminSampleString = "nonAdminSampleString"
	sampleString1        = "sampleString1"
	sampleString2        = "sampleString2"
	uniqueEmail          = "unique@Email.com"
	uniqueUser           = model.User{
		UserID:    0,
		Username:  "uniqueUsername",
		Password:  "uniquePassword",
		Email:     &uniqueEmail,
		Name:      nil,
		Lastname:  nil,
		CreatedOn: sampleTimestamp,
		LastLogin: sampleTimestamp,
		Isadmin:   false,
	}
	sampleTimestamp     = pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now(), InfinityModifier: pgtype.None}
	test_db_Host        = config.GetEnv.GetString("TEST_DB_HOST")
	test_db_Port        = config.GetEnv.GetString("TEST_DB_PORT")
	test_db_Name        = config.GetEnv.GetString("TEST_DB_NAME")
	test_db_User        = config.GetEnv.GetString("TEST_DB_USER")
	test_db_Password    = config.GetEnv.GetString("TEST_DB_PASSWORD")
	db_InitSQL_Location = config.GetEnv.GetString("TEST_INIT_SQL_LOC")
)

var movie_list = []string{"tt0120737", "tt0111161", "tt0068646", "tt0468569", "tt0071562", "tt0050083", "tt0108052", "tt0167260"}

//long runner series
//var series_list = []string{"tt0108778", "tt0944947", "tt1286039", "tt2294189", "tt4635282", "tt0460091", "tt2375692", "tt0092400"} //long series

//mini-series for brevity
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
	teardown()
	os.Exit(run)
}

//TestAllBySequentially is wrapper for all test functions for DB queries. This method is followed for brevity to make unit test easier.
func TestAllBySequentially(t *testing.T) {
	//test account-management.go
	testCreateNewUser(t)
	testQueryLogin(t)
	testIsAdmin(t)
	testUpdateLastLogin(t)
	testUpdateUserInfo(t)
	testQueryUserInfo(t)
	//test content-management.go
	testAddMovieContent(t)
	testAddSeriesContent(t)
	//test public-queries
	testGetThisMovieFromDB(t)
	testGetThisSeriesFromDB(t)
	testGetEpisodesForaSeasonFromDB(t)
	testGetMoviesListWithPage(t)
	testGetSeriesListWithPage(t)
	testSearchContent(t)
	testFindSimilarContent(t)
	//favorites
	testAddContentToFavorites(t)
	testgetFavoriteContents(t)
	testSearchFavorites(t)

}

func teardown() {
	databaseURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", test_db_Host, test_db_Port, test_db_User, test_db_Password, test_db_Name)
	conn, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		panic("Teardown failed")
	}
	conn.Exec(context.Background(), "TRUNCATE Table users,movies,episodes,favorite_movies,favorite_series,seasons,series cascade ;")
	defer conn.Close(context.Background())
}
