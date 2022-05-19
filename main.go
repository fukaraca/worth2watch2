package main

import (
	"github.com/fukaraca/worth2watch2/api"
	"github.com/fukaraca/worth2watch2/auth"
	"github.com/fukaraca/worth2watch2/db"
	"log"
)

func main() {

	auth.AuthService.InitializeCache()
	defer auth.AuthService.CloseCacheConnection()

	db.DBService.InitializeDB()
	defer db.DBService.CloseDB()

	log.Fatalln("router has encountered an error while main.run: ", api.ListenRouter())
}
