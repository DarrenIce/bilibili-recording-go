package danmu

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/issue9/term/v2/colors"
	"github.com/kataras/golog"
)

func (d *DanmuClient) process() {
	for {
		select {
		case m, ok := <- d.unzlibChannel:
			if !ok {
				return
			}
			uz := m[16:]
			js := new(receivedInfo)
			json.Unmarshal(uz, js)
			switch js.Cmd {
			case "DANMU_MSG":
				d.DanmuMsg(uz)
			}
		}
	}
}

func (d *DanmuClient) DanmuMsg(bs []byte) {
	js := new(receivedInfo)
	if err := json.Unmarshal(bs, js); err != nil {
		golog.Error(fmt.Sprintf("json.Unmarshal: %s", err))
	}
	info := js.Info
	ditem := DanmuItem{}
	if len(info) > 0 {
		ditem.msg = info[1].(string)
	}
	if len(info) > 1 {
		i := info[2].([]interface{})
		if len(i) > 0 {
			ditem.uid = i[0].(float64)
		}
		if len(i) > 1 {
			ditem.uname = i[1]
		}
	}
	dm := colors.New(colors.Normal, 158, colors.Black)
	dm.Printf("[%s] %s: [%s] %s. \n", time.Now().Format("2006-01-02 15:04:05"), js.Cmd, ditem.uname, ditem.msg)
	d.Ass.WriteDanmu(ditem.msg)
}