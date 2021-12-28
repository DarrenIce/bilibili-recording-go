package tools

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/asmcos/requests"
	"github.com/kataras/golog"
	"github.com/tidwall/gjson"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"

	"bilibili-recording-go/config"
)

const (
	contentType     = "Content-Type"
	contentTypeJSON = "application/json"
)

var (
	lastBytesSent uint64 = 0
	lastBytesRecv uint64 = 0
)

// Exists 检测文件或文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// Mkdir 新建文件夹
func Mkdir(path string) {
	if !Exists(path) {
		err := os.MkdirAll(path, os.ModeDir)
		if err != nil {
			golog.Fatal(err)
		}
	} else {
		golog.Info(path + "已存在")
	}
}

func writeMsg(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(msg))
}

// WriteJSON write to resp
func WriteJSON(w http.ResponseWriter, obj interface{}) {
	b, err := json.Marshal(obj)
	if err != nil {
		writeMsg(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set(contentType, contentTypeJSON)
	_, _ = w.Write(b)
}

// MkDuration make duration
func MkDuration(startTime string, endTime string) (time.Time, time.Time) {
	t := time.Now()
	tt := t.Format("20060102 150405")
	head := strings.Split(tt, " ")[0]
	loc, _ := time.LoadLocation("PRC")
	if startTime == endTime && startTime == "0" {
		st, _ := time.ParseInLocation("20060102 150405", fmt.Sprint(head, " 000000"), loc)
		et, _ := time.ParseInLocation("20060102 150405", fmt.Sprint(head, " 235959"), loc)
		return st, et
	} else if startTime < endTime {
		st, _ := time.ParseInLocation("20060102 150405", fmt.Sprint(head, " ", startTime), loc)
		et, _ := time.ParseInLocation("20060102 150405", fmt.Sprint(head, " ", endTime), loc)
		return st, et
	} else {
		tmp, _ := time.ParseInLocation("20060102 150405", fmt.Sprint(head, " ", startTime), loc)
		var baseTime1 string
		var baseTime2 string
		if t.After(tmp) {
			baseTime1 = head
			baseTime2 = t.AddDate(0, 0, 1).Format("20060102")
		} else {
			baseTime1 = t.AddDate(0, 0, -1).Format("20060102")
			baseTime2 = head
		}
		st, _ := time.ParseInLocation("20060102 150405", fmt.Sprint(baseTime1, " ", startTime), loc)
		et, _ := time.ParseInLocation("20060102 150405", fmt.Sprint(baseTime2, " ", endTime), loc)
		return st, et
	}
}

// JudgeInDuration judge duration
func JudgeInDuration(startTime time.Time, endTime time.Time) bool {
	t := time.Now()
	if t.After(startTime) && t.Before(endTime) {
		return true
	}
	return false
}

// ListDir listdir
func ListDir(dirPath string) (files []string) {
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		golog.Fatal(err)
		return nil
	}
	for _, file := range dir {
		if file.IsDir() {
			continue
		} else {
			files = append(files, dirPath+"/"+file.Name())
		}
	}
	return files
}

// GetFileCreateTime get
func GetFileCreateTime(filePath string) int64 {
	fileInfo, _ := os.Stat(filePath)
	fileSys := fileInfo.Sys().(*syscall.Win32FileAttributeData)
	nanoseconds := fileSys.CreationTime.Nanoseconds()
	createTime := nanoseconds / 1e9
	return createTime
}

// GetFileLastModifyTime get
func GetFileLastModifyTime(filePath string) int64 {
	fileInfo, _ := os.Stat(filePath)
	fileSys := fileInfo.Sys().(*syscall.Win32FileAttributeData)
	nanoseconds := fileSys.LastWriteTime.Nanoseconds()
	lastWriteTime := nanoseconds / 1e9
	return lastWriteTime
}

// GetTimeDeltaFromTimestamp time1 - time2
func GetTimeDeltaFromTimestamp(time1 int64, time2 int64) int {
	t1 := time.Unix(time1, 0)
	t2 := time.Unix(time2, 0)
	return int(t1.Sub(t2).Seconds())
}

// LiveOutput output for exec
func LiveOutput(out io.ReadCloser) {
	for {
		tmp := make([]byte, 1024)
		_, err := out.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}
}

// EveryDayTimer timer
func EveryDayTimer(t string, c chan int) {
	golog.Info("EveryDayTimer Start, set time as ", t)
	timeNow := time.Now()
	loc, _ := time.LoadLocation("PRC")
	setTime, _ := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprint(time.Now().Format("2006-01-02")+" "+t), loc)
	if setTime.Before(timeNow) {
		setTime = setTime.AddDate(0, 0, 1)
	}
	timer := time.NewTimer(setTime.Sub(timeNow))
	<-timer.C
	c <- 1
	golog.Info("EveryDayTimer Work at ", time.Now().Format("2006-01-02 15:04:05"), " Next work time is ", setTime.AddDate(0, 0, 1).Format("2006-01-02 15:04:05"))
	var ticker *time.Ticker = time.NewTicker(24 * time.Hour)
	ticks := ticker.C
	for range ticks {
		c <- 1
		golog.Info("EveryDayTimer Work at ", time.Now().Format("2006-01-02 15:04:05"), "Next work time is ", time.Now().Format("2006-01-02 15:04:05"))
	}
}

func md5V(t string) string {
	h := md5.New()
	h.Write([]byte(t))
	return hex.EncodeToString(h.Sum(nil))
}

func Upload2BaiduPCS() {
	c := config.New()
	for _, v := range c.Conf.Live {
		uname, err := GetUname(v.RoomID)
		if err != nil {
			continue
		}
		localBasePath := fmt.Sprint("./recording/", uname)
		if !Exists(localBasePath) {
			continue
		}
		pcsBasePath := fmt.Sprint("/录播/", uname)
		cmd := exec.Command("./BaiduPCS-Go.exe", "mkdir", pcsBasePath)
		stdout, _ := cmd.StdoutPipe()
		cmd.Stderr = cmd.Stdout
		cmd.Start()
		LiveOutput(stdout)
		for _, f := range ListDir(localBasePath) {
			if o, _ := os.Stat(f); !o.IsDir() {
				// s1 := strings.Split(f, "\\")
				// filename := s1[len(s1)-1]
				// s2 := strings.Split(filename, "_")
				// d := s2[len(s2)-1]
				// s3 := strings.Split(d, ".")[0]
				// fmt.Println(s3)
				// if time.Now().AddDate(0, 0, -1).Format("20060102") == s3 {
				cmd = exec.Command("./BaiduPCS-Go.exe", "upload", f, pcsBasePath)
				stdout, _ := cmd.StdoutPipe()
				cmd.Stderr = cmd.Stdout
				cmd.Start()
				LiveOutput(stdout)
				// }
			}
		}
		cmd = exec.Command("./BaiduPCS-Go.exe", "export", pcsBasePath, "--link")
		stdout, _ = cmd.StdoutPipe()
		cmd.Stderr = cmd.Stdout
		cmd.Start()
		LiveOutput(stdout)
	}
}

// GetUname, 感觉可以加个缓存，不然频繁调用容易触发风控
func GetUname(roomID string) (string, error) {
	url := fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom?room_id=%s", roomID)
	resp, err := requests.Get(url)
	if err != nil {
		return "", err
	}
	data := gjson.Get(resp.Text(), "data")
	uname := data.Get("anchor_info").Get("base_info").Get("uname").String()
	time.Sleep(3 * time.Second)
	return uname, nil
}

func DirSize(dirPath string, dirSize int64) int64 {
	flist, err := ioutil.ReadDir(dirPath)
	if err != nil {
		golog.Error(err)
		return 0
	}
	for _, f := range flist {
		if f.IsDir() {
			dirSize = DirSize(dirPath+"/"+f.Name(), dirSize)
		} else {
			dirSize = f.Size() + dirSize
		}
	}
	return dirSize
}

func CacRecordingFileNum() int64 {
	fileNum := 0
	flist, err := ioutil.ReadDir("./recording/")
	if err != nil {
		golog.Error(err)
		return 0
	}
	for _, d := range flist {
		if d.IsDir() {
			localBasePath := fmt.Sprint("./recording/", d.Name(), "/tmp")
			if !Exists(localBasePath) {
				continue
			}
			for _, f := range ListDir(localBasePath) {
				if strings.HasSuffix(f, "flv") {
					fileNum++
				}
			}
		}
	}
	return int64(fileNum)
}

type DeviceInfo struct {
	TotalCpuUsage	float64
	PerCpuUsage		[]float64
	MemUsage		uint64
	MemTotal		uint64
	DiskName		string
	DiskUsage		uint64
	DiskTotal		uint64
	NetUpPerSec		uint64
	NetDownPerSec	uint64
}

func GetDeviceInfo() (deviceInfo DeviceInfo) {
	percent, _ := cpu.Percent(time.Second, true)
	deviceInfo.PerCpuUsage = percent
	percent, _ = cpu.Percent(time.Second, false)
	deviceInfo.TotalCpuUsage = percent[0]
	mem, _ := mem.VirtualMemory()
	deviceInfo.MemUsage = mem.Used
	deviceInfo.MemTotal = mem.Total
	nett, _ := net.IOCounters(false)
	deviceInfo.NetUpPerSec = nett[0].BytesSent - lastBytesSent
	lastBytesSent = nett[0].BytesSent
	deviceInfo.NetDownPerSec = nett[0].BytesRecv - lastBytesRecv
	lastBytesRecv = nett[0].BytesRecv
	pwd, _ := os.Getwd()
	disk, _ := disk.Usage(fmt.Sprint(strings.Split(pwd, "\\")[0], "\\"))
	deviceInfo.DiskName = strings.Split(pwd, "\\")[0]
	deviceInfo.DiskUsage = disk.Used
	deviceInfo.DiskTotal = disk.Total
	return deviceInfo
}

func BytesToStringFast(b []byte) string {
    return *(*string)(unsafe.Pointer(&b))
}