package db

import (
	"context"
	"fmt"
	"github.com/fukaraca/worth2watch2/config"
	"github.com/fukaraca/worth2watch2/model"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"

	"log"
	"os"
	"strings"
	"time"
)

var (
	conn                *pgxpool.Pool
	db_Host             = config.GetEnv.GetString("DB_HOST")
	db_Port             = config.GetEnv.GetString("DB_PORT")
	db_Name             = config.GetEnv.GetString("DB_NAME")
	db_User             = config.GetEnv.GetString("DB_USER")
	db_Password         = config.GetEnv.GetString("DB_PASSWORD")
	db_InitSQL_Location = config.GetEnv.GetString("INIT_SQL_LOC")
)

type dbImp struct {
	conn *pgxpool.Pool
}

type Repository interface {
	GetThisMovieFromDB(c *gin.Context, id string) (*model.Movie, error)
	GetThisSeriesFromDB(c *gin.Context, id string) (*model.Series, *[]model.Seasons, error)
	GetEpisodesForaSeasonFromDB(c *gin.Context, seriesID, sN string) (*[]model.Episodes, error)
	GetMoviesListWithPage(c *gin.Context, page, items int) (*[]model.Movie, error)
	GetSeriesListWithPage(c *gin.Context, page, items int) (*[]model.Series, error)
	SearchContent(c *gin.Context, name string, genres []string, page, items int) (*[]model.Movie, *[]model.Series, error)
	FindSimilarContent(c *gin.Context, id, cType string) (*[]model.Movie, *[]model.Series, error)
	AddContentToFavorites(c *gin.Context, IMDB, cType, username string) error
	GetFavoriteContents(c *gin.Context, page, items int, username string) (*[]model.Movie, *[]model.Series, error)
	SearchFavorites(c *gin.Context, name, username string, genres []string, page, items int) (*[]model.Movie, *[]model.Series, error)
	QueryLogin(c *gin.Context, username string) (string, error)
	CreateNewUser(c *gin.Context, newUser *model.User) error
	UpdateLastLogin(c *gin.Context, lastLoginTime time.Time, logUsername string) error
	UpdateUserInfo(c *gin.Context, firstname, lastname, username string) error
	QueryUserInfo(c *gin.Context, username string) (*model.User, error)
	AddMovieContentWithID(imdb string)
	AddSeriesContentWithID(imdb string)
	AddMovieContentWithStruct(ctx context.Context, movie *model.Movie) error
	AddSeriesContentWithStruct(ctx context.Context, series *model.Series) error
	DeleteContent(c *gin.Context, id, contentType string) error
	CloseDB()
}

//InitializeDB function creates a connection pool to PSQL DBService.
func (dbi *dbImp) initializeDB(host, port, user, password, name string) {

	databaseURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, name)
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
	dbi.conn = pool
}

func (dbi *dbImp) CloseDB() {
	dbi.conn.Close()
}

//CheckIfInitialized functions checks existance of tables and creates if necessary.
func (dbi *dbImp) checkIfInitialized(initSQL string) {
	sqlScript, err := os.ReadFile(initSQL)
	if err != nil {
		log.Fatalln("init.sql file couldn't be read: ", err)
	}
	statements := strings.Split(string(sqlScript), ";\n")
	for _, statement := range statements {
		comm, err := dbi.conn.Exec(context.Background(), statement)
		if err != nil {
			log.Println("checked for initial DB structure : ", err, comm.String())
		}
	}

}

func NewRepository() *dbImp {
	database := &dbImp{}
	database.initializeDB(db_Host, db_Port, db_User, db_Password, db_Name)
	database.checkIfInitialized(db_InitSQL_Location)
	return database
}

func NewTestDB(host, port, user, password, name, initSQL string) *dbImp {
	database := &dbImp{}
	database.initializeDB(host, port, user, password, name)
	database.checkIfInitialized(initSQL)
	return database
}
