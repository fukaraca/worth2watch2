package db

import (
	"context"
	"fmt"
	"github.com/fukaraca/worth2watch2/config"
	"github.com/fukaraca/worth2watch2/model"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	context2 "golang.org/x/net/context"
	"log"
	"os"
	"strings"
	"time"
)

var db_Host = config.GetEnv.GetString("DB_HOST")
var db_Port = config.GetEnv.GetString("DB_PORT")
var db_Name = config.GetEnv.GetString("DB_NAME")
var db_User = config.GetEnv.GetString("DB_USER")
var db_Password = config.GetEnv.GetString("DB_PASSWORD")

var (
	conn *pgxpool.Pool
)

//var DBService DBserver = &dbImp{}

type dbImp struct{}

type DBserver interface {
	GetThisMovieFromDB(c *gin.Context, id string) (*model.Movie, error)
	GetThisSeriesFromDB(c *gin.Context, id string) (*model.Series, *[]model.Seasons, error)
	GetEpisodesForaSeasonFromDB(c *gin.Context, seriesID, sN string) (*[]model.Episodes, error)
	GetMoviesListWithPage(c *gin.Context, page, items int) (*[]model.Movie, error)
	GetSeriesListWithPage(c *gin.Context, page, items int) (*[]model.Series, error)
	SearchContent(c *gin.Context, name string, genres []string, page, items int) (*[]model.Movie, *[]model.Series, error)
	FindSimilarContent(c *gin.Context, id, cType string) (*[]model.Movie, *[]model.Series, error)
	AddContentToFavorites(c *gin.Context, IMDB, cType string) error
	GetFavoriteContents(c *gin.Context, page, items int) (*[]model.Movie, *[]model.Series, error)
	SearchFavorites(c *gin.Context, name string, genres []string, page, items int) (*[]model.Movie, *[]model.Series, error)
	QueryLogin(c *gin.Context, username string) (string, error)
	//IsAdmin(c *gin.Context, username string) (bool, error)
	CreateNewUser(c *gin.Context, newUser *model.User) error
	UpdateLastLogin(c *gin.Context, lastLoginTime time.Time, logUsername string) error
	UpdateUserInfo(c *gin.Context, firstname, lastname, username string) error
	QueryUserInfo(c *gin.Context, username string) (*model.User, error)
	AddMovieContentWithID(imdb string)
	AddSeriesContentWithID(imdb string)
	AddMovieContentWithStruct(ctx context2.Context, movie *model.Movie) error
	AddSeriesContentWithStruct(ctx context2.Context, series *model.Series) error
	DeleteContent(c *gin.Context, id, contentType string) error
	InitializeDB()
	CloseDB()
}

//InitializeDB function creates a connection pool to PSQL DBService.
func (dbi *dbImp) InitializeDB() {

	databaseURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", db_Host, db_Port, db_User, db_Password, db_Name)
	pool, err := pgxpool.Connect(context.Background(), databaseURL)
	if err != nil {
		log.Fatalln("DB connection error:", err)
	}
	//check whether connection is ok or not
	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatalln("Ping to DB error:", err)
	}

	conn = pool
	checkIfInitialized()
}

func (dbi *dbImp) CloseDB() {
	conn.Close()
}

//CheckIfInitialized functions checks existance of tables and creates if necessary.
func checkIfInitialized() {
	sqlScript, err := os.ReadFile("./db/init.sql")
	if err != nil {
		log.Fatalln("init.sql file couldn't be read: ", err)
	}
	statements := strings.Split(string(sqlScript), ";\n")
	for _, statement := range statements {
		comm, err := conn.Exec(context.Background(), statement)
		if err != nil {
			log.Println("checked for initial DB structure : ", err, comm.String())
		}
	}

}

func NewDBServer() *dbImp {
	return &dbImp{}
}
