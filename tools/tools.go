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
	"strings"
	"syscall"
	"time"

	"github.com/kataras/golog"
)

const (
	contentType     = "Content-Type"
	contentTypeJSON = "application/json"
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
	setTime, _ := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprint(time.Now().Format("2006-01-02") + " " + t), loc)
	if setTime.After(timeNow) {
		setTime.AddDate(0,0,1)
	}
	timer := time.NewTimer(setTime.Sub(timeNow))
	<- timer.C
	c <- 1
	golog.Info("EveryDayTimer Work at ", time.Now().Format("2006-01-02 15:04:05"), " Next work time is ", setTime.AddDate(0,0,1).Format("2006-01-02 15:04:05"))
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