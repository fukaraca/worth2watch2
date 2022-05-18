package auth

import (
	"github.com/fukaraca/worth2watch/config"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
	"log"
)

var redis_Host = config.GetEnv.GetString("REDIS_HOST")
var redis_Port = config.GetEnv.GetString("REDIS_PORT")
var redis_Password = config.GetEnv.GetString("REDIS_PASSWORD")
var redis_DB = config.GetEnv.GetInt("REDIS_DB")

type authImp struct {
	client *redis.Client
}

type AuthServer interface {
	CheckCookie(c *gin.Context, toBeChecked, userId string) bool
	CreateSession(username string, c *gin.Context)
	CheckSession(c *gin.Context) bool
	DeleteSession(c *gin.Context) (bool, error)
	CheckAdminForLoggedIn(c *gin.Context, username string) bool
	CloseCacheConnection()
	InitializeCache()
}

//var AuthService AuthServer = &authImp{}

func (chc *authImp) InitializeCache() {

	client := redis.NewClient(&redis.Options{
		Addr:     redis_Host + redis_Port,
		Password: redis_Password,
		DB:       redis_DB,
	})

	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalln("redis ping error:", err)
	}
	log.Println(pong, " redis activated")

	//enable notifications from redis
	client.ConfigSet(context.Background(), "notify-keyspace-events", "KEA")
	chc.client = client

}

func (chc *authImp) CloseCacheConnection() {
	err := chc.client.Close()
	if err != nil {
		log.Println(err)
	}
}

func NewAuthServer() *authImp {
	return &authImp{}
}