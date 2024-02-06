package service

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/chauvm/timetravel/database"
	"github.com/chauvm/timetravel/entity"
)

var ErrRecordDoesNotExist = errors.New("record with that id does not exist")
var ErrRecordIDInvalid = errors.New("record id must >= 0")
var ErrRecordAlreadyExists = errors.New("record already exists")

// Implements method to get, create, and update record data.
type RecordService interface {

	// GetRecord will retrieve an record.
	GetRecord(ctx context.Context, id int) (entity.Record, error)

	// CreateRecord will insert a new record.
	//
	// If it a record with that id already exists it will fail.
	CreateRecord(ctx context.Context, record entity.Record) error

	// UpdateRecord will change the internal `Map` values of the record if they exist.
	// if the update[key] is null it will delete that key from the record's Map.
	//
	// UpdateRecord will error if id <= 0 or the record does not exist with that id.
	UpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.Record, error)
}

// InMemoryRecordService is an in-memory implementation of RecordService.
type InMemoryRecordService struct {
	data map[int]entity.Record
}

func NewInMemoryRecordService() InMemoryRecordService {
	return InMemoryRecordService{
		data: map[int]entity.Record{},
	}
}

func (s *InMemoryRecordService) GetRecord(ctx context.Context, id int) (entity.Record, error) {
	record := s.data[id]
	if record.ID == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}

	record = record.Copy() // copy is necessary so modifations to the record don't change the stored record
	return record, nil
}

func (s *InMemoryRecordService) CreateRecord(ctx context.Context, record entity.Record) error {
	id := record.ID
	if id <= 0 {
		return ErrRecordIDInvalid
	}

	existingRecord := s.data[id]
	if existingRecord.ID != 0 {
		return ErrRecordAlreadyExists
	}

	s.data[id] = record
	return nil
}

func (s *InMemoryRecordService) UpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.Record, error) {
	entry := s.data[id]
	if entry.ID == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}

	for key, value := range updates {
		if value == nil { // deletion update
			delete(entry.Data, key)
		} else {
			entry.Data[key] = *value
		}
	}

	return entry.Copy(), nil
}

// PersistentRecordService is a persistent implementation of RecordService.
type PersistentRecordService struct {
	db *sql.DB
}

func NewPersistentRecordService(db *sql.DB) PersistentRecordService {
	return PersistentRecordService{
		db: db,
	}
}

func (s *PersistentRecordService) GetRecord(ctx context.Context, id int) (entity.Record, error) {
	// Approach 2.2 first, assume a row's accumulated_data has everything we need
	latestRecord, err := database.GetLatestRecord(s.db, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Record{}, ErrRecordDoesNotExist
		}
		return entity.Record{}, err
	}

	return *latestRecord, nil
}

func (s *PersistentRecordService) CreateRecord(ctx context.Context, record entity.Record) error {
	log.Printf("CreateRecord in PersistentRecordService %v", record)
	id, err := database.InsertRecord(s.db, record)
	log.Printf("CreateRecord in PersistentRecordService with id %v", id)
	if err != nil {
		log.Fatal(err)
		return err
	}
	// id := record.ID
	// if id <= 0 {
	// 	return ErrRecordIDInvalid
	// }

	// existingRecord := s.data[id]
	// if existingRecord.ID != 0 {
	// 	return ErrRecordAlreadyExists
	// }

	// s.data[id] = record
	return nil
}

func (s *PersistentRecordService) UpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.Record, error) {
	latestRecord, err := s.GetRecord(ctx, id)
	if err != nil {
		return entity.Record{}, err
	}
	latestRecordVersion := latestRecord.Version

	// newRecordData is a copy of the latestRecord's data
	newRecordData := map[string]string{}
	for key, value := range latestRecord.Data {
		newRecordData[key] = value
	}

	// apply the updates to newRecordData
	for key, value := range updates {
		if value == nil {
			delete(newRecordData, key)
		} else {
			newRecordData[key] = *value
		}
	}

	// create a new record with the updated data
	newRecord := entity.Record{
		ID:      id,
		Data:    newRecordData,
		Updates: latestRecord.Updates,
		Version: latestRecordVersion + 1,
	}

	_, err = database.InsertRecord(s.db, newRecord)

	if err != nil {
		return entity.Record{}, err
	}

	return newRecord, nil
}
