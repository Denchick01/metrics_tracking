package db_connector

import (
	"database/sql"
	"fmt"
	conf "github.com/Denchick01/metrics_tracking/src/configuration"
	_ "github.com/lib/pq"
	"math/rand"
	"testing"
	"time"
)

var mainConfig *conf.MainConfig
var testDBManager *sql.DB

func TestInit(t *testing.T) {
	var err error

	mainConfig, err = conf.ReadMainConfig("../../etc/config.yaml")
	if err != nil {
		t.Error("Can't read config:", err)
	}

	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		mainConfig.DB.Host,
		mainConfig.DB.Port,
		mainConfig.DB.User,
		mainConfig.DB.Password,
		mainConfig.DB.DBname,
	)
	testDBManager, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		t.Error("DB connection Error:", err)
	}
	err = testDBManager.Ping()
	if err != nil {
		t.Error("DB connection Error:", err)
	}

}

func getRandID() int32 {
	src := rand.NewSource(time.Now().UnixNano())
	rd := rand.New(src)
	return rd.Int31()
}

func TestSimpleTest(t *testing.T) {
	var err error

	dbMng, err := NewDB(
		mainConfig.DB.Host,
		mainConfig.DB.Port,
		mainConfig.DB.User,
		mainConfig.DB.Password,
		mainConfig.DB.DBname,
	)

	if err != nil {
		t.Error(err)
		return
	}

	userID := getRandID()
	userName := "TestName"
	userEmail := "test@email.test"

	_, err = testDBManager.Exec("INSERT INTO users (id, name, email) VALUES ($1, $2, $3)", userID, userName, userEmail)

	if err != nil {
		t.Error(err)
		return
	}

	userData, err := dbMng.GetUserData(int(userID))

	if err != nil {
		t.Error(err)
		return
	}

	if userData.Name != userName {
		t.Error("Expect: ", userName, ",but got: ", userData.Name)
	}

	if userData.Email != userEmail {
		t.Error("Expect: ", userEmail, ",but got: ", userData.Email)
	}

	deviceID := getRandID()
	deviceName := "Megafon"

	_, err = testDBManager.Exec("INSERT INTO devices (id, name, user_id) VALUES ($1, $2, $3)", deviceID, deviceName, userID)

	if err != nil {
		t.Error(err)
		return
	}

	deviceData, err := dbMng.GetDevicesData(int(deviceID))

	if err != nil {
		t.Error(err)
		return
	}

	if deviceData.Name != deviceName {
		t.Error("Expect: ", deviceName, ",but got: ", deviceData.Name)
	}

	testTime, err := dbMng.GetMaxTimesTamp()

	if err != nil {
		t.Error(err)
		return
	}

	metricID := getRandID()
	metric1 := 432
	metric2 := 3234
	localTime := time.Now()

	_, err = testDBManager.Exec(
		"INSERT INTO device_metrics (id, device_id, metric_1, metric_2, metric_3, metric_4, metric_5, local_time) VALUES ($1, $2, $3, $4, NULL, NULL, NULL, $5)",
		metricID,
		deviceID,
		metric1,
		metric2,
		localTime,
	)

	if err != nil {
		t.Error(err)
		return
	}

	metricsData, err := dbMng.GetDeviceMetricsData(testTime, 10)

	for metricsData.Next() {
		fmt.Printf("Device_metrics row Id: %d, DeviceId: %d, localTime: %s\n", metricsData.Id, metricsData.DeviceId, metricsData.LocalTime)
	}
}

func TestDeviceAlert(t *testing.T) {
	var err error

	dbMng, err := NewDB(
		mainConfig.DB.Host,
		mainConfig.DB.Port,
		mainConfig.DB.User,
		mainConfig.DB.Password,
		mainConfig.DB.DBname,
	)

	if err != nil {
		t.Error(err)
		return
	}

	deviceID := getRandID()
	msg := "It's test message!"

	err = dbMng.SetDeviceAlert(int(deviceID), msg)

	if err != nil {
		t.Error(err)
		return
	}

}

func TestMaxTimesTamp(t *testing.T) {
	var err error

	dbMng, err := NewDB(
		mainConfig.DB.Host,
		mainConfig.DB.Port,
		mainConfig.DB.User,
		mainConfig.DB.Password,
		mainConfig.DB.DBname,
	)

	if err != nil {
		t.Error(err)
		return
	}

	testTime, err := dbMng.GetMaxTimesTamp()
	fmt.Println("Max TimesTamp: ", testTime)
}
