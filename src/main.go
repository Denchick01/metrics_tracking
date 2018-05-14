package main

import (
	"database/sql"
	"fmt"
	conf "github.com/Denchick01/metrics_tracking/src/configuration"
	dbmng "github.com/Denchick01/metrics_tracking/src/db_connector"
	rdmng "github.com/Denchick01/metrics_tracking/src/rd_connector"
	"github.com/Denchick01/metrics_tracking/src/smtp_sender"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	PathToMainConf = "../etc/config.yaml"
	PathToVergConf = "../etc/metric_verge.yaml"
)

var senderParam *smtp_sender.SenderParam
var dbMng *dbmng.Dbmanager
var rdMng *rdmng.RdManager
var metricsVergeConfig *conf.MetricsVergeConfig
var mainConfig *conf.MainConfig

type alertTask struct {
	DeviceId    int
	MetricName  string
	MetricValue int64
	DeviceTime  time.Time
	wg          sync.WaitGroup
}

type pool struct {
	concurrency int
	tasksChan   chan *alertTask
	wg          sync.WaitGroup
}

var tasksPool *pool

func updateConfigs(c chan os.Signal) {
	var err error
	for {
		<-c
		mainConfig, err = conf.ReadMainConfig(PathToMainConf)
		if err != nil {
			log.Println("Can't read MainConfig")
		}
		metricsVergeConfig, err = conf.ReadMetricsVergeConfig(PathToVergConf)
		if err != nil {
			log.Println("Can't read VergeConfig")
		}
		log.Println("Config has been update")
	}
}

func announcingAlert(deviceId int, metricName string, merticValue int64, deviceTime time.Time) {
	deviceData, err := dbMng.GetDevicesData(deviceId)
	if err != nil {
		log.Println("GetDevicesData Error:", err)
		return
	}

	alertMsg := fmt.Sprintf(
		"Metric %s have invalid value: %d [deviceId: %d, deviceName: %s]",
		metricName,
		merticValue,
		deviceId,
		deviceData.Name,
	)

	currTime, err := rdMng.GetAlertTime(deviceId)

	if err != nil && err != rdmng.ErrorKeyNotFound {
		log.Println("Can't get alert from Redis:", err)
	} else if err == rdmng.ErrorKeyNotFound || currTime.Before(deviceTime) {
		err = rdMng.SetAlertToRd(deviceId, alertMsg, deviceTime)
		if err != nil {
			log.Println("Can't set alert to Redis:", err)
		}

	}

	err = dbMng.SetDeviceAlert(deviceId, alertMsg)
	if err != nil {
		log.Println("Can't set alert to DataBase:", err)
	}

	userData, err := dbMng.GetUserData(deviceData.UserId)
	if err != nil {
		log.Println("GetUserData Error:", err)
		return
	}

	err = senderParam.SendMsg([]string{userData.Email}, "Device Alert", alertMsg)
	if err != nil {
		log.Println("Can't send massage:", err)
	}

}

func newPool(concurrency int) *pool {
	return &pool{
		concurrency: concurrency,
		tasksChan:   make(chan *alertTask),
	}
}

func (p *pool) run() {
	for i := 0; i < p.concurrency; i++ {
		p.wg.Add(1)
		go p.runWorker()
	}
}

func (p *pool) stop() {
	close(p.tasksChan)
	p.wg.Wait()
}

func (p *pool) runWorker() {
	for t := range p.tasksChan {
		announcingAlert(t.DeviceId, t.MetricName, t.MetricValue, t.DeviceTime)
		t.wg.Done()
	}
	p.wg.Done()
}

func (p *pool) addAlertTask(deviceId int, metricName string, merticValue int64, deviceTime time.Time) {
	t := alertTask{
		DeviceId:    deviceId,
		MetricName:  metricName,
		MetricValue: merticValue,
		DeviceTime:  deviceTime,
		wg:          sync.WaitGroup{},
	}

	t.wg.Add(1)
	select {
	case p.tasksChan <- &t:
		break
	case <-time.After(mainConfig.TaskTimeout):
		log.Println("Job timed out")
		return
	}

	t.wg.Wait()
}

func checkMetrics(metricsArr []sql.NullInt64, deviceId int, deviceTime time.Time) {
	var refValues map[string]conf.MetricMinMax
	//Если в конфиге нет настроек для устройства, проверка не делается
	refValues, ok := metricsVergeConfig.DevicesId[deviceId]
	if ok != true {
		return
	}

	for num, value := range metricsArr {
		//Если NULL то проверка не делается
		if value.Valid != true {
			continue
		}
		metricName := fmt.Sprintf("metric%d", num+1)

		refRange, ok := refValues[metricName]

		//Если референсного значения для метрики нет, проверка не делается
		if ok != true {
			continue
		}

		if refRange.Min > value.Int64 {
			log.Printf("[Min] Invalide metrict %s value: %d", metricName, value.Int64)
			go tasksPool.addAlertTask(deviceId, metricName, value.Int64, deviceTime)
		} else if refRange.Max < value.Int64 {
			log.Printf("[Max] Invalide metrict %s value: %d", metricName, value.Int64)
			go tasksPool.addAlertTask(deviceId, metricName, value.Int64, deviceTime)
		}
	}
}

func init() {
	log.Println("Init metrics_tracking")

	var err error

	mainConfig, err = conf.ReadMainConfig(PathToMainConf)
	if err != nil {
		log.Fatal("Can't read MainConfig:", err)
	}

	metricsVergeConfig, err = conf.ReadMetricsVergeConfig(PathToVergConf)
	if err != nil {
		log.Fatal("Can't read VergeConfig:", err)
	}

	senderParam = smtp_sender.NewMailSender(
		mainConfig.MailSender.Address,
		mainConfig.MailSender.Password,
		mainConfig.MailSender.SmtpHost,
		mainConfig.MailSender.SmtpPort,
	)

	rdMng, err = rdmng.NewRD(
		mainConfig.Redis.Host,
		mainConfig.Redis.Port,
		mainConfig.Redis.Password,
		mainConfig.Redis.Db,
	)
	if err != nil {
		log.Fatal("Can't connect to Redis:", err)
	}

	dbMng, err = dbmng.NewDB(
		mainConfig.DB.Host,
		mainConfig.DB.Port,
		mainConfig.DB.User,
		mainConfig.DB.Password,
		mainConfig.DB.DBname,
	)
	if err != nil {
		log.Fatal("Can't connect to DataBase:", err)
	}

	tasksPool = newPool(mainConfig.MaxTasksForAlert)
	tasksPool.run()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGUSR1)
	go updateConfigs(ch)
}

func main() {
	log.Println("Start metrics_tracking")

	var err error

	startTime, err := dbMng.GetMaxTimesTamp()
	if err != nil {
		log.Fatal("Can't find start time:", err)
	}

	for {
		metricsData, err := dbMng.GetDeviceMetricsData(startTime, mainConfig.MaxReqLimit)
		if err != nil {
			log.Println("[Error] Get device data metric: ", err)
			continue
		}
		for metricsData.Next() {
			checkMetrics(metricsData.Metrics[:], metricsData.DeviceId, metricsData.LocalTime)
			if metricsData.ServerTime.After(startTime) {
				startTime = metricsData.ServerTime
			}
		}
		startTime = startTime.Add(time.Nanosecond * 1000)
	}

	tasksPool.stop()
	log.Println("End metrics_tracking")
}
