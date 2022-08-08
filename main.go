package main

import (
	_ "bilibili-recording-go/tools"
	_ "bilibili-recording-go/danmu"
	_ "bilibili-recording-go/decode"
	_ "bilibili-recording-go/live"
	_ "bilibili-recording-go/monitor"
	"bilibili-recording-go/routers"
	
)

func main() {
	routers.GIN.Run()
}
