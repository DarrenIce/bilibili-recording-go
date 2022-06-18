package danmu

import (
	"time"

	"github.com/gorilla/websocket"
)

type DanmuClient struct {
	roomID uint32
	conn   *websocket.Conn
	Ass    *Ass

	unzlibChannel       chan []byte
	heartBeatChannel    chan []byte
	serverNoticeChannel chan []byte
	Stop                chan struct{}
	stopReceive         chan struct{}
	heartTimer		  *time.Ticker
}

type receivedInfo struct {
	Cmd        string                 `json:"cmd"`
	Data       map[string]interface{} `json:"data"`
	Info       []interface{}          `json:"info"`
	Full       map[string]interface{} `json:"full"`
	Half       map[string]interface{} `json:"half"`
	Side       map[string]interface{} `json:"side"`
	RoomID     uint32                 `json:"roomid"`
	RealRoomID uint32                 `json:"real_roomid"`
	MsgCommon  string                 `json:"msg_common"`
	MsgSelf    string                 `json:"msg_self"`
	LinkUrl    string                 `json:"link_url"`
	MsgType    string                 `json:"msg_type"`
	ShieldUID  string                 `json:"shield_uid"`
	BusinessID string                 `json:"business_id"`
	Scatter    map[string]interface{} `json:"scatter"`
}

func NewDanmuClient(roomID uint32, uname string, file string) *DanmuClient {
	return &DanmuClient{
		roomID:              roomID,
		conn:                new(websocket.Conn),
		Ass:                 NewAss(uname, file),
		unzlibChannel:       make(chan []byte, 100),
		// heartBeatChannel:    make(chan []byte, 100),
		// serverNoticeChannel: make(chan []byte, 100),
		Stop:                make(chan struct{}),
		stopReceive:         make(chan struct{}),
	}
}
