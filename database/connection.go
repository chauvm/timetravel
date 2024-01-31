package database

import (
	"database/sql"
	"log"

	"github.com/chauvm/timetravel/entity"
)

const DATABASE_FILE string = "rainbow.db"

const INIT_DB string = `
 CREATE TABLE IF NOT EXISTS records (
 id INTEGER NOT NULL PRIMARY KEY,
 timestamp DATETIME NOT NULL,
 data STRING NOT NULL,
 accumulated STRING,
 version INTEGER NOT NULL
 );`

// create a SQLite3 database connection, or create the SQLite file if not existed yet
func createConnection() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", DATABASE_FILE)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	// create the table if not existed yet
	if _, err := db.Exec(INIT_DB); err != nil {
		log.Fatal(err)
		return nil, err
	}
	return db, nil
}

func InsertRecord(db *sql.DB, record entity.Record) (int, error) {
	res, err := db.Exec("INSERT INTO records VALUES(NULL,CURRENT_TIMESTAMP,?,?,?);", record.Data, record.Accumulated, record.Version)
	if err != nil {
		return 0, err
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		return 0, err
	}
	return int(id), nil
}
