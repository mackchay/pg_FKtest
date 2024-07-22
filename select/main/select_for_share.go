package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"strconv"
	"sync"
	"time"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func wrapQueryTxLocks(tx *sql.Tx, tableName string, goroutine int) {
	rows, err := tx.Query("SELECT * FROM pgrowlocks('" + tableName + "') limit 1")
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

func wrapExec(db *sql.DB, query string, args ...interface{}) {
	_, err := db.Exec(query, args...)
	handleError(err)
}

func tableExists(db *sql.DB, tableName string) bool {
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = '" + tableName + "')").
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

	//Create pgrowlocks extension
	wrapExec(db, "CREATE EXTENSION IF NOT EXISTS pgrowlocks")

	// Checking if tables exist
	if !tableExists(db, "pk_table") {
		wrapExec(db, "CREATE TABLE pk_table (pktid serial PRIMARY KEY)")
		for i := 1; i <= 10; i++ {
			wrapExec(db, "INSERT INTO pk_table (pktid) VALUES ("+strconv.FormatInt(int64(i), 10)+")")
		}
		fmt.Println("Table created")
	} else {
		fmt.Println("Table already exists, skipping data insertion")
	}
	if !tableExists(db, "fk_table") {
		wrapExec(db,
			"CREATE TABLE fk_table (fktid serial PRIMARY KEY, pktid INTEGER REFERENCES pk_table (pktid) ON UPDATE CASCADE)")
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
		wrapQueryTxLocks(tx, "fk_table", goroutine)
		wrapQueryTxLocks(tx, "pk_table", goroutine)

		time.Sleep(sleepSecs * time.Second)
		handleError(err)
		err = tx.Commit()
	}

	numTx := 5
	wg.Add(numTx * 2)
	for i := 1; i <= numTx; i++ {
		time.Sleep(1 * time.Second)
		go execTx("SELECT * FROM fk_table FOR SHARE", i, 12)
		time.Sleep(1 * time.Second)
		go execTx("UPDATE pk_table SET pktid = pktid + 1000 WHERE pktid != 1", i+10, 1)
	}
	wg.Wait()
}
