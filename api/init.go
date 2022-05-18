package api

import (
	"github.com/fukaraca/worth2watch/auth"
	"github.com/fukaraca/worth2watch/db"
	"github.com/fukaraca/worth2watch/model"
	"github.com/fukaraca/worth2watch/util"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/ratelimit"
	"io"
	"log"
	"os"
)

type service struct {
	db.DBserver
	auth.AuthServer
	util.Utilizer
}

var (
	h *service
)

func newService() *service {
	return &service{
		DBserver:   db.NewDBServer(),
		AuthServer: auth.NewAuthServer(),
		Utilizer:   util.NewUtilizer(),
	}
}

//ListenRouter initiates the server
func ListenRouter() error {
	return setupRouter().Run(model.ServerPort)
}

func setupRouter() *gin.Engine {
	//logger middleware teed to log.file
	logfile, err := os.OpenFile("./logs/log.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Could not create/open log file")
	}
	errlogfile, err := os.OpenFile("./logs/err.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Could not create/open err log file")
	}
	gin.DefaultWriter = io.MultiWriter(logfile, os.Stdout)
	gin.DefaultErrorWriter = io.MultiWriter(errlogfile, os.Stdout)
	//starts with builtin Logger() and Recovery() middlewares
	r := gin.Default()

	//rate limiter
	rLimit := ratelimit.New(20)
	leakBucket := func(limiter ratelimit.Limiter) gin.HandlerFunc {
		return func(ctx *gin.Context) {
			limiter.Take()
		}
	}
	r.Use(leakBucket(rLimit))
	r.Use(requestid.New())
	h = newService()
	endpoints(r)
	return r
}