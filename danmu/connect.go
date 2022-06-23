package danmu

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/asmcos/requests"
	"github.com/gorilla/websocket"
	"github.com/kataras/golog"
	"github.com/tidwall/gjson"
)

var (
	getDanmuInfo = "https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?id=%d&type=0"
)

type handShakeInfo struct {
	UID       uint8  `json:"uid"`
	Roomid    uint32 `json:"roomid"`
	Protover  uint8  `json:"protover"`
	Platform  string `json:"platform"`
	Clientver string `json:"clientver"`
	Type      uint8  `json:"type"`
	Key       string `json:"key"`
}

func (d *DanmuClient) connect() {
	r, err := requests.Get(fmt.Sprintf(getDanmuInfo, d.roomID))
	if err != nil {
		fmt.Println("request.Get DanmuInfo: ", err)
	}
	fmt.Println("获取弹幕服务器")
	token := gjson.Get(r.Text(), "data.token").String()
	hostList := []string{}
	gjson.Get(r.Text(), "data.host_list").ForEach(func(key, value gjson.Result) bool {
		hostList = append(hostList, value.Get("host").String())
		return true
	})
	hsInfo := handShakeInfo{
		UID:       0,
		Roomid:    d.roomID,
		Protover:  2,
		Platform:  "web",
		Clientver: "1.10.2",
		Type:      2,
		Key:       token,
	}
	for _, h := range hostList {
		d.conn, _, err = websocket.DefaultDialer.Dial(fmt.Sprintf("wss://%s:443/sub", h), nil)
		if err != nil {
			fmt.Println("websocket.Dial: ", err)
			continue
		}
		fmt.Printf("连接弹幕服务器[%s]成功\n", hostList[0])
		break
	}
	if err != nil {
		fmt.Println("websocket.Dial Error")
	}
	jm, err := json.Marshal(hsInfo)
	if err != nil {
		fmt.Println("json.Marshal: ", err)
	}
	err = d.sendPackage(0, 16, 1, 7, 1, jm)
	if err != nil {
		fmt.Println("Conn SendPackage: ", err)
	}
	fmt.Printf("连接房间[%d]成功\n", d.roomID)
}

func (d *DanmuClient) heartBeat() {
	d.heartTimer = time.NewTicker(time.Second * 30)
	obj := []byte("5b6f626a656374204f626a6563745d")
	d.sendPackage(0, 16, 0, 2, 1, obj)
	for {
		_, ok := <-d.heartTimer.C
		if !ok {
			fmt.Printf("[%d] heartTimer stop\n", d.roomID)
			return
		}
		obj := []byte("5b6f626a656374204f626a6563745d")
		if err := d.sendPackage(0, 16, 0, 2, 1, obj); err != nil {
			golog.Error(fmt.Sprintf("[%d]heart beat err: %s, try to reconnect.", d.roomID, err))
			golog.Error(fmt.Sprintf("[%d]now ass file is %s", d.roomID, d.Ass.File))
			d.connect()
		}
	}
}
func (d *DanmuClient) receiveRawMsg() {
	for {
		select {
		case <-d.stopReceive:
			return
		default:
			_, msg, err := d.conn.ReadMessage()
			if err != nil {
				fmt.Printf("[%d]%s Ws Receive raw msg error: %s\n", d.roomID, time.Now().Format("2006-01-02 15:04:05"), err)
				return
			}
			if len(msg) == 0 {
				continue
			}
			if msg[7] == 2 {
				// fmt.Printf("[%d]UnZlib..\n", d.roomID)
				msgs := splitMsg(zlibUnCompress(msg[16:]))
				for _, m := range msgs {
					// Normal danmu
					d.unzlibChannel <- m
				}
			}
			//  else if msg[11] == 3 {
			// 	// HeartBeat
			// 	d.heartBeatChannel <- msg
			// } else {
			// 	// System
			// 	d.serverNoticeChannel <- msg
			// }
		}
	}
}

func (d *DanmuClient) Run() {
	d.connect()
	go d.process()
	go d.heartBeat()
	go d.receiveRawMsg()
	for {
		select {
		case <-d.Stop:
			d.stopReceive <- struct{}{}
			d.heartTimer.Stop()
			d.conn.Close()
			close(d.unzlibChannel)
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
