package infos

import (
	"sync"
	"time"

	"github.com/tidwall/gjson"

	"bilibili-recording-go/config"
)

type RoomInfo struct {
	RoomID     string
	StartTime  string
	EndTime    string
	AutoRecord bool
	AutoUpload bool

	RealID        string
	LiveStatus    int
	LockStatus    int
	Uname         string
	UID           string
	Title         string
	LiveStartTime int64
	AreaName	  string

	RecordStatus    int
	RecordStartTime int64
	RecordEndTime   int64
	DecodeStatus    int
	DecodeStartTime int64
	DecodeEndTime   int64
	UploadStatus    int
	UploadStartTime int64
	UploadEndTime   int64
	NeedUpload		bool
	St              time.Time
	Et              time.Time
	State           uint32

	UploadName		string
	FilePath		string
}

type biliInfo struct {
	Username string
	Password string
	Cookies  string
}

// LiveInfos liveinfos
type LiveInfos struct {
	BiliInfo  *biliInfo
	RoomInfos map[string]*RoomInfo

	lock *sync.Mutex
}

var (
	once sync.Once

	instance *LiveInfos
)

// New new
func New() *LiveInfos {
	once.Do(func() {
		instance = new(LiveInfos)
		instance.init()
	})

	return instance
}

func (l *LiveInfos) init() {
	c := config.New()
	l.BiliInfo = new(biliInfo)
	l.RoomInfos = make(map[string]*RoomInfo)
	l.lock = new(sync.Mutex)
	for _, v := range c.Conf.Live {
		l.RoomInfos[v.RoomID] = new(RoomInfo)
	}
}

// UpdateFromGJSON update
func (l *LiveInfos) UpdateFromGJSON(roomID string, res gjson.Result) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.RoomInfos[roomID].RealID = res.Get("room_info").Get("room_id").String()
	l.RoomInfos[roomID].LiveStatus = int(res.Get("room_info").Get("live_status").Int())
	l.RoomInfos[roomID].LockStatus = int(res.Get("room_info").Get("lock_status").Int())
	l.RoomInfos[roomID].Uname = res.Get("anchor_info").Get("base_info").Get("uname").String()
	l.RoomInfos[roomID].UID = res.Get("room_info").Get("uid").String()
	l.RoomInfos[roomID].Title = res.Get("room_info").Get("title").String()
	l.RoomInfos[roomID].LiveStartTime = res.Get("room_info").Get("live_start_time").Int()
	l.RoomInfos[roomID].AreaName = res.Get("room_info").Get("area_name").String()
}

// UpadteFromConfig update
func (l *LiveInfos) UpadteFromConfig(roomID string, v config.RoomConfigInfo) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.RoomInfos[roomID].RoomID = v.RoomID
	l.RoomInfos[roomID].StartTime = v.StartTime
	l.RoomInfos[roomID].EndTime = v.EndTime
	l.RoomInfos[roomID].AutoRecord = v.AutoRecord
	l.RoomInfos[roomID].AutoUpload = v.AutoUpload
}

// DeleteRoomInfo delete
func (l *LiveInfos) DeleteRoomInfo(roomID string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	_, ok := l.RoomInfos[roomID]
	if ok {
		delete(l.RoomInfos, roomID)
	}
}