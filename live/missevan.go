package live

import (
	"bilibili-recording-go/config"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/asmcos/requests"
	"github.com/kataras/golog"
	"github.com/tidwall/gjson"
)

func init() {
	registerSite("missevan", &missevan{})
	PRLock.Lock()
	PlatformRooms["missevan"] = make([]string, 0)
	PRLock.Unlock()
	go flushPlatformLives("missevan")
}

type missevan struct {
	liveUrl string
}

func (s *missevan) Name() string {
	return "猫耳"
}

func (s *missevan) SetCookies(cookies string) {}

func (s *missevan) GetInfoByRoom(r *Live) SiteInfo {
	req := requests.Requests()
	c := config.New()
	if c.Conf.RcConfig.NeedProxy {
		req.Proxy(c.Conf.RcConfig.Proxy)
	}
	headers := requests.Header{
		"User-Agent":   "Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko); Chrome/75.0.3770.100 Mobile Safari/537.36",
		"Content-Type": "application/x-www-form-urlencoded",
	}
	resp, err := req.Get(fmt.Sprintf("https://fm.missevan.com/api/v2/live/%s", r.RoomID), headers)
	if err != nil {
		return SiteInfo{
			Title: err.Error(),
		}
	}
	data := resp.Text()
	sInfo := SiteInfo{}
	sInfo.LiveStatus = int(gjson.Get(data, "info.room.status.open").Int())
	if (sInfo.LiveStatus == 1) {
		s.liveUrl = gjson.Get(data, "info.room.channel.flv_pull_url").String()
		sInfo.LiveStartTime = gjson.Get(data, "info.room.status.open_time").Int() / 1000
	}
	sInfo.RealID = gjson.Get(data, "info.room.room_id").String()
	sInfo.LockStatus = 0
	sInfo.Uname = gjson.Get(data, "info.room.creator_username").String()
	sInfo.UID = gjson.Get(data, "info.room.creator_id").String()
	sInfo.Title = gjson.Get(data, "info.room.name").String()
	sInfo.AreaName = ""

	return sInfo
}

func (s *missevan) GetRoomLiveURL(roomID string) (string, bool) {
	return "", false
}

func (s *missevan) DownloadLive(r *Live) {
	uname := r.Uname
	outputName := r.AreaName + "_" + r.Title + "_" + fmt.Sprint(time.Unix(r.RecordStartTime, 0).Format("20060102150405")) + ".m4a"
	golog.Info(fmt.Sprintf("%s[RoomID: %s] 本次录制文件为：%s", r.Uname, r.RoomID, outputName))
	r.TmpFilePath = fmt.Sprintf("./recording/%s/tmp/%s", uname, outputName)
	middle, _ := filepath.Abs(fmt.Sprintf("./recording/%s/tmp", uname))
	outputFile := fmt.Sprint(middle + "\\" + outputName)
	r.downloadCmd = exec.Command("ffmpeg", "-i", s.liveUrl, "-c", "copy", outputFile)
	// r.downloadCmd = exec.Command("streamlink", "-f", "-o", outputFile, s.liveUrl, "best")
	// stdout, _ := r.downloadCmd.StdoutPipe()
	// r.downloadCmd.Stderr = r.downloadCmd.Stdout
	if err := r.downloadCmd.Start(); err != nil {
		golog.Error(err)
		r.downloadCmd.Process.Kill()
	}
	// tools.LiveOutput(stdout)
	r.downloadCmd.Wait()
	r.RecordEndTime = time.Now().Unix()
	golog.Info(fmt.Sprintf("%s[RoomID: %s] 录制结束", r.Uname, r.RoomID))
	r.unlive()
}
