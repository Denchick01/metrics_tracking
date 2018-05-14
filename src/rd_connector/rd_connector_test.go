package rd_connector

import (
	conf "github.com/Denchick01/metrics_tracking/src/configuration"
	"testing"
	"time"
)

func TestGetDeviceTime(t *testing.T) {
	var err error

	mainConfig, err := conf.ReadMainConfig("../../etc/config.yaml")
	if err != nil {
		t.Error("Can't read config:", err)
	}

	rdClient, err := NewRD(
		mainConfig.Redis.Host,
		mainConfig.Redis.Port,
		mainConfig.Redis.Password,
		mainConfig.Redis.Db,
	)

	deviceID := 42
	alertMsg := "It's test"
	deviceTime := time.Now()

	err = rdClient.SetAlertToRd(deviceID, alertMsg, deviceTime)

	if err != nil {
		t.Error("Error when set value: ", err)
		return
	}

	value, err := rdClient.GetAlertTime(deviceID)

	if err != nil {
		t.Error("Error when get value: ", err)
		return
	}

	if deviceTime.Unix() != value.Unix() {
		t.Error("Epected: ", deviceTime.Unix(), " but got: ", value.Unix())
		return
	}
}
