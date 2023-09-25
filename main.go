package main

import (
	"bilibili-recording-go/config"
	_ "bilibili-recording-go/danmu"
	_ "bilibili-recording-go/decode"
	"bilibili-recording-go/live"
	_ "bilibili-recording-go/monitor"
	"bilibili-recording-go/routers"
	_ "bilibili-recording-go/tools"
	"time"

	"flag"
	"fmt"

	"github.com/kataras/golog"
)

var (
	port string
	configName string
)

func init() {
	flag.StringVar(&port, "p", "8000", "端口号")
	flag.StringVar(&configName, "c", "config.yml", "配置文件")
	flag.Parse()
	config.ConfigFile = fmt.Sprintf("./%s", configName)
	c := config.New()
	err := c.LoadConfig()
	if err != nil {
		golog.Fatal(fmt.Sprintf("Load config error: %s", err))
	}
	for _, v := range c.Conf.Live {
		live.AddRoom(v.RoomID)
		if v.Platform == "douyin" {
			time.Sleep(10 * time.Second)
		}
	}
	// go flushLiveStatus()
	live.StartTimingTask("Upload2BaiduPCS", c.Conf.RcConfig.NeedBdPan, c.Conf.RcConfig.UploadTime, live.Upload2BaiduPCS)
	live.StartTimingTask("CleanRecordingDir", c.Conf.RcConfig.NeedRegularClean, c.Conf.RcConfig.RegularCleanTime, live.CleanRecordingDir)
}

func main() {
	
	routers.GIN.Run(fmt.Sprintf(":%s", port))
}
