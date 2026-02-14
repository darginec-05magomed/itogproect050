package main

import (
	"go1f/pkg/db"
	"go1f/pkg/server"
	"go1f/tests"
	"log"
	"os"
	"strconv"
)

func main() {
	if err := db.Init(db.DBFile); err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}
	port := tests.Port
	if p := os.Getenv("TODO_PORT"); p != "" {
		if parse, err := strconv.Atoi(p); err == nil {
			port = parse
		}
	}
	server.Run(port)
}
