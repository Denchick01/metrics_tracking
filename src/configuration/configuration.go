package configuration

import (
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"path/filepath"
	"time"
)

type MainConfig struct {
	DB struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBname   string `yaml:"dbname"`
	} `yaml:"db"`
	MailSender struct {
		SmtpHost string `yaml:"smtpHost"`
		SmtpPort int    `yaml:"smtpPort"`
		Address  string `yaml:"address"`
		Password string `yaml:"password"`
	} `yaml:"mailSender"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Db       int    `yaml:"db"`
		Password string `yaml:"password"`
	} `yaml:"redis"`
	MaxReqLimit      int           `yaml:"maxReqLimit"`
	MaxTasksForAlert int           `yaml:"maxTasksForAlert"`
	TaskTimeout      time.Duration `yaml:"taskTimeout"`
}

type MetricMinMax struct {
	Min int64 `yaml:"min"`
	Max int64 `yaml:"max"`
}

type MetricsVergeConfig struct {
	DevicesId map[int]map[string]MetricMinMax `yaml:"devicesId"`
}

func read(config interface{}, path string) (interface{}, error) {
	filename, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	configFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(configFile, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func ReadMainConfig(pathToConfig string) (*MainConfig, error) {
	conf, err := read(&MainConfig{}, pathToConfig)
	return conf.(*MainConfig), err
}

func ReadMetricsVergeConfig(pathToConfig string) (*MetricsVergeConfig, error) {
	conf, err := read(&MetricsVergeConfig{}, pathToConfig)
	return conf.(*MetricsVergeConfig), err
}
