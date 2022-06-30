package controllers

import (
	"bilibili-recording-go/liveback"

	"github.com/gin-gonic/gin"
)

func GetAnchorLivebackList(c *gin.Context) {
	name := c.Param("name")
	c.JSON(200, liveback.GetAnchorLivebackList(name))
}

type receiveInfo3 struct {
	AnchorName string `json:"anchor_name"`
	BrgFileName string `json:"brg_file_name"`
}

func GetLivebackStatistics(c *gin.Context) {
	info := receiveInfo3{}
	if c.ShouldBind(&info) == nil {
		if l, ok := liveback.GetLivebackStatistics(info.AnchorName, info.BrgFileName); ok {
			c.JSON(200, l)
		} else {
			c.JSON(400, nil)
		}
	}
}

func GetWordCloud(c *gin.Context) {
	info := receiveInfo3{}
	if c.ShouldBind(&info) == nil {
		if l, ok := liveback.CreateWordClod(info.AnchorName, info.BrgFileName); ok {
			c.JSON(200, l)
		} else {
			c.JSON(400, nil)
		}
	}
}