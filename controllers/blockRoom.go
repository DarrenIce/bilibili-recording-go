package controllers

import (
	"fmt"

	"bilibili-recording-go/config"
	"bilibili-recording-go/monitor"

	"github.com/gin-gonic/gin"
	"github.com/kataras/golog"
)

func ProcessBlockRoom(c *gin.Context) {
	info := receiveInfo2{}
	if c.ShouldBind(&info) == nil {
		fmt.Println(info)
		c.JSON(200, &struct {
			Msg bool `json:"msg"`
		}{
			addBlockRoom(info.RoomID),
		})
	}
}

func addBlockRoom(roomID string) bool {
	c := config.New()
	for _, v := range c.Conf.BlockedRooms {
		if v == roomID {
			return false
		}
	}
	c.AddBlockedRoom(roomID)
	if err := c.Marshal(); err != nil {
		golog.Error(err)
		return false
	}
	monitor.Lock.Lock()
	defer monitor.Lock.Unlock()
	for k := range monitor.AreaMonitorMap {
		for kk, v := range monitor.AreaMonitorMap[k].Data {
			if v.RoomID == roomID {
				newData := append(monitor.AreaMonitorMap[k].Data[:kk], monitor.AreaMonitorMap[k].Data[kk+1:]...)
				numNum := monitor.AreaMonitorMap[k].Nums - 1
				newArea := &monitor.AreaList{
					Data: newData,
					Nums: numNum,
				}
				monitor.AreaMonitorMap[k] = *newArea
			}
		}
	}
	return true
}
