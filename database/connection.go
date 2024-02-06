package database

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"

	"github.com/chauvm/timetravel/entity"
	_ "github.com/mattn/go-sqlite3"
)

const DATABASE_FILE string = "./rainbow.db"
const DATABASE_FILE_UNIT_TEST string = "./rainbow_test.db"

const INIT_DB string = `
 CREATE TABLE IF NOT EXISTS records (
 id INTEGER NOT NULL,
 timestamp DATETIME NOT NULL,
 data STRING NOT NULL,
 accumulated STRING,
 version INTEGER NOT NULL,
 PRIMARY KEY (id ASC, version DESC)
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

func CreateConnectionUnitTests() (*sql.DB, error) {
	// if DATABASE_FILE_UNIT_TEST exists, remove it
	if err := os.Remove(DATABASE_FILE_UNIT_TEST); err != nil {
		log.Println(err)
	}
	db, err := sql.Open("sqlite3", DATABASE_FILE_UNIT_TEST)
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
	log.Printf("InsertRecord in database %v", record)
	// res, err := db.Exec("INSERT INTO records VALUES(NULL,CURRENT_TIMESTAMP,?,?,?);", record.Data, record.Accumulated, record.Version)
	dataJson, err := json.Marshal(record.Data)
	if err != nil {
		return 0, err
	}
	accumulatedJson, err := json.Marshal(record.Accumulated)
	if err != nil {
		return 0, err
	}
	res, err := db.Exec("INSERT INTO records (id, version, timestamp, data, accumulated) VALUES (?, ?, CURRENT_TIMESTAMP, ?, ?)",
		record.ID, record.Version, dataJson, accumulatedJson)

	if err != nil {
		return 0, err
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		return 0, err
	}
	return int(id), nil
}

func GetLatestRecord(db *sql.DB, id int) (*entity.Record, error) {
	row := db.QueryRow("SELECT id, timestamp, data, accumulated, version FROM records WHERE id = ? ORDER BY timestamp DESC LIMIT 1", id)
	record := entity.Record{}

	if row == nil {
		return &record, nil
	}

	// parse the row and put in the record
	var rawData string
	var rawAccumulated string
	err := row.Scan(&record.ID, &record.Timestamp, &rawData, &rawAccumulated, &record.Version)
	if err != nil {
		return nil, err
	}

	// parse the insertion data
	var data map[string]string = make(map[string]string)
	err = json.Unmarshal([]byte(rawData), &data)
	if err != nil {
		return &record, err
	}

	// parse the accumulated data
	var accumulated map[string]string = make(map[string]string)
	err = json.Unmarshal([]byte(rawAccumulated), &accumulated)
	if err != nil {
		return &record, err
	}

	record.Data = data
	record.Accumulated = accumulated

	return &record, nil
}

func GetRecords(db *sql.DB, id int) ([]*entity.Record, error) {
	rows, err := db.Query("SELECT * FROM records WHERE id = ?;", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*entity.Record
	for rows.Next() {
		var record entity.Record
		err = rows.Scan(&record.ID, &record.Timestamp, &record.Data, &record.Accumulated, &record.Version)
		if err != nil {
			return nil, err
		}
		records = append(records, &record)
	}
	return records, nil
}
