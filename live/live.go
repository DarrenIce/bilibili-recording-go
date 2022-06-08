package live

import (
	"fmt"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"bilibili-recording-go/config"
	"bilibili-recording-go/tools"

	"github.com/asmcos/requests"
	"github.com/kataras/golog"
	"github.com/tidwall/gjson"
)

// Live 主类
type Live struct {
	config.RoomConfigInfo
	downloadCmd *exec.Cmd
	lock        *sync.Mutex
	State       uint32
	stop        chan struct{}
	site		Site

	SiteInfo

	RecordStatus    int
	RecordStartTime int64
	RecordEndTime   int64
	DecodeStatus    int
	DecodeStartTime int64
	DecodeEndTime   int64
	UploadStatus    int
	UploadStartTime int64
	UploadEndTime   int64
	NeedUpload      bool
	St              time.Time
	Et              time.Time

	UploadName string
	FilePath   string
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
	Lives    map[string]*Live
	LmapLock *sync.Mutex

	decodeChan chan string
	uploadChan chan string
)

func init() {
	decodeChan = make(chan string)
	uploadChan = make(chan string)
	Lives = make(map[string]*Live)

	LmapLock = new(sync.Mutex)
	go flushLiveStatus()
	go uploadWorker()
	go decodeWorker()
}

// Init Live
func (l *Live) Init(roomID string) {
	l.lock = new(sync.Mutex)
	l.downloadCmd = new(exec.Cmd)
	l.State = iinit
	l.stop = make(chan struct{})
	var ok bool
	l.site, ok = Sniff(l.Platform)
	if !ok {
		golog.Fatal(fmt.Sprintf("[%s] Platform %s hasn't been supported.", roomID, l.Platform))
	}

	c := config.New()

	if _, ok := c.Conf.Live[roomID]; !ok {
		golog.Error(fmt.Sprintf("Room %s Init ERROR", roomID))
	}
	// 读的时候暂时没加锁
	l.RoomConfigInfo = c.Conf.Live[roomID]
	l.St, l.Et = tools.MkDuration(l.StartTime, l.EndTime)
}

// AddRoom ADD
func AddRoom(roomID string) {
	live := new(Live)
	live.Init((roomID))
	LmapLock.Lock()
	Lives[roomID] = live
	LmapLock.Unlock()
	go live.start()
}

// DeleteRoom deleteroom
func DeleteRoom(roomID string) {
	LmapLock.Lock()
	defer LmapLock.Unlock()
	live := new(Live)
	if _, ok := Lives[roomID]; ok {
		live = Lives[roomID]
		delete(Lives, roomID)
	}
	live.stop <- struct{}{}
	// 如何释放？
	// State == start || State == restart
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

func ManualUpload(roomID string) bool {
	LmapLock.Lock()
	defer LmapLock.Unlock()
	if _, ok := Lives[roomID]; ok {
		if atomic.CompareAndSwapUint32(&Lives[roomID].State, start, uploadWait) || Lives[roomID].State == uploadWait {
			uploadChan <- roomID
			return true
		}
	}
	return false
}

func ManualDecode(roomID string) bool {
	LmapLock.Lock()
	defer LmapLock.Unlock()
	if _, ok := Lives[roomID]; ok {
		if atomic.CompareAndSwapUint32(&Lives[roomID].State, start, waiting) {
			decodeChan <- roomID
			return true
		}
	}
	return false
}

// UpadteFromConfig update
func (l *Live) UpadteFromConfig(v config.RoomConfigInfo) {
	l.lock.Lock()
	LmapLock.Lock()
	defer LmapLock.Unlock()
	defer l.lock.Unlock()
	dividebeforestaus := l.DivideByTitle
	l.RoomConfigInfo = v
	if dividebeforestaus != l.DivideByTitle && l.State == running {
		l.downloadCmd.Process.Kill()
	}
}

func (l *Live) UpdateSiteInfo() {
	siteInfo := l.site.GetInfoByRoom(l)
	if l.Uname == "" {
		l.Uname = siteInfo.Uname
	}
	l.RealID = siteInfo.RealID
	l.LiveStatus = siteInfo.LiveStatus
	l.LockStatus = siteInfo.LockStatus
	l.UID = siteInfo.UID
	l.LiveStartTime = siteInfo.LiveStartTime
	if l.Title != siteInfo.Title && l.Title != "" {
		golog.Info(fmt.Sprintf("%s[RoomID: %s] 标题更换 %s -> %s", l.Uname, l.RoomID, l.Title, siteInfo.Title))
		l.Title = siteInfo.Title
		if l.DivideByTitle && l.State == running {
			l.downloadCmd.Process.Kill()
		}
	}
	l.Title = siteInfo.Title
	if l.AreaName != siteInfo.AreaName {
		golog.Info(fmt.Sprintf("%s[RoomID: %s] 直播分区更换 %s -> %s", l.Uname, l.RoomID, l.AreaName, siteInfo.AreaName))
		if !l.AreaLock && l.State == running {
			l.downloadCmd.Process.Kill()
		}
	}
}

func flushLiveStatus() {
	for {
		lst := make([]string, len(Lives))
		LmapLock.Lock()
		for k := range Lives {
			lst = append(lst, k)
		}
		LmapLock.Unlock()
		for _, v := range lst {
			LmapLock.Lock()
			live, ok := Lives[v]
			if !ok {
				LmapLock.Unlock()
				continue
			}
			LmapLock.Unlock()
			live.UpdateSiteInfo()
			time.Sleep(10 * time.Second)
		}
	}
}
