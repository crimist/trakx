package tracker_test

import (
	"database/sql"
	"log"
	"testing"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

func TestDB(t *testing.T) {
	db, err := sql.Open("mysql", "root@/bittorrent")
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	table := "test"

	_, err = db.Exec("CREATE TABLE ? (x int)", table)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO ? VALUES (?)", table, 69)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO ? VALUES (?)", table, 420)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("SELECT * FROM ?", table)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var x string
		err = rows.Scan(&x)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("X: ", x)
	}

	_, err = db.Exec("DROP TABLE ?", table)
	if err != nil {
		log.Fatal(err)
	}
}
