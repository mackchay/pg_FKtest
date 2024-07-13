package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func wrapExec(db *sql.DB, query string, args ...interface{}) {
	_, err := db.Exec(query, args...)
	handleError(err)
}

func tableExists(db *sql.DB, tableName string) bool {
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)", tableName).
		Scan(&exists)
	handleError(err)
	return exists
}

func main() {
	connStr := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	handleError(err)
	defer func(db *sql.DB) {
		err := db.Close()
		handleError(err)
	}(db)

	// Checking if tables exist
	if !tableExists(db, "teams") {
		wrapExec(db, "CREATE TABLE teams (tid serial PRIMARY KEY)")
		fmt.Println("Table created")
	} else {
		fmt.Println("Table already exists, skipping data insertion")
	}
	if !tableExists(db, "employees") {
		wrapExec(db,
			"CREATE TABLE employees (eid serial PRIMARY KEY, tid INTEGER REFERENCES teams (tid) ON UPDATE CASCADE)")
		fmt.Println("Table created")
	} else {
		fmt.Println("Table already exists, skipping data insertion")
	}

	var oldLSN, newLSN string
	err = db.QueryRow("select pg_current_wal_lsn()").Scan(&oldLSN)
	handleError(err)
	//wrapExec(db, "INSERT INTO teams(tid) VALUES (DEFAULT)")
	wrapExec(db, "INSERT INTO employees(eid, tid) VALUES (DEFAULT, 1)")
	err = db.QueryRow("select pg_current_wal_lsn()").Scan(&newLSN)

	fmt.Println("pg_waldump", "-s", oldLSN, "-e", newLSN)
}
