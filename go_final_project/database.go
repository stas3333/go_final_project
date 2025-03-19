package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func DbInstall() error {
	EnvDbFile := os.Getenv("TODO_DBFILE")
	appPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fileDB := filepath.Join(appPath, EnvDbFile)
	// _, err = os.Stat(fileDB)
	// if err != nil {
	// 	fmt.Println("создаем базу данных")
	// }
	log.Println("создаем базу данных")
	db, err := sql.Open("sqlite", fileDB)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE scheduler (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date CHAR(8) NOT NULL DEFAULT "",
			title TEXT NOT NULL DEFAULT "",
			comment  TEXT NOT NULL DEFAULT "",
			repeat VARCHAR(128) NOT NULL DEFAULT ""
		);
		CREATE INDEX idx_date ON scheduler(date);
	`)
	return err
}
