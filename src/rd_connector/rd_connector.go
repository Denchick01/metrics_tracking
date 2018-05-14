package rd_connector

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

const (
	ErrorKeyNotFound = redis.Nil
	DataFormat       = time.UnixDate
)

type RdManager struct {
	RdClient *redis.Client
}

func NewRD(host string, port int, password string, db int) (*RdManager, error) {
	rdClient := redis.NewClient(
		&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", host, port),
			Password: password,
			DB:       db,
		},
	)

	var err error

	pong, err := rdClient.Ping().Result()
	if err != nil {
		return nil, errors.New(fmt.Sprintln(pong, err))
	}

	return &RdManager{rdClient}, nil
}

func (self *RdManager) SetAlertToRd(deviceId int, msg string, deviceTime time.Time) error {
	key := fmt.Sprintf("Alert_%d", deviceId)
	value := map[string]interface{}{"msg": msg, "deviceTime": deviceTime.Format(DataFormat)}
	err := self.RdClient.HMSet(key, value).Err()
	return err
}

func (self *RdManager) GetAlertTime(deviceId int) (time.Time, error) {
	value, err := self.RdClient.HGet(fmt.Sprintf("Alert_%d", deviceId), "deviceTime").Result()

	if err != nil {
		return time.Time{}, ErrorKeyNotFound
	}

	resValue, err := time.Parse(DataFormat, value)

	if err != nil {
		return time.Time{}, err
	}

	return resValue, nil
}
