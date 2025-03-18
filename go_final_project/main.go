package main

import (
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

	webDir := "./web"
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", getNextDate)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/tasks", getListTasks)
	http.HandleFunc("/api/task/done", doneTask)

	DbInstall()
	fmt.Println("Запускаем сервер")

	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("Завершаем работу")

}
