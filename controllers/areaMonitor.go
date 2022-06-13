package controllers

import (
	"bilibili-recording-go/monitor"

	"github.com/gin-gonic/gin"
)

func GetAreaInfos(c *gin.Context) {
	c.JSON(200, monitor.AreaMonitorMap)
}