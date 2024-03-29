package live

import (
	"bilibili-recording-go/config"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/asmcos/requests"
	"github.com/kataras/golog"
	"github.com/tidwall/gjson"
)

func init() {
	registerSite("huya", &huya{})
	PRLock.Lock()
	PlatformRooms["huya"] = make([]string, 0)
	PRLock.Unlock()
	go flushPlatformLives("huya")
}

type huya struct {
	liveUrl string
}

func (s *huya) Name() string {
	return "虎牙"
}

func (s *huya) SetCookies(cookies string) {}

func (s *huya) GetInfoByRoom(r *Live) SiteInfo {
	req := requests.Requests()
	c := config.New()
	if c.Conf.RcConfig.NeedProxy {
		req.Proxy(c.Conf.RcConfig.Proxy)
	}
	headers := requests.Header{
		"User-Agent":   "Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko); Chrome/75.0.3770.100 Mobile Safari/537.36",
		"Content-Type": "application/x-www-form-urlencoded",
	}
	resp, err := req.Get(fmt.Sprintf("https://m.huya.com/%s", r.RoomID), headers)
	if err != nil {
		return SiteInfo{
			Title: err.Error(),
		}
	}
	re := regexp.MustCompile(`window.HNF_GLOBAL_INIT = ({.*})\s+<\/script>`)
	data := re.FindStringSubmatch(resp.Text())
	if len(data) < 2 {
		return SiteInfo{
			Title: "暂不支持",
		}
	}
	sInfo := SiteInfo{}
	sInfo.LiveStatus = int(gjson.Get(data[1], "roomInfo.eLiveStatus").Int()) - 1
	if sInfo.LiveStatus == 0 {
		sInfo.LiveStartTime = 0
	} else if sInfo.LiveStatus == -1 {
		sInfo.LockStatus = 1
	} else {
		sInfo.LiveStartTime = gjson.Get(data[1], "roomInfo.tLiveInfo.iStartTime").Int()
		liveUrl, _ := base64.RawStdEncoding.DecodeString(gjson.Get(data[1], "roomProfile.liveLineUrl").String())
		s.liveUrl = string(liveUrl)
		s.getLiveUrl(s.liveUrl)
	}
	sInfo.RealID = gjson.Get(data[1], "roomInfo.tProfileInfo.lProfileRoom").String()
	sInfo.LockStatus = 0
	sInfo.Uname = gjson.Get(data[1], "roomInfo.tProfileInfo.sNick").String()
	sInfo.UID = gjson.Get(data[1], "roomInfo.tProfileInfo.lUid").String()
	sInfo.Title = gjson.Get(data[1], "roomInfo.tLiveInfo.sRoomName").String()
	if sInfo.Title == "" {
		sInfo.Title = gjson.Get(data[1], "roomInfo.tLiveInfo.sIntroduction").String()
	}
	sInfo.AreaName = gjson.Get(data[1], "roomInfo.tLiveInfo.sGameFullName").String()

	return sInfo
}

func (s *huya) getLiveUrl(liveUrl string) (string, bool) {
	ib := strings.Split(liveUrl, "?")
	if len(ib) < 2 {
		fmt.Println("huya getLiveUrl failed.")
		return "", false
	}
	i, b := ib[0], ib[1]
	r := strings.Split(i, "/")
	ss := strings.ReplaceAll(r[len(r)-1], ".flv", "")
	ss = strings.ReplaceAll(ss, ".m3u8", "")
	c := strings.SplitN(b, "&", 4)
	y := c[len(c)-1]
	n := make(map[string]string)
	for _, v := range c {
		if v == "" {
			continue
		}
		vs := strings.SplitN(v, "=", 2)
		n[vs[0]] = vs[1]
	}
	fm := url.PathEscape(n["fm"])
	ub, _ := base64.RawStdEncoding.DecodeString(fm)
	u := string(ub)
	p := strings.Split(u, "_")[0]
	f := strconv.FormatInt(time.Now().UnixNano()/100, 10)
	l := n["wsTime"]
	t := "0"
	h := strings.Join([]string{p, t, ss, f, l}, "_")
	m := md5.New()
	io.WriteString(m, h)
	url := fmt.Sprintf("%s?wsSecret=%x&wsTime=%s&u=%s&seqid=%s&%s", i, m.Sum(nil), l, t, f, y)
	url = "https:" + url
	url = strings.ReplaceAll(url, "hls", "flv")
	url = strings.ReplaceAll(url, "m3u8", "flv")
	s.liveUrl = url
	return url, true
}

func (s *huya) GetRoomLiveURL(roomID string) (string, bool) {
	req := requests.Requests()
	c := config.New()
	if c.Conf.RcConfig.NeedProxy {
		req.Proxy(c.Conf.RcConfig.Proxy)
	}
	headers := requests.Header{
		"User-Agent":   "Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko); Chrome/75.0.3770.100 Mobile Safari/537.36",
		"Content-Type": "application/x-www-form-urlencoded",
	}
	resp, err := req.Get(fmt.Sprintf("https://m.huya.com/%s", roomID), headers)
	if err != nil {
		return "", false
	}
	re := regexp.MustCompile(`window.HNF_GLOBAL_INIT = ({.*})\s+<\/script>`)
	data := re.FindStringSubmatch(resp.Text())
	if len(data) < 2 {
		return "", false
	}
	status := int(gjson.Get(data[1], "roomInfo.eLiveStatus").Int()) - 1
	if status == 0 {
		return "", false
	} else {
		liveUrl, _ := base64.RawStdEncoding.DecodeString(gjson.Get(data[1], "roomProfile.liveLineUrl").String())
		return s.getLiveUrl(string(liveUrl))
	}
}

func (s *huya) DownloadLive(r *Live) {
	uname := r.Uname
	outputName := r.AreaName + "_" + r.Title + "_" + fmt.Sprint(time.Unix(r.RecordStartTime, 0).Format("20060102150405")) + ".flv"
	golog.Info(fmt.Sprintf("%s[RoomID: %s] 本次录制文件为：%s", r.Uname, r.RoomID, outputName))
	r.TmpFilePath = fmt.Sprintf("./recording/%s/tmp/%s", uname, outputName)
	middle, _ := filepath.Abs(fmt.Sprintf("./recording/%s/tmp", uname))
	outputFile := fmt.Sprint(middle + "\\" + outputName)
	url := fmt.Sprintf("http://127.0.0.1:1769/huya/%s", r.RoomID)
	r.downloadCmd = exec.Command("ffmpeg", "-i", url, "-c", "copy", outputFile)
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
	time.Sleep(time.Second * 120)
	r.unlive()
}
