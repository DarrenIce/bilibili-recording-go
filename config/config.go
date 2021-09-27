package config

import (
	"io/ioutil"
	"sync"

	"gopkg.in/yaml.v2"
)

type bili struct {
	User     string            `yaml:"user"`
	Password string            `yaml:"password"`
	Cookies  map[string]string `yaml:"cookies"`
}

type RecordConfig struct {
	NeedProxy 	bool	`yaml:"needProxy"`
	Proxy     	string	`yaml:"proxy"`
	NeedBdPan	bool	`yaml:"needBdPan"`
	UploadTime	string	`yaml:"uploadTime"`
	ApiAddr		string	`yaml:"apiAddr"`
}

// RoomConfigInfo room config info
type RoomConfigInfo struct {
	RoomID     string `yaml:"roomID"`
	StartTime  string `yaml:"startTime"`
	EndTime    string `yaml:"endTime"`
	AutoRecord bool   `yaml:"autorecord"`
	AutoUpload bool   `yaml:"autoupload"`
}

// Config 配置文件
type config struct {
	Bili     bili                      `yaml:"bilibili"`
	RcConfig RecordConfig              `yaml:"record"`
	Live     map[string]RoomConfigInfo `yaml:"live"`
}

// Config out
type Config struct {
	Conf *config
	lock *sync.Mutex
}

var (
	once sync.Once

	instance *Config
)

// New init
func New() *Config {
	once.Do(func() {
		instance = &Config{Conf: new(config), lock: new(sync.Mutex)}
	})

	return instance
}

// LoadConfig 加载配置文件
func (c *Config) LoadConfig() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	b, err := ioutil.ReadFile("./config.yml")
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(b, c.Conf); err != nil {
		return err
	}
	return nil
}

// Marshal save
func (c *Config) Marshal() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	b, err := yaml.Marshal(c.Conf)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("./config.yml", b, 0777)
}

// AddRoom add
func (c *Config) AddRoom(roomInfo RoomConfigInfo) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Conf.Live[roomInfo.RoomID] = roomInfo
}

func (c *Config) DeleteRoom(roomID string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	_, ok := c.Conf.Live[roomID]
	if ok {
		delete(c.Conf.Live, roomID)
	}
}
