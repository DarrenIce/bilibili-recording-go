package live

import (
	"fmt"
	"os/exec"
	"sync"

	"bilibili-recording-go/config"
	"bilibili-recording-go/infos"

	"github.com/asmcos/requests"
	"github.com/kataras/golog"
	"github.com/tidwall/gjson"
)

// Live 主类
type Live struct {
	rooms map[string]config.RoomConfigInfo

	stop          		chan string
	recordChannel 		chan config.RoomConfigInfo
	unliveChannel 		chan string
	decodeChannel 		chan string
	uploadChannel 		chan string
	downloadCmds  		map[string]*exec.Cmd
	state         		sync.Map
	lock          		*sync.Mutex
}

const (
	iinit   uint32 = iota
	start          // 开始监听
	restart        // 因各种原因导致重新录制时的状态转移
	running        // 正在录制
	waiting        // 在unlive中从running转移到waiting，如果不在录制时间段内就跳到waiting
	decoding
	decodeEnd
	uploadWait
	uploading
	uploadEnd
	stop
	// 转码上传完成后，从waiting回到start
)

var (
	once sync.Once

	instance *Live
)

// New new
func New() *Live {
	once.Do(func() {
		instance = new(Live)
		instance.Init()
	})

	return instance
}

// Init Live
func (l *Live) Init() {
	l.rooms = make(map[string]config.RoomConfigInfo)
	l.stop = make(chan string)
	l.recordChannel = make(chan config.RoomConfigInfo)
	l.unliveChannel = make(chan string)
	l.decodeChannel = make(chan string)
	l.uploadChannel = make(chan string)
	l.lock = new(sync.Mutex)
	l.downloadCmds = make(map[string]*exec.Cmd)

	c := config.New()
	err := c.LoadConfig()
	if err != nil {
		golog.Fatal(err)
	}
	for _, v := range c.Conf.Live {
		l.state.Store(v.RoomID, iinit)
	}

	go l.recordWorker()
	go l.decodeWorker()
	go l.uploadWorker()
	go l.flushLiveStatus()
}

// AddRoom ADD
func (l *Live) AddRoom(info config.RoomConfigInfo) {
	infs := infos.New()
	infs.RoomInfos[info.RoomID] = new(infos.RoomInfo)
	l.recordChannel <- info
}

// DeleteRoom deleteroom
func (l *Live) DeleteRoom(roomID string) {
	l.Stop(roomID)
}

// InfoResponse response
type InfoResponse struct {
	RoomID     string
	RealID     string
	Uname      string
	Title      string
	LiveStatus int
	AutoRecord bool
	AutoUpload bool
}

// GetRoomInfoForResp get
func GetRoomInfoForResp(info config.RoomConfigInfo) (InfoResponse, error) {
	url := fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom?room_id=%s", info.RoomID)
	resp, err := requests.Get(url)
	if err != nil {
		return InfoResponse{}, err
	}
	data := gjson.Get(resp.Text(), "data")
	inf := InfoResponse{}
	inf.RoomID = info.RoomID
	inf.AutoRecord = info.AutoRecord
	inf.AutoUpload = info.AutoUpload
	inf.RealID = data.Get("room_info").Get("room_id").String()
	inf.LiveStatus = int(data.Get("room_info").Get("live_status").Int())
	inf.Uname = data.Get("anchor_info").Get("base_info").Get("uname").String()
	inf.Title = data.Get("room_info").Get("title").String()
	return inf, nil
}
