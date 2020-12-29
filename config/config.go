package config

import (
	"io/ioutil"
	"sync"

	"github.com/kataras/golog"
	"gopkg.in/yaml.v2"
)

type bili struct {
	User		string `yaml:"user"`
	Password	string `yaml:"password"`
}

type RoomConfigInfo struct {
	RoomID		string `yaml:"roomID"`
	StartTime	string `yaml:"startTime"`
	EndTime		string `yaml:"endTime"`
	AutoRecord	bool `yaml:"autorecord"`
	AutoUpload	bool `yaml:"autoupload"`
}

// Config 配置文件
type config struct {
	Bili	bili `yaml:"bilibili"`
	Live	map[string]RoomConfigInfo `yaml:"live"`
}

type Config struct {
	config	config
	lock	*sync.Mutex
}

var defaultconfig config

var (
    once sync.Once

    instance *Config
)

func InitConfig() (*Config) {
	once.Do(func() {
		instance = &Config{config: config{}, lock: new(sync.Mutex)}
	})

	return instance
}

// LoadConfig 加载配置文件
func (c *Config) LoadConfig() (*config, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	b, err := ioutil.ReadFile("./config.yml")
	if err != nil {
		golog.Fatal(err)
	}
	config := &defaultconfig
	if err = yaml.Unmarshal(b, config); err != nil {
		golog.Fatal(err)
	}
	return config, nil
}