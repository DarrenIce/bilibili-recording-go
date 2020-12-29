package config

import (
	"io/ioutil"

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
type Config struct {
	Bili	bili `yaml:"bilibili"`
	Live	map[string]RoomConfigInfo `yaml:"live"`
}

var defaultconfig Config

// LoadConfig 加载配置文件
func LoadConfig() (*Config, error) {
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