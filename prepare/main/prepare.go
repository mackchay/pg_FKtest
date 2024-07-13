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

func wrapQueryRow(db *sql.DB, query string, args ...interface{}) *sql.Row {
	row := db.QueryRow(query, args...)
	return row
}

func main() {
	connStr := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	handleError(err)
	defer func(db *sql.DB) {
		err := db.Close()
		handleError(err)
	}(db)

	err = db.Ping()
	handleError(err)

	// Checking if table exists
	var tableExists bool
	err = wrapQueryRow(db, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'racers')").Scan(&tableExists)
	handleError(err)

	if !tableExists {
		wrapExec(db, "CREATE TABLE racers (pos INTEGER)")
		wrapExec(db, "INSERT INTO racers(pos) SELECT generate_series(1, 10)")
		fmt.Println("Table created and data inserted")
	} else {
		fmt.Println("Table already exists, skipping data insertion")
	}

	stmt, err := db.Prepare("SELECT * FROM racers WHERE pos = $1")
	handleError(err)
	defer func(stmt *sql.Stmt) {
		handleError(stmt.Close())
	}(stmt)

	for i := 1; i <= 10; i++ {
		rows, err := stmt.Query(i)
		handleError(err)
		defer func(rows *sql.Rows) {
			handleError(rows.Close())
		}(rows)

		for rows.Next() {
			var pos int
			err = rows.Scan(&pos)
			handleError(err)
			fmt.Printf("Position: %d\n", pos)
		}

		handleError(rows.Err())
	}
}
