package live

import (
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/asmcos/requests"
	"github.com/kataras/golog"
	"github.com/tidwall/gjson"

	"bilibili-recording-go/config"
	"bilibili-recording-go/tools"
)

func init() {
	registerSite("douyin", &douyin{})
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
	if s.cookies == "" {
		return SiteInfo{
			Title: "Cookie未添加",
		}
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
	resp, err := req.Get(fmt.Sprintf("https://live.douyin.com/%s", r.RoomID), headers)
	if err != nil {
		golog.Error(err)
		return SiteInfo{
			Title: err.Error(),
		}
	}
	splits := strings.Split(resp.Text(), `<script id="RENDER_DATA" type="application/json">`)
	if len(splits) < 2 {
		return SiteInfo{
			Title: "Fail to find url",
		}
	}
	resps := splits[1]
	resps = strings.Split(resps, `</script>`)[0]
	resps, err = url.QueryUnescape(resps)
	if err != nil {
		golog.Error(err)
		return SiteInfo{
			Title: err.Error(),
		}
	}

	data := gjson.Get(resps, "initialState.roomStore.roomInfo")
	siteInfo := SiteInfo{}
	status := int(data.Get("room.status").Int())
	if status == 2 {
		siteInfo.LiveStatus = 1
		s.liveUrl = data.Get("room.stream_url.flv_pull_url.FULL_HD1").String()
	} else if status == 4 {
		siteInfo.LiveStatus = 0
	} else {
		siteInfo.LiveStatus = status
	}
	siteInfo.RealID = data.Get("room.id_str").String()
	siteInfo.LockStatus = 0
	siteInfo.Uname = data.Get("anchor.nickname").String()
	siteInfo.UID = data.Get("anchor.id_str").String()
	siteInfo.Title = data.Get("room.title").String()
	siteInfo.LiveStartTime = time.Now().Unix()
	siteInfo.AreaName = data.Get("partition_road_map.partition.title").String()
	return siteInfo
}

func (s *douyin) DownloadLive(r *Live) {
	uname := r.Uname
	tools.Mkdir(fmt.Sprintf("./recording/%s/tmp", uname))
	exp := regexp.MustCompile(`[\/:*?"<>|]`)
	title := exp.ReplaceAllString(r.Title, " ")
	outputName := r.AreaName + "_" + title + "_" + fmt.Sprint(time.Now().Format("20060102150405")) + ".flv"
	golog.Info(fmt.Sprintf("%s[RoomID: %s] 本次录制文件为：%s", r.Uname, r.RoomID, outputName))
	middle, _ := filepath.Abs(fmt.Sprintf("./recording/%s/tmp", uname))
	outputFile := fmt.Sprint(middle + "\\" + outputName)
	r.downloadCmd = exec.Command("ffmpeg", "-i", s.liveUrl, "-c", "copy", outputFile)
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