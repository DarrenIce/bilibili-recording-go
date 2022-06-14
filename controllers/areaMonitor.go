package controllers

import (
	"bilibili-recording-go/config"
	"bilibili-recording-go/monitor"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kataras/golog"
)

type receiveAreaInfo struct {
	Handle string   `json:"handle"`
	Data   config.MonitorArea `json:"data"`
}

func GetAreaInfos(c *gin.Context) {
	c.JSON(200, monitor.AreaMonitorMap)
}

func AreaHandle(c *gin.Context) {
	info := new(receiveAreaInfo)
	if c.ShouldBind(&info) == nil {
		fmt.Println(info)
		if info.Handle == "add" {
			c.JSON(200, &struct {
				Msg bool `json:"msg"`
			}{
				addArea(info.Data),
			})
		} else if info.Handle == "delete" {
			c.JSON(200, &struct {
				Msg bool `json:"msg"`
			}{
				deleteArea(info.Data),
			})
		} else {
			c.JSON(400, &struct {
				Msg bool `json:"msg"`
			}{
				false,
			})
		}
	}
}

func addArea(info config.MonitorArea) bool {
	c := config.New()
	c.AddMonitorArea(info)
	if err := c.Marshal(); err != nil {
		golog.Error(err)
		return false
	}
	return true
}

func deleteArea(info config.MonitorArea) bool {
	c := config.New()
	c.DeleteMonitorArea(info)
	if err := c.Marshal(); err != nil {
		golog.Error(err)
		return false
	}
	return true
}