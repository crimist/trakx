package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
	_ "github.com/lib/pq"
)

func main() {
	pgdb, err := sql.Open("postgres", "dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer pgdb.Close()
	err = pgdb.Ping()
	if err != nil {
		log.Fatal(err)
	}

	_, err = pgdb.Exec("CREATE TABLE IF NOT EXISTS peerdb (ts TIMESTAMP DEFAULT now(), bytes TEXT)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = pgdb.Query("INSERT INTO peerdb(bytes) VALUES($1)", []byte("TESTDATA"))
	if err != nil {
		log.Fatal(err)
	}

	// get latest backup
	var data []byte
	err = pgdb.QueryRow("SELECT bytes FROM peerdb ORDER BY ts DESC LIMIT 1").Scan(&data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", data)

	// delete records older than 7 days
	result, err := pgdb.Exec("DELETE FROM peerdb WHERE ts < NOW() - INTERVAL '7 days'")
	if err != nil {
		log.Fatal(err)
	}
	spew.Dump(result.RowsAffected())
}
