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
	for k, v := range monitor.AreaInfoList {
		if v.RoomID == roomID {
			monitor.AreaInfoList = append(monitor.AreaInfoList[:k], monitor.AreaInfoList[k+1:]...)
		}
	}
	return true
}
