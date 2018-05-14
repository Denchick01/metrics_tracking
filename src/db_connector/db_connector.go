package db_connector

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"math/rand"
	"time"
)

type Dbmanager struct {
	db *sql.DB
}

func NewDB(Host string, Port int, User string, Password string, DBname string) (*Dbmanager, error) {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		Host, Port, User, Password, DBname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &Dbmanager{db: db}, nil
}

type UserData struct {
	Id    int
	Name  string
	Email string
}

func (mgr *Dbmanager) GetUserData(id int) (*UserData, error) {
	row := mgr.db.QueryRow("SELECT * FROM users WHERE id=$1", id)
	d := &UserData{}
	err := row.Scan(&d.Id, &d.Name, &d.Email)

	return d, err
}

type DevicesData struct {
	Id     int
	Name   string
	UserId int
}

func (mgr *Dbmanager) GetDevicesData(id int) (*DevicesData, error) {
	row := mgr.db.QueryRow("SELECT * FROM devices WHERE id=$1", id)
	d := &DevicesData{}
	err := row.Scan(&d.Id, &d.Name, &d.UserId)

	return d, err
}

func (mgr *Dbmanager) SetDeviceAlert(device_id int, msg string) error {
	src := rand.NewSource(time.Now().UnixNano())
	rd := rand.New(src)
	id := rd.Int31()
	_, err := mgr.db.Exec("INSERT INTO device_alerts (id, device_id, message) VALUES ($1, $2, $3)", id, device_id, msg)
	return err
}

const NumOfMetrics int = 5

type DeviceMetricsData struct {
	Id         int
	DeviceId   int
	Metrics    [NumOfMetrics]sql.NullInt64
	LocalTime  time.Time
	ServerTime time.Time
	rows       *sql.Rows
}

func (self *DeviceMetricsData) Next() bool {
	notEnd := self.rows.Next()
	if notEnd == false {
		return notEnd
	}
	err := self.rows.Scan(
		&self.Id,
		&self.DeviceId,
		&self.Metrics[0],
		&self.Metrics[1],
		&self.Metrics[2],
		&self.Metrics[3],
		&self.Metrics[4],
		&self.LocalTime,
		&self.ServerTime,
	)
	if err != nil {
		log.Println("Data base Error:", err)
		return false
	}
	return true
}

func (self *DeviceMetricsData) Error() error {
	return self.rows.Err()
}

func (mgr *Dbmanager) GetDeviceMetricsData(tTamp time.Time, limit int) (*DeviceMetricsData, error) {
	rows, err := mgr.db.Query("SELECT * FROM device_metrics WHERE server_time>=$1 LIMIT $2", tTamp, limit)
	if err != nil {
		return nil, err
	}
	return &DeviceMetricsData{rows: rows}, nil
}

func (mgr *Dbmanager) GetMaxTimesTamp() (time.Time, error) {
	row := mgr.db.QueryRow("SELECT MAX(server_time) FROM device_metrics")

	var maxTimesTamp time.Time
	err := row.Scan(&maxTimesTamp)

	return maxTimesTamp, err
}
