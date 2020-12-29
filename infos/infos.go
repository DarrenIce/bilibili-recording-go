package infos

import (
	"sync"
	"time"

	"github.com/tidwall/gjson"

	"bilibili-recording-go/config"
)

type roomInfo struct {
	RoomID			string
	StartTime		string
	EndTime			string
	AutoRecord		bool
	AutoUpload		bool

	RealID			string
	LiveStatus		int
	LockStatus		int
	Uname			string
	UID				string
	Title			string
	LiveStartTime	string

	RecordStatus	int
	RecordStartTime	string
	RecordEndTime	string
	DecodeStatus	int
	DecodeStartTime	string
	DecodeEndTime	string
	UploadStatus	int
	UploadStartTime	string
	UploadEndTime	string
	St				time.Time
	Et				time.Time
	State			uint32
}

type biliInfo struct {
	Username	string
	Password	string
	Cookies		string

}

type liveInfos struct {
	BiliInfo	*biliInfo
	RoomInfos	map[string]*roomInfo
}

var (
    once sync.Once

    instance *liveInfos
)

func New() *liveInfos {
	once.Do(func() {
		instance = &liveInfos{BiliInfo: &biliInfo{}, RoomInfos: make(map[string]*roomInfo)}
		instance.Init()
	})

	return instance
}

func (l *liveInfos) Init() {
	c := config.InitConfig()
	config, _ := c.LoadConfig()
	for _, v := range config.Live {
		l.RoomInfos[v.RoomID] = &roomInfo{}
	}
}

func (l *liveInfos) UpdateFromGJSON(roomID string, res gjson.Result) {
	l.RoomInfos[roomID].RealID = res.Get("room_info").Get("room_id").String()
	l.RoomInfos[roomID].LiveStatus = int(res.Get("room_info").Get("live_status").Int())
	l.RoomInfos[roomID].LockStatus = int(res.Get("room_info").Get("lock_status").Int())
	l.RoomInfos[roomID].Uname = res.Get("anchor_info").Get("base_info").Get("uname").String()
	l.RoomInfos[roomID].UID = res.Get("room_info").Get("uid").String()
	l.RoomInfos[roomID].Title = res.Get("room_info").Get("title").String()
	l.RoomInfos[roomID].LiveStartTime = res.Get("room_info").Get("live_start_time").String()
}

func (l *liveInfos) UpadteFromConfig(roomID string, v config.RoomConfigInfo) {
	l.RoomInfos[roomID].RoomID = v.RoomID
	l.RoomInfos[roomID].StartTime = v.StartTime
	l.RoomInfos[roomID].EndTime = v.EndTime
	l.RoomInfos[roomID].AutoRecord = v.AutoRecord
	l.RoomInfos[roomID].AutoUpload = v.AutoUpload
}