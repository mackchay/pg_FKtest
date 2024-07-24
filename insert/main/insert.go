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

func startTransaction(db *sql.DB) {
	var wg sync.WaitGroup
	execTx := func(query string, goroutine int, sleepSecs time.Duration) {
		fmt.Println("Start goroutine: ", goroutine)
		defer wg.Done()
		tx, err := db.Begin()
		handleError(err)
		//1 query
		_, err = tx.Exec(query)
		handleError(err)

		//2 query
		_, err = tx.Exec(query)
		handleError(err)
		wrapQueryTxLocks(tx, "'fk_table'", goroutine)
		wrapQueryTxLocks(tx, "'pk_table'", goroutine)

		time.Sleep(sleepSecs * time.Second)
		handleError(err)
		err = tx.Commit()
	}

	numTx := 5
	wg.Add(numTx)
	for i := 1; i <= numTx; i++ {
		time.Sleep(time.Second)
		go execTx("INSERT INTO fk_table(fktid, pktid) VALUES (DEFAULT, 1)", i, 12)
	}
	wg.Wait()
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
	if !tableExists(db, "'pk_table'") {
		wrapExec(db, "CREATE TABLE pk_table (pktid serial PRIMARY KEY)")
		for i := 1; i <= 10; i++ {
			wrapExec(db, "INSERT INTO pk_table (pktid) VALUES (DEFAULT)")
		}
		fmt.Println("Table created")
	} else {
		fmt.Println("Table already exists, skipping data insertion")
	}
	if !tableExists(db, "'fk_table'") {
		wrapExec(db,
			"CREATE TABLE fk_table (fktid serial PRIMARY KEY, pktid INTEGER REFERENCES pk_table (pktid) ON UPDATE CASCADE)")
		fmt.Println("Table created")
	} else {
		fmt.Println("Table already exists, skipping data insertion")
	}

	var oldLSN, newLSN string
	err = db.QueryRow("select pg_current_wal_lsn()").Scan(&oldLSN)
	handleError(err)
	startTransaction(db)

	err = db.QueryRow("select pg_current_wal_lsn()").Scan(&newLSN)
	handleError(err)

	//Print wal_dump command to use in "postgres" user.
	fmt.Println("pg_waldump", "-s", oldLSN, "-e", newLSN)
}
