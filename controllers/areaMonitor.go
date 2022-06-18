package controllers

import (
	"bilibili-recording-go/config"
	"bilibili-recording-go/monitor"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kataras/golog"
)

func GetAreaInfos(c *gin.Context) {
	c.JSON(200, monitor.AreaInfoList)
}

func AreaHandle(c *gin.Context) {
	var info map[string]interface{}
	if c.ShouldBind(&info) == nil {
		fmt.Println("areaHandle")
		data := config.MonitorArea{
			Platform: info["data"].(map[string]interface{})["platform"].(string),
			AreaName: info["data"].(map[string]interface{})["areaName"].(string),
			AreaID: info["data"].(map[string]interface{})["areaID"].(string),
			ParentID: info["data"].(map[string]interface{})["parentID"].(string),
		}
		if info["handle"] == "add" {
			c.JSON(200, &struct {
				Msg bool `json:"msg"`
			}{
				addArea(data),
			})
		} else if info["handle"] == "delete" {
			c.JSON(200, &struct {
				Msg bool `json:"msg"`
			}{
				deleteArea(data),
			})
		} else {
			c.JSON(400, &struct {
				Msg bool `json:"msg"`
			}{
				false,
			})
		}
	} else {
		c.JSON(400, &struct {
			Msg bool `json:"msg"`
		}{
			false,
		})
	}
	
}

func addArea(info config.MonitorArea) bool {
	c := config.New()
	for _, v := range c.Conf.MonitorAreas {
		if v.AreaID == info.AreaID {
			return false
		}
	}
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