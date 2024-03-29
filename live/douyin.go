package live

import (
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/asmcos/requests"
	"github.com/kataras/golog"
	"github.com/tidwall/gjson"
	_ "github.com/u2takey/ffmpeg-go"

	"bilibili-recording-go/config"
)

func init() {
	registerSite("douyin", &douyin{})
	PRLock.Lock()
	PlatformRooms["douyin"] = make([]string, 0)
	PRLock.Unlock()
	go flushPlatformLives("douyin")
}

type douyin struct {
	cookies string
	liveUrl string
}

func (s *douyin) Name() string {
	return "抖音"
}

func (s *douyin) SetCookies(cookies string) {
	s.cookies = cookies
}

func (s *douyin) GetInfoByRoom(r *Live) SiteInfo {
	if r.Cookies == "" && s.cookies == "" {
		return SiteInfo{
			Title: "Cookie未添加",
		}
	}
	req := requests.Requests()
	c := config.New()
	if c.Conf.RcConfig.NeedProxy {
		req.Proxy(c.Conf.RcConfig.Proxy)
	}
	cookies := ""
	if r.Cookies != "" {
		cookies = r.Cookies
	} else {
		cookies = s.cookies
	}
	headers := requests.Header{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36",
		"referer":    "https://live.douyin.com/",
		"cookie":     cookies,
	}
	resp, err := req.Get(fmt.Sprintf("https://live.douyin.com/webcast/room/web/enter/?web_rid=%s&aid=6383&device_platform=web&browser_language=zh-CN'&browser_platform=Win32&browser_name=Chrome&browser_version=92.0.4515.159", r.RoomID), headers)
	if err != nil {
		golog.Error(err)
		return SiteInfo{}
	}
	// splits := strings.Split(resp.Text(), `<script id="RENDER_DATA" type="application/json">`)
	// if len(splits) < 2 {
	// 	return SiteInfo{}
	// }
	// resps := splits[1]
	// resps = strings.Split(resps, `</script>`)[0]
	// resps, err = url.QueryUnescape(resps)
	// if err != nil {
	// 	golog.Error(err)
	// 	return SiteInfo{}
	// }
	data := gjson.Parse(resp.Text())
	siteInfo := SiteInfo{}
	status := int(data.Get("data.data.0.status").Int())
	if status == 2 {
		siteInfo.LiveStatus = 1
		s.liveUrl = data.Get("data.data.0.stream_url.flv_pull_url.FULL_HD1").String()
	} else if status == 4 {
		siteInfo.LiveStatus = 0
	} else {
		siteInfo.LiveStatus = status
	}
	siteInfo.RealID = data.Get("data.data.0.id_str").String()
	siteInfo.LockStatus = 0
	siteInfo.Uname = data.Get("data.user.nickname").String()
	siteInfo.UID = data.Get("data.user.id_str").String()
	siteInfo.Title = data.Get("data.data.0.title").String()
	siteInfo.LiveStartTime = time.Now().Unix()
	siteInfo.AreaName = data.Get("data.partition_road_map.partition.title").String()
	return siteInfo
}

func (s *douyin) GetRoomLiveURL(roomID string) (string, bool) {
	if s.cookies == "" {
		return "", false
	}
	req := requests.Requests()
	c := config.New()
	if c.Conf.RcConfig.NeedProxy {
		req.Proxy(c.Conf.RcConfig.Proxy)
	}
	headers := requests.Header{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.71 Safari/537.36 Edg/94.0.992.38",
		"referer":    "https://live.douyin.com/",
		"Content-Type": "application/form-data",
		"cookie":     s.cookies,
	}
	resp, err := req.Get(fmt.Sprintf("https://live.douyin.com/%s", roomID), headers)
	if err != nil {
		golog.Error(err)
		return "", false
	}
	splits := strings.Split(resp.Text(), `<script id="RENDER_DATA" type="application/json">`)
	if len(splits) < 2 {
		return "", false
	}
	resps := splits[1]
	resps = strings.Split(resps, `</script>`)[0]
	resps, err = url.QueryUnescape(resps)
	if err != nil {
		golog.Error(err)
		return "", false
	}

	data := gjson.Get(resps, "initialState.roomStore.roomInfo")
	status := int(data.Get("room.status").Int())
	if status == 2 {
		return data.Get("room.stream_url.flv_pull_url.FULL_HD1").String(), true
	} else {
		return "", false
	}
}

func (s *douyin) DownloadLive(r *Live) {
	uname := r.Uname
	outputName := r.AreaName + "_" + r.Title + "_" + fmt.Sprint(time.Unix(r.RecordStartTime, 0).Format("20060102150405")) + ".flv"
	golog.Info(fmt.Sprintf("%s[RoomID: %s] 本次录制文件为：%s", r.Uname, r.RoomID, outputName))
	r.TmpFilePath = fmt.Sprintf("./recording/%s/tmp/%s", uname, outputName)
	middle, _ := filepath.Abs(fmt.Sprintf("./recording/%s/tmp", uname))
	outputFile := fmt.Sprint(middle + "\\" + outputName)
	r.downloadCmd = exec.Command("ffmpeg", "-rw_timeout", "10000000", "-i", s.liveUrl, "-c", "copy", outputFile)
	// ffmpeg_go.Input(s.liveUrl).Output(outputFile, ffmpeg_go.KwArgs{"c": "copy"}).OverWriteOutput().ErrorToStdOut().Run()
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
	time.Sleep(120 * time.Second)
	r.UpdateSiteInfo()
	r.unlive()
}