package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/kataras/golog"

	"bilibili-recording-go/config"
	"bilibili-recording-go/live"
	"bilibili-recording-go/server"
	"bilibili-recording-go/tools"
)

// Init 初始化函数
func Init() {
	tools.Mkdir("./log")
	logFile, err := os.OpenFile("./log/log.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		golog.Fatal(err)
	}
	golog.SetLevel("debug")
	golog.SetTimeFormat("2006/01/02 15:04:05")
	golog.Handle(func(l *golog.Log) bool {
		prefix := golog.GetTextForLevel(l.Level, false)
		file := "???"
		line := 0

		pc := make([]uintptr, 64)
		n := runtime.Callers(3, pc)
		if n != 0 {
			pc = pc[:n]
			frames := runtime.CallersFrames(pc)

			for {
				frame, more := frames.Next()
				if !strings.Contains(frame.File, "github.com/kataras/golog") {
					file = frame.File
					line = frame.Line
					break
				}
				if !more {
					break
				}
			}
		}

		slices := strings.Split(file, "/")
		file = slices[len(slices)-1]

		message := fmt.Sprintf("%s %s [%s:%d] %s",
			prefix, l.FormatTime(), file, line, l.Message)

		if l.NewLine {
			message += "\n"
		}

		// fixme https://stackoverflow.com/a/14694666
		logFile.WriteString(message)
		return true
	})
	golog.Debug("log file Init.")
	tools.Mkdir("./recording")
}

var (
	liver = live.Live{}
)

func main() {
	Init()
	liver := live.New()
	c := config.New()
	for _, v := range c.Conf.Live {
		liver.AddRoom(v)
	}
	server := server.New()
	server.Start()

	// time.Sleep(60 * time.Second)
	// liver.ManualUpload()

	upload2baidu := make(chan int)
	go tools.EveryDayTimer("05:00:00", upload2baidu)
	go func() {
		for {
			select {
			case <-upload2baidu:
				tools.Upload2BaiduPCS()
			default:
				continue
			}
		}
	}()
	for {
		continue
	}
}
