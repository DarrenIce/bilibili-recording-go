package controllers

import (
	"fmt"
	"strings"

	"bilibili-recording-go/config"
	"bilibili-recording-go/live"

	"github.com/gin-gonic/gin"
	"github.com/kataras/golog"
)

type receiveInfo struct {
	Handle string   `json:"handle"`
	Data   roomInfo `json:"data"`
}

type roomInfo struct {
	Platform       string `json:"platform"`
	RoomID         string `json:"roomID"`
	RecordMode     bool   `json:"recordMode"`
	StartTime      string `json:"startTime"`
	EndTime        string `json:"endTime"`
	AutoRecord     bool   `json:"autoRecord"`
	AutoUpload     bool   `json:"autoUpload"`
	NeedM4a        bool   `json:"needM4a"`
	Mp4Compress    bool   `json:"mp4Compress"`
	DivideByTitle  bool   `json:"divideByTitle"`
	CleanUpRegular bool   `json:"cleanUpRegular"`
	SaveDuration   string `json:"saveDuration"`
	AreaLock       bool   `json:"areaLock"`
	AreaLimit      string `json:"areaLimit"`
}

func RoomHandle(c *gin.Context) {
	info := new(receiveInfo)
	if c.ShouldBind(&info) == nil {
		fmt.Println(info)
		if info.Handle == "add" {
			c.JSON(200, &struct {
				Msg bool `json:"msg"`
			}{
				addRoom(info.Data),
			})
		} else if info.Handle == "edit" {
			c.JSON(200, &struct {
				Msg bool `json:"msg"`
			}{
				editRoom(info.Data),
			})
		} else if info.Handle == "delete" {
			c.JSON(200, &struct {
				Msg bool `json:"msg"`
			}{
				deleteRoom(info.Data),
			})
		}
	}
}

func addRoom(info roomInfo) bool {
	c := config.New()
	c.AddRoom(config.RoomConfigInfo{
		Platform:       info.Platform,
		RoomID:         info.RoomID,
		StartTime:      formatTime(info.StartTime),
		EndTime:        formatTime(info.EndTime),
		AutoRecord:     info.AutoRecord,
		AutoUpload:     info.AutoUpload,
		RecordMode:     info.RecordMode,
		NeedM4a:        info.NeedM4a,
		Mp4Compress:    info.Mp4Compress,
		DivideByTitle:  info.DivideByTitle,
		CleanUpRegular: info.CleanUpRegular,
		SaveDuration:   info.SaveDuration,
		AreaLock:       info.AreaLock,
		AreaLimit:      info.AreaLimit,
	})
	live.AddRoom(info.RoomID)
	if err := c.Marshal(); err != nil {
		golog.Error(err)
		return false
	}
	return true
}

func editRoom(info roomInfo) bool {
	roominfo := config.RoomConfigInfo{
		Platform:       info.Platform,
		RoomID:         info.RoomID,
		StartTime:      formatTime(info.StartTime),
		EndTime:        formatTime(info.EndTime),
		AutoRecord:     info.AutoRecord,
		AutoUpload:     info.AutoUpload,
		RecordMode:     info.RecordMode,
		NeedM4a:        info.NeedM4a,
		Mp4Compress:    info.Mp4Compress,
		DivideByTitle:  info.DivideByTitle,
		CleanUpRegular: info.CleanUpRegular,
		SaveDuration:   info.SaveDuration,
		AreaLock:       info.AreaLock,
		AreaLimit:      info.AreaLimit,
	}
	c := config.New()
	c.EditRoom(roominfo)
	live.Lives[info.RoomID].UpadteFromConfig(roominfo)
	if err := c.Marshal(); err != nil {
		golog.Error(err)
		return false
	}
	return true
}

func deleteRoom(info roomInfo) bool {
	c := config.New()
	c.DeleteRoom(info.RoomID)
	live.DeleteRoom(info.RoomID)
	if err := c.Marshal(); err != nil {
		golog.Error(err)
		return false
	}
	return true
}

func formatTime(time string) string {
	return strings.Join(strings.Split(strings.Split(time, " ")[1], ":"), "")
}
