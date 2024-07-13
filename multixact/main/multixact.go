package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"sync"
	"time"
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
		handleError(db.Close())
	}(db)
	handleError(db.Ping())

	_, err = db.Exec("CREATE EXTENSION IF NOT EXISTS pgrowlocks")
	handleError(err)
	handleError(err)

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

	var wg sync.WaitGroup

	execTx := func(query string, goroutine int, sleepSecs time.Duration) {
		defer wg.Done()
		tx, err := db.Begin()
		handleError(err)
		_, err = tx.Exec(query)
		rows, err := tx.Query("SELECT * FROM pgrowlocks('employees') limit 1")
		handleError(err)
		for rows.Next() {
			var lockedRow, locker, multi, xids, modes, pids string
			err := rows.Scan(&lockedRow, &locker, &multi, &xids, &modes, &pids)
			handleError(err)
			fmt.Println("goroutine:", goroutine, ", table: employees", ", locked row:", lockedRow, ", locker:", locker,
				", multi:", multi,
				", xids:", xids, ", modes:", modes, ", pids:", pids)
		}
		rows, err = tx.Query("SELECT * FROM pgrowlocks('teams') limit 1")
		for rows.Next() {
			var lockedRow, locker, multi, xids, modes, pids string
			err := rows.Scan(&lockedRow, &locker, &multi, &xids, &modes, &pids)
			handleError(err)
			fmt.Println("goroutine:", goroutine, ", table: teams", ", locked row:", lockedRow, ", locker:", locker,
				", multi:", multi,
				", xids:", xids, ", modes:", modes, ", pids:", pids)
		}
		handleError(err)
		time.Sleep(sleepSecs * time.Second)
		handleError(err)
		err = tx.Commit()
	}

	numTx := 5
	wg.Add(numTx * 2)
	for i := 1; i <= numTx; i++ {
		time.Sleep(1 * time.Second)
		go execTx("SELECT * FROM employees FOR SHARE", i, 12)
		time.Sleep(1 * time.Second)
		go execTx("UPDATE teams SET tid = tid + 100", i+10, 1)
	}
	wg.Wait()
}
