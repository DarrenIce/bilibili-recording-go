package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/beego/beego/v2/server/web/filter/cors"
	"github.com/kataras/golog"

	"bilibili-recording-go/config"
	"bilibili-recording-go/live"
	_ "bilibili-recording-go/routers"
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

func beegoInit() {
	// beego.SetViewsPath("static")
	beego.InsertFilter("/", beego.BeforeRouter, WebServerFilter)
	beego.InsertFilter("/*", beego.BeforeRouter, WebServerFilter)
	beego.InsertFilter("/*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))
}

//静态文件后缀
const staticFileExt = ".html|.js|.css|.png|.jpg|.jpeg|.ico|.otf"

//web 服务过滤器,实现静态文件发布
func WebServerFilter(ctx *context.Context) {
	urlPath := ctx.Request.URL.Path
	// golog.Info(urlPath)
	if urlPath == "" || urlPath == "/" {
		urlPath = "index.html"
	}
    
    
	ext := path.Ext(urlPath)
	if ext == "" {
		return
	}
	index := strings.Index(staticFileExt, ext)
	if index >= 0 {
		http.ServeFile(ctx.ResponseWriter, ctx.Request, "static/"+urlPath)
	}
}


func main() {
	beegoInit()
	Init()
	c := config.New()
	err := c.LoadConfig()
	if err != nil {
		golog.Fatal(err)
	}
	for _, v := range c.Conf.Live {
		live.AddRoom(v.RoomID)
	}
	server := server.New()
	server.Start()

	// time.Sleep(60 * time.Second)
	// liver.ManualUpload()
	if c.Conf.RcConfig.NeedBdPan {
		upload2baidu := make(chan int)
		go tools.EveryDayTimer(c.Conf.RcConfig.UploadTime, upload2baidu)
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
	}
	
	beego.Run()
}
