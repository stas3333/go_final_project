package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

var webDir = "./web"

var DateFormat = "20060102"
var dbFile = "./scheduler.db"

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	port := os.Getenv("TODO_PORT")
	if err != nil {
		log.Fatal(err)
	}

	// Открываем соединение с базой данных
	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatalf("Не удалось открыть базу данных: %v", err)
	}
	defer db.Close()

	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", getNextDate)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/tasks", taskHandlerDb(getListTasks))
	http.HandleFunc("/api/task/done", taskHandlerDb(doneTask))

	base := DbInstall()
	log.Println(base)

	fmt.Println("Запускаем сервер на порту", port)

	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("Завершаем работу")

}
