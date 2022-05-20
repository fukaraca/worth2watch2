package main

import (
	"github.com/fukaraca/worth2watch2/api"
	"log"
)

func main() {
	s := api.NewServer()
	defer s.CloseCacheConnection()
	defer s.CloseDB()

	log.Fatalln("router has encountered an error while main.run: ", s.ListenRouter())
}
