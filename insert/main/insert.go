package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"sync"
	"time"
)

func tableExists(db *sql.DB, tableName string) bool {
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = " + tableName + ")").
		Scan(&exists)
	handleError(err)
	return exists
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func wrapExec(db *sql.DB, query string, args ...interface{}) {
	_, err := db.Exec(query, args...)
	handleError(err)
}

func wrapQueryTxLocks(tx *sql.Tx, tableName string, goroutine int) {
	rows, err := tx.Query("SELECT * FROM pgrowlocks(" + tableName + ") limit 1")
	handleError(err)
	for rows.Next() {
		var lockedRow, locker, multi, xids, modes, pids string
		err := rows.Scan(&lockedRow, &locker, &multi, &xids, &modes, &pids)
		handleError(err)
		fmt.Println("goroutine:", goroutine, ", table: ", tableName, ", locked row:", lockedRow, ", locker:", locker,
			", multi:", multi,
			", xids:", xids, ", modes:", modes, ", pids:", pids)
	}
}

func main() {
	connStr := "user=postgres dbname=postgres password=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	handleError(err)
	defer func(db *sql.DB) {
		err := db.Close()
		handleError(err)
	}(db)

	// Checking if table exists
	if !tableExists(db, "'teams'") {
		wrapExec(db, "CREATE TABLE teams (tid serial PRIMARY KEY)")
		for i := 1; i <= 10; i++ {
			wrapExec(db, "INSERT INTO teams (tid) VALUES (DEFAULT)")
		}
		fmt.Println("Table created")
	} else {
		fmt.Println("Table already exists, skipping data insertion")
	}
	if !tableExists(db, "'employees'") {
		wrapExec(db,
			"CREATE TABLE employees (eid serial PRIMARY KEY, tid INTEGER REFERENCES teams (tid) ON UPDATE CASCADE)")
		fmt.Println("Table created")
	} else {
		fmt.Println("Table already exists, skipping data insertion")
	}

	var wg sync.WaitGroup
	execTx := func(query string, goroutine int, sleepSecs time.Duration) {
		fmt.Println("Start goroutine: ", goroutine)
		defer wg.Done()
		tx, err := db.Begin()
		handleError(err)
		_, err = tx.Exec(query)
		handleError(err)
		wrapQueryTxLocks(tx, "'employees'", goroutine)
		wrapQueryTxLocks(tx, "'teams'", goroutine)

		time.Sleep(sleepSecs * time.Second)
		handleError(err)
		err = tx.Commit()
	}

	numTx := 25
	wg.Add(numTx)
	for i := 1; i <= numTx; i++ {
		time.Sleep(time.Second)
		go execTx("INSERT INTO employees(eid, tid) VALUES (DEFAULT, 1)", i, 12)
	}

	wg.Wait()
}
