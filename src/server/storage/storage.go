package storage

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

type DeviceStorage struct {
	db *sql.DB
}

const (
	host     = "localhost"
	port     = 5432
	user     = "user"
	password = "pass"
	dbname   = "demo_db"
)

type Device struct {
	Id   string
	Type string
}

func NewDeviceStorage() DeviceStorage {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname))
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}

	// Check if the connection is successful
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging the database: ", err)
	}
	fmt.Println("Successfully connected to PostgreSQL!")

	return DeviceStorage{
		db: db,
	}
}

func (cs *DeviceStorage) AddDevice() {
	deviceId := "device_id_" + strconv.Itoa(time.Now().Second())
	_, err := cs.db.Exec("INSERT INTO devices(device_id, device_type) VALUES($1, $2)", deviceId, "sensor")
	if err != nil {
		log.Fatal("Error executing query: ", err)
	}

	log.Printf("Record inserted, id: %s", deviceId)
}

func (cs *DeviceStorage) GetAllDevices() ([]Device, error) {
	rows, err := cs.db.Query("SELECT device_id, device_type FROM devices")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices := make([]Device, 0)
	for rows.Next() {
		var device Device
		if err := rows.Scan(&device.Id, &device.Type); err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}

	return devices, nil
}
