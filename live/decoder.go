package live

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/kataras/golog"

	"bilibili-recording-go/infos"
	"bilibili-recording-go/tools"
)

func (l *Live) decodeWorker() {
	for {
		roomID := <-l.decodeChannel
		if l.compareAndSwapUint32(roomID, waiting, decoding) {
			infs := infos.New()
			infs.RoomInfos[roomID].DecodeStartTime = time.Now().Unix()
			golog.Debug(fmt.Sprintf("%s[RoomID: %s] 开始转码", infs.RoomInfos[roomID].Uname, roomID))
			Decode(roomID)
			golog.Debug(fmt.Sprintf("%s[RoomID: %s] 结束转码", infs.RoomInfos[roomID].Uname, roomID))
			infs.RoomInfos[roomID].DecodeEndTime = time.Now().Unix()
			l.compareAndSwapUint32(roomID, decoding, decodeEnd)
			if infs.RoomInfos[roomID].AutoUpload {
				l.compareAndSwapUint32(roomID, decodeEnd, updateWait)
				l.uploadChannel <- roomID
			} else {
				l.compareAndSwapUint32(roomID, decodeEnd, start)
			}
		}
	}
}

// Decode 转码
func Decode(roomID string) {
	infs := infos.New()
	roomInfo := infs.RoomInfos[roomID]
	var fileLst []string
	var timeLst []int64
	tmpDir := fmt.Sprintf("./recording/%s/tmp", roomInfo.Uname)
	for _, f := range tools.ListDir(tmpDir) {
		if ok := strings.HasSuffix(f, ".flv"); ok {
			fileLst = append(fileLst, f)
		}
	}
	sort.Sort(sort.StringSlice(fileLst))
	for _, f := range fileLst {
		timeLst = append(timeLst, tools.GetFileLastModifyTime(f))
	}
	latestTime := timeLst[len(timeLst)-1]
	var inputFile []string
	for k, v := range timeLst {
		if tools.GetTimeDeltaFromTimestamp(latestTime, v) < 28800 {
			inputFile = append(inputFile, fileLst[k])
		}
	}
	fileTime := tools.GetFileCreateTime(inputFile[0])
	loc, _ := time.LoadLocation("PRC")
	tNow, _ := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprint(time.Now().Format("2006-01-02"), " ", "03:00:00"), loc)
	var ftime string
	if time.Unix(fileTime, 0).Before(tNow) {
		ftime = time.Now().AddDate(0, 0, -1).Format("20060102")
	} else {
		ftime = time.Now().Format("20060102")
	}
	uploadName := fmt.Sprintf("%s%s", roomInfo.Uname, ftime)
	outputName := fmt.Sprintf("%s_%s", roomInfo.Uname, ftime)
	pwd, _ := os.Getwd()
	outputFile := filepath.Join(pwd, "recording", roomInfo.Uname, fmt.Sprintf("%s.mp4", outputName))
	golog.Info(fmt.Sprintf("%s[RoomID: %s] 本次录制文件: %s, 最终上传: %s", roomInfo.Uname, roomInfo.UID, strings.Join(inputFile, " "), uploadName))
	var middleLst []string
	for _, f := range inputFile {
		middleLst = append(middleLst, strings.Replace(f, ".flv", ".ts", -1))
	}

	var middleToFileLst []string

	for _, f := range middleLst {
		absPath, _ := filepath.Abs(f)
		middleToFileLst = append(middleToFileLst, fmt.Sprintf("file '%s'", absPath))
	}

	concatFilePath, _ := filepath.Abs(fmt.Sprintf("./recording/%s/tmp/concat.txt", roomInfo.Uname))
	concatFile, _ := os.OpenFile(concatFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	writeString := strings.Join(middleToFileLst, "\n")
	io.WriteString(concatFile, writeString)

	for k := range inputFile {
		cmd := exec.Command("ffmpeg", "-i", inputFile[k], "-c", "copy", "-y", middleLst[k])
		// stdout, _ := cmd.StdoutPipe()
		// cmd.Stderr = cmd.Stdout
		cmd.Run()
		// tools.LiveOutput(stdout)
	}

	reg, _ := regexp.Compile("bitrate: (\\d+) kb/s")
	flag := false

	for _, f := range middleLst {
		getBitRateCmd := exec.Command("ffprobe", f)
		out, _ := getBitRateCmd.CombinedOutput()
		bitRate := reg.FindAllStringSubmatch(string(out), -1)
		if bitRate[0][1] > "3000" {
			flag = true
		}
	}

	if flag {
		cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", concatFilePath, "-c:v", "libx264", "-c:a", "copy", "-crf", "17", "-maxrate", "3M", "-bufsize", "3M", "-preset", "fast", "-y", outputFile)
		// stdout, _ := cmd.StdoutPipe()
		// cmd.Stderr = cmd.Stdout
		cmd.Run()
		// tools.LiveOutput(stdout)
	} else {
		cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", concatFilePath, "-c", "copy", "-y", outputFile)
		// stdout, _ := cmd.StdoutPipe()
		// cmd.Stderr = cmd.Stdout
		cmd.Run()
		// tools.LiveOutput(stdout)
	}

	cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", concatFilePath, "-acodec", "copy", "-vn", "-y", strings.Replace(outputFile, ".mp4", ".m4a", -1))
	// stdout, _ := cmd.StdoutPipe()
	// cmd.Stderr = cmd.Stdout
	cmd.Run()
	// tools.LiveOutput(stdout)

	for _, f := range middleLst {
		err := os.Remove(f)
		if err != nil {
			golog.Error(err)
		} else {
			golog.Info(f, "has been removed")
		}
	}

	golog.Info(fmt.Sprintf("%s[RoomID: %s] 转码完成", roomInfo.Uname, roomInfo.UID))
}
