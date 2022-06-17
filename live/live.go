package live

import (
	"fmt"
	"os/exec"
	"sync"
	"time"

	"bilibili-recording-go/config"
	"bilibili-recording-go/tools"

	"github.com/Andrew-M-C/go.emoji"
	"github.com/kataras/golog"
)

// Live 主类
type Live struct {
	config.RoomConfigInfo
	downloadCmd *exec.Cmd
	lock        *sync.Mutex
	State       uint32
	stop        chan struct{}
	site        Site

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
	TmpFilePath string
}

type LiveSnapshot struct {
	config.RoomConfigInfo
	State	uint32
	SiteInfo

	UploadName string
	FilePath  string
	TmpFilePath string
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
	DecodeChan chan *LiveSnapshot
	uploadChan chan string
)

func init() {
	DecodeChan = make(chan *LiveSnapshot, 100)
	uploadChan = make(chan string)
	Lives = make(map[string]*Live)
	LmapLock = new(sync.Mutex)

	c := config.New()
	err := c.LoadConfig()
	if err != nil {
		golog.Fatal(fmt.Sprintf("Load config error: %s", err))
	}
	for _, v := range c.Conf.Live {
		AddRoom(v.RoomID)
	}
	go flushLiveStatus()
	StartTimingTask("Upload2BaiduPCS", c.Conf.RcConfig.NeedBdPan, c.Conf.RcConfig.UploadTime, Upload2BaiduPCS)
	StartTimingTask("CleanRecordingDir", c.Conf.RcConfig.NeedRegularClean, c.Conf.RcConfig.RegularCleanTime, CleanRecordingDir)
	// go uploadWorker()
}

// Init Live
func (l *Live) Init(roomID string) {
	c := config.New()

	l.lock = new(sync.Mutex)
	l.downloadCmd = new(exec.Cmd)
	l.State = iinit
	l.stop = make(chan struct{})
	// 读的时候暂时没加锁
	l.RoomConfigInfo = c.Conf.Live[roomID]
	var ok bool
	l.site, ok = Sniff(l.Platform)
	if !ok {
		golog.Fatal(fmt.Sprintf("[%s] Platform %s hasn't been supported.", roomID, l.Platform))
	}
	if l.Platform == "douyin" {
		l.site.SetCookies(c.Conf.Douyin.Cookies)
	}
	l.St, l.Et = tools.MkDuration(l.StartTime, l.EndTime)

	if _, ok := c.Conf.Live[roomID]; !ok {
		golog.Error(fmt.Sprintf("Room %s Init ERROR", roomID))
	}
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
	siteInfo.Uname = emoji.ReplaceAllEmojiFunc(siteInfo.Uname, func(emoji string) string {
		return ""
	})
	if l.Uname == "" {
		l.Uname = siteInfo.Uname
	}
	l.RealID = siteInfo.RealID
	l.LiveStatus = siteInfo.LiveStatus
	l.LockStatus = siteInfo.LockStatus
	l.UID = siteInfo.UID
	l.LiveStartTime = siteInfo.LiveStartTime
	siteInfo.Title = emoji.ReplaceAllEmojiFunc(siteInfo.Title, func(emoji string) string {
		return ""
	})
	if l.Title != siteInfo.Title && l.Title != "" {
		golog.Info(fmt.Sprintf("%s[RoomID: %s] 标题更换 %s -> %s", l.Uname, l.RoomID, l.Title, siteInfo.Title))
		l.Title = siteInfo.Title
		if l.DivideByTitle && l.State == running {
			l.downloadCmd.Process.Kill()
		}
	}
	if siteInfo.Title != "" {
		l.Title = siteInfo.Title
	}
	if l.AreaName != siteInfo.AreaName && l.AreaName != ""{
		golog.Info(fmt.Sprintf("%s[RoomID: %s] 直播分区更换 %s -> %s", l.Uname, l.RoomID, l.AreaName, siteInfo.AreaName))
		l.AreaName = siteInfo.AreaName
		if !l.AreaLock && l.State == running {
			l.downloadCmd.Process.Kill()
		}
	}
	if siteInfo.AreaName != "" {
		l.AreaName = siteInfo.AreaName
	}
}
