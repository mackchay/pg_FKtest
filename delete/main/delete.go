package main

import (
	"database/sql"
	"flag"
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

func wrapExecTx(tx *sql.Tx, query string, args ...interface{}) {
	_, err := tx.Exec(query, args...)
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

func execQuery(db *sql.DB, query string, args ...interface{}) {
	var wg sync.WaitGroup

	execTx := func(query string, goroutine int, sleepSecs time.Duration) {
		fmt.Println("Start goroutine: ", goroutine)
		defer wg.Done()
		tx, err := db.Begin()
		handleError(err)

		_, err = tx.Exec(query)
		handleError(err)

		time.Sleep(sleepSecs * time.Second)
		err = tx.Commit()
		handleError(err)
	}

	numTx := 2
	wg.Add(numTx)
	for i := 1; i <= numTx; i++ {
		time.Sleep(10 * time.Millisecond)
		go execTx(query, i, 1)
	}
	wg.Wait()
}

func main() {
	//Cmd
	debug := flag.Bool("fk", false, "Choose query in table with foreign key")
	flag.Parse()

	var query string
	if *debug {
		fmt.Println("Query with foreign key")
		query = "DELETE FROM pk_table WHERE pktid = 1"
	} else {
		fmt.Println("Query without foreign key, use '-fk' for query with foreign key")
		query = "DELETE FROM fk_table WHERE fktid = 1"
	}

	connStr := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	handleError(err)
	defer func(db *sql.DB) {
		err := db.Close()
		handleError(err)
	}(db)

	tx, err := db.Begin()
	handleError(err)

	// Checking if tables exist
	if !tableExists(db, "pk_table") {
		wrapExecTx(tx, "CREATE TABLE pk_table (pktid serial PRIMARY KEY)")
		for i := 1; i <= 5; i++ {
			wrapExecTx(tx, "INSERT INTO pk_table (pktid) VALUES (DEFAULT)")
		}
		fmt.Println("Table created")
	} else {
		fmt.Println("Table already exists, skipping data insertion")
	}
	if !tableExists(db, "fk_table") {
		wrapExecTx(tx,
			"CREATE TABLE fk_table (fktid serial PRIMARY KEY, pktid INTEGER REFERENCES pk_table (pktid) ON DELETE CASCADE)")
		for i := 1; i <= 5; i++ {
			wrapExecTx(tx, "INSERT INTO fk_table (fktid, pktid) VALUES (DEFAULT, "+fmt.Sprintf("%d\n", i)+")")
		}
		fmt.Println("Table created")
	} else {
		fmt.Println("Table already exists, skipping data insertion")
	}
	err = tx.Commit()
	handleError(err)

	var oldLSN, newLSN string
	err = db.QueryRow("select pg_current_wal_lsn()").Scan(&oldLSN)
	handleError(err)
	execQuery(db, query)
	err = db.QueryRow("select pg_current_wal_lsn()").Scan(&newLSN)

	//Print wal_dump command to use in "postgres" user.
	fmt.Println("pg_waldump", "-s", oldLSN, "-e", newLSN)

	tx, err = db.Begin()
	handleError(err)

	wrapExecTx(tx, "DROP TABLE fk_table")
	wrapExecTx(tx, "DROP TABLE pk_table")
	err = tx.Commit()
	handleError(err)
}
