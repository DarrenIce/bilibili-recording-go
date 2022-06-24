package danmu

import (
	"encoding/json"
	"fmt"
	"strings"

	// "github.com/issue9/term/v2/colors"
	"github.com/kataras/golog"
)

func (d *DanmuClient) process() {
	for {
		m, ok := <-d.unzlibChannel
		if !ok {
			fmt.Printf("[%d] unzlibChannel stop\n", d.roomID)
			return
		}
		uz := m[16:]
		js := new(receivedInfo)
		json.Unmarshal(uz, js)
		switch js.Cmd {
		case "DANMU_MSG":
			d.DanmuMsg(uz)
		case "GUARD_BUY":
			d.GuardBuy(uz)
		case "SEND_GIFT":
			d.SendGift(uz)
		case "SUPER_CHAT_MESSAGE":
			d.SuperChatMessage(uz)
		}
	}
}

func (d *DanmuClient) DanmuMsg(bs []byte) {
	defer func() {
		if v := recover(); v != nil {
			golog.Error("捕获了一个恐慌：", v)
		}
	}()
	js := new(receivedInfo)
	if err := json.Unmarshal(bs, js); err != nil {
		golog.Error(fmt.Sprintf("json.Unmarshal: %s", err))
	}
	info := js.Info
	ditem := DanmuItem{}
	if len(info) > 0 {
		ditem.msg = info[1].(string)
	}
	d.Ass.WriteDanmu(ditem.msg)
	if len(info) > 9 {
		cmd := js.Cmd
		uid := int64(info[2].([]interface{})[0].(float64))
		uname := info[2].([]interface{})[1].(string)
		msg := strings.ReplaceAll(info[1].(string), ",", " ")
		medal_info := info[3].([]interface{})
		medal_name := ""
		medal_level := 0
		medal_anchor := ""
		if len(medal_info) > 0 {
			medal_name = info[3].([]interface{})[1].(string)
			medal_level = int(info[3].([]interface{})[0].(float64))
			medal_anchor = info[3].([]interface{})[2].(string)
		}
		user_info := info[4].([]interface{})
		user_level := 0
		if len(user_info) > 0 {
			user_level = int(info[4].([]interface{})[0].(float64))
		}
		timestamp := int64(info[9].(map[string]interface{})["ts"].(float64))
		d.Brg.WriteMsg(fmt.Sprintf("%s, %d, %s, %d, %s, %d, %s, %s, %d\n", cmd, uid, uname, user_level, medal_name, medal_level, medal_anchor, msg, timestamp))
	}
}

func (d *DanmuClient) GuardBuy(bs []byte) {
	defer func() {
		if v := recover(); v != nil {
			golog.Error("捕获了一个恐慌：", v)
		}
	}()
	js := new(receivedInfo)
	if err := json.Unmarshal(bs, js); err != nil {
		golog.Error(fmt.Sprintf("json.Unmarshal: %s", err))
	}
	data := js.Data
	cmd := js.Cmd
	uid := int64(data["uid"].(float64))
	uname := data["username"].(string)
	guard_level := int(data["guard_level"].(float64))
	gift_name := data["gift_name"].(string)
	gift_price := int64(data["price"].(float64))
	gift_num := int(data["num"].(float64))
	timestamp := int64(data["start_time"].(float64))
	d.Brg.WriteMsg(fmt.Sprintf("%s, %d, %s, %d, %s, %d, %d, %d\n", cmd, uid, uname, guard_level, gift_name, gift_price, gift_num, timestamp))
}

func (d *DanmuClient) SendGift(bs []byte) {
	defer func() {
		if v := recover(); v != nil {
			golog.Error("捕获了一个恐慌：", v)
		}
	}()
	js := new(receivedInfo)
	if err := json.Unmarshal(bs, js); err != nil {
		golog.Error(fmt.Sprintf("json.Unmarshal: %s", err))
	}
	data := js.Data
	cmd := js.Cmd
	uid := int64(data["uid"].(float64))
	uname := data["uname"].(string)
	medal_name := data["medal_info"].(map[string]interface{})["medal_name"].(string)
	medal_level := int(data["medal_info"].(map[string]interface{})["medal_level"].(float64))
	medal_anchor := data["medal_info"].(map[string]interface{})["anchor_uname"].(string)
	gift_name := data["giftName"].(string)
	gift_price := int64(data["price"].(float64))
	coin_type := data["coin_type"].(string)
	if coin_type == "silver" {
		return
	}
	gift_num := int(data["num"].(float64))
	timestamp := int64(data["timestamp"].(float64))
	d.Brg.WriteMsg(fmt.Sprintf("%s, %d, %s, %s, %d, %s, %s, %d, %s, %d, %d\n", cmd, uid, uname, medal_name, medal_level, medal_anchor, gift_name, gift_price, coin_type, gift_num, timestamp))
}

func (d *DanmuClient) SuperChatMessage(bs []byte) {
	defer func() {
		if v := recover(); v != nil {
			golog.Error("捕获了一个恐慌：", v)
		}
	}()
	js := new(receivedInfo)
	if err := json.Unmarshal(bs, js); err != nil {
		golog.Error(fmt.Sprintf("json.Unmarshal: %s", err))
	}
	data := js.Data
	cmd := js.Cmd
	uid := int64(data["uid"].(float64))
	uname := data["user_info"].(map[string]interface{})["uname"].(string)
	user_level := int(data["user_info"].(map[string]interface{})["user_level"].(float64))
	medal_name := data["medal_info"].(map[string]interface{})["medal_name"].(string)
	medal_level := int(data["medal_info"].(map[string]interface{})["medal_level"].(float64))
	medal_anchor := data["medal_info"].(map[string]interface{})["anchor_uname"].(string)
	sc_price := int64(data["price"].(float64)) * 1000
	sc_time := int(data["time"].(float64))
	msg := data["message"].(string)
	timestamp := int64(data["ts"].(float64))
	d.Brg.WriteMsg(fmt.Sprintf("%s, %d, %s, %d, %s, %d, %s, %d, %d, %s, %d\n", cmd, uid, uname, user_level, medal_name, medal_level, medal_anchor, sc_price, sc_time, msg, timestamp))
}