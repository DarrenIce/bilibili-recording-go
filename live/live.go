package live

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"bilibili-recording-go/config"
	"bilibili-recording-go/danmu"
	"bilibili-recording-go/tools"

	emoji "github.com/Andrew-M-C/go.emoji"
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

	UploadName  string
	FilePath    string
	TmpFilePath string

	danmuClient *danmu.DanmuClient
}

type LiveSnapshot struct {
	config.RoomConfigInfo
	State uint32
	SiteInfo

	UploadName  string
	FilePath    string
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
	Lives         map[string]*Live
	LmapLock      *sync.Mutex
	DecodeChan    chan *LiveSnapshot
	uploadChan    chan string
	PlatformRooms = make(map[string][]string)
	PRLock        = new(sync.Mutex)
)

func init() {
	DecodeChan = make(chan *LiveSnapshot, 100)
	uploadChan = make(chan string)
	Lives = make(map[string]*Live)
	LmapLock = new(sync.Mutex)
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
	if l.Name != "" {
		l.Uname = l.Name
	}

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
	if l.Uname == "" && siteInfo.Uname != "" {
		l.Uname = siteInfo.Uname
		tools.Mkdir(fmt.Sprintf("./recording/%s/brg", l.Uname))
		tools.Mkdir(fmt.Sprintf("./recording/%s/tmp", l.Uname))
		tools.Mkdir(fmt.Sprintf("./recording/%s/ass", l.Uname))
	}
	l.RealID = siteInfo.RealID
	l.LiveStatus = siteInfo.LiveStatus
	l.LockStatus = siteInfo.LockStatus
	l.UID = siteInfo.UID
	l.LiveStartTime = siteInfo.LiveStartTime
	siteInfo.Title = emoji.ReplaceAllEmojiFunc(siteInfo.Title, func(emoji string) string {
		return ""
	})
	exp := regexp.MustCompile(`[\/:*?"<>|]`)
	siteInfo.Title = exp.ReplaceAllString(siteInfo.Title, " ")
	siteInfo.AreaName = exp.ReplaceAllString(siteInfo.AreaName, " ")
	isTitleChanged := false
	isAreaChanged := false
	if l.Title != siteInfo.Title && l.Title != "" && siteInfo.Title != "" && siteInfo.Title != "bilibili主播的直播间" && !strings.Contains(siteInfo.Title, "暂不支持") {
		golog.Info(fmt.Sprintf("%s[RoomID: %s] 标题更换 %s -> %s", l.Uname, l.RoomID, l.Title, siteInfo.Title))
		l.Title = siteInfo.Title
		isTitleChanged = true
	}
	if siteInfo.Title != "" && l.Title == "" {
		l.Title = siteInfo.Title
	}
	if l.AreaName != siteInfo.AreaName && l.AreaName != "" && siteInfo.AreaName != "" {
		golog.Info(fmt.Sprintf("%s[RoomID: %s] 直播分区更换 %s -> %s", l.Uname, l.RoomID, l.AreaName, siteInfo.AreaName))
		l.AreaName = siteInfo.AreaName
		isAreaChanged = true
	}
	if siteInfo.AreaName != "" && l.AreaName == "" {
		l.AreaName = siteInfo.AreaName
	}
	if (isTitleChanged && l.DivideByTitle && l.State == running) || (isAreaChanged && !l.AreaLock && l.State == running && l.Platform != "douyin") {
		if l.downloadCmd.Process != nil {
			l.downloadCmd.Process.Kill()
		}
	}

}
