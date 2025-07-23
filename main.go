package main

import (
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"

	"github.com/RE-L0825/final-worked/pkg/api"
	"github.com/RE-L0825/final-worked/pkg/db"
)

func main() {
	if err := db.Init("scheduler.db"); err != nil {
		log.Fatal("Ошибка БД:", err)
	}
	defer db.DB.Close()

	api.Init()

	port := "7540"
	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		port = envPort
	}

	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	log.Printf("Сервер запущен на порту :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
