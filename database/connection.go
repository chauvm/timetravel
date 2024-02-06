package database

import (
	"database/sql"
	"log"

	"github.com/chauvm/timetravel/entity"
	_ "github.com/mattn/go-sqlite3"
)

const DATABASE_FILE string = "./rainbow.db"

const INIT_DB string = `
 CREATE TABLE IF NOT EXISTS records (
 id INTEGER NOT NULL PRIMARY KEY,
 timestamp DATETIME NOT NULL,
 data STRING NOT NULL,
 accumulated STRING,
 version INTEGER NOT NULL
 );`

// create a SQLite3 database connection, or create the SQLite file if not existed yet
func CreateConnection() (*sql.DB, error) {
	log.Println("Creating database connection...")
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

func GetRecord(db *sql.DB, id int) (*entity.Record, error) {
	rows, err := db.Query("SELECT * FROM records WHERE id = ? ORDER BY timestamp DESC LIMIT 1;", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var record entity.Record
	for rows.Next() {
		err = rows.Scan(&record.ID, &record.Timestamp, &record.Data, &record.Accumulated, &record.Version)
		if err != nil {
			return nil, err
		}
	}
	return &record, nil
}
