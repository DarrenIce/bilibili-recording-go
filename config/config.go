package config

import (
	"io/ioutil"
	"sync"

	"gopkg.in/yaml.v2"
)

var (
	ConfigFile = "./config.yml"
)

type bili struct {
	User     string            `yaml:"user"`
	Password string            `yaml:"password"`
	Cookies  map[string]string `yaml:"cookies"`
}

type Douyin struct {
	Cookies string `yaml:"cookies"`
}

type RecordConfig struct {
	NeedProxy        bool   `yaml:"needProxy"`
	Proxy            string `yaml:"proxy"`
	NeedBdPan        bool   `yaml:"needBdPan"`
	UploadTime       string `yaml:"uploadTime"`
	NeedRegularClean bool   `yaml:"needRegularClean"`
	RegularCleanTime string `yaml:"regularCleanTime"`
}

// RoomConfigInfo room config info
type RoomConfigInfo struct {
	Platform       string `yaml:"platform"`
	RoomID         string `yaml:"roomID"`
	Name           string `yaml:"name"`
	RecordMode     bool   `yaml:"recordMode"`
	StartTime      string `yaml:"startTime"`
	EndTime        string `yaml:"endTime"`
	AutoRecord     bool   `yaml:"autorecord"`
	AutoUpload     bool   `yaml:"autoupload"`
	NeedM4a        bool   `yaml:"needM4a"`
	Mp4Compress    bool   `yaml:"mp4Compress"`
	DivideByTitle  bool   `yaml:"divideByTitle"`
	CleanUpRegular bool   `yaml:"cleanUpRegular"`
	SaveDuration   string `yaml:"saveDuration"`
	AreaLock       bool   `yaml:"areaLock"`
	AreaLimit      string `yaml:"areaLimit"`
	SaveDanmu      bool   `yaml:"saveDanmu"`
	Cookies        string `yaml:"cookies"`
}

type MonitorArea struct {
	Platform string `yaml:"platform"`
	AreaName string `yaml:"areaName"`
	ParentID string `yaml:"parentID"`
	AreaID   string `yaml:"areaID"`
}

// Config 配置文件
type config struct {
	Bili         bili                      `yaml:"bilibili"`
	Douyin       Douyin                    `yaml:"douyin"`
	RcConfig     RecordConfig              `yaml:"record"`
	Live         map[string]RoomConfigInfo `yaml:"live"`
	MonitorAreas []MonitorArea             `yaml:"monitorAreas"`
	BlockedRooms []string                  `yaml:"blockedRooms"`
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
	b, err := ioutil.ReadFile(ConfigFile)
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
	return ioutil.WriteFile(ConfigFile, b, 0777)
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

func (c *Config) EditRoom(roomInfo RoomConfigInfo) {
	c.lock.Lock()
	defer c.lock.Unlock()
	_, ok := c.Conf.Live[roomInfo.RoomID]
	if ok {
		c.Conf.Live[roomInfo.RoomID] = roomInfo
	}
}

func (c *Config) AddMonitorArea(area MonitorArea) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Conf.MonitorAreas = append(c.Conf.MonitorAreas, area)
}

// 对于MonitorArea线程不安全
func (c *Config) DeleteMonitorArea(area MonitorArea) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for i, v := range c.Conf.MonitorAreas {
		if v.AreaID == area.AreaID {
			c.Conf.MonitorAreas = append(c.Conf.MonitorAreas[:i], c.Conf.MonitorAreas[i+1:]...)
			break
		}
	}
}

func (c *Config) AddBlockedRoom(roomID string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Conf.BlockedRooms = append(c.Conf.BlockedRooms, roomID)
}
