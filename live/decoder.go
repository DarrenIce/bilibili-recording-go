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
	"sync/atomic"
	"time"
	"bytes"

	"github.com/kataras/golog"

	"bilibili-recording-go/tools"
)

func decodeWorker() {
	for {
		roomID := <-decodeChan
		if _, ok := Lives[roomID]; !ok {
			continue
		}
		LmapLock.Lock()
		live := Lives[roomID]
		LmapLock.Unlock()
		if atomic.CompareAndSwapUint32(&live.State, waiting, decoding) {
			live.DecodeStartTime = time.Now().Unix()
			golog.Info(fmt.Sprintf("%s[RoomID: %s] 开始转码", live.Uname, roomID))
			live.Decode()
			golog.Info(fmt.Sprintf("%s[RoomID: %s] 结束转码", live.Uname, roomID))
			live.DecodeEndTime = time.Now().Unix()
			atomic.CompareAndSwapUint32(&live.State, decoding, decodeEnd)
			if live.AutoUpload {
				atomic.CompareAndSwapUint32(&live.State, decodeEnd, uploadWait)
				uploadChan <- roomID
			} else {
				atomic.CompareAndSwapUint32(&live.State, decodeEnd, start)
			}
		}
	}
}

type fileInfo struct {
	fileName string
	lastModifyTime int64
}

// Decode 转码
func (l *Live) Decode() {
	var fileLst []fileInfo
	tmpDir := fmt.Sprintf("./recording/%s/tmp", l.Uname)
	for _, f := range tools.ListDir(tmpDir) {
		if ok := strings.HasSuffix(f, ".flv"); ok {
			fileLst = append(fileLst, fileInfo{fileName: f, lastModifyTime: tools.GetFileLastModifyTime(f)})
		}
	}
	sort.Slice(fileLst, func(i, j int) bool { return fileLst[i].lastModifyTime < fileLst[j].lastModifyTime })
	latestTime := fileLst[len(fileLst)-1].lastModifyTime
	var inputFile []string
	if l.RecordMode || l.DivideByTitle {
		inputFile = append(inputFile, fileLst[len(fileLst)-1].fileName)
	} else {
		for k, v := range fileLst {
			if tools.GetTimeDeltaFromTimestamp(latestTime, v.lastModifyTime) < tools.GetTimeDeltaFromTimestamp(l.Et.Unix(), l.St.Unix()) {
				inputFile = append(inputFile, fileLst[k].fileName)
			}
		}
	}
	fileTime := tools.GetFileCreateTime(inputFile[0])
	loc, _ := time.LoadLocation("PRC")
	tNow, _ := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprint(time.Unix(fileTime, 0).Format("2006-01-02"), " ", "06:00:00"), loc)
	var ftime string
	if time.Unix(fileTime, 0).Before(tNow) {
		ftime = tNow.AddDate(0, 0, -1).Format("20060102")
	} else {
		ftime = tNow.Format("20060102")
	}
	if l.RecordMode {
		ftime = fmt.Sprintf("%s场", time.Unix(fileTime, 0).Format("2006-01-02 15时04分"))
	}
	if l.DivideByTitle {
		filesplit := strings.Split(inputFile[0], "/")
		titleWithTsp := strings.TrimSuffix(filesplit[len(filesplit)-1], ".flv")
		titleSplits := strings.Split(titleWithTsp, "_")
		title := strings.Join(titleSplits[:len(titleSplits)-1], "_")
		ftime = fmt.Sprintf("%s场_%s", time.Unix(fileTime, 0).Format("2006-01-02 15时04分"), title)
	}
	uploadName := fmt.Sprintf("%s%s", l.Uname, ftime)
	outputName := fmt.Sprintf("%s_%s", l.Uname, ftime)
	pwd, _ := os.Getwd()
	outputFile := filepath.Join(pwd, "recording", l.Uname, fmt.Sprintf("%s.mp4", outputName))
	l.UploadName = uploadName
	l.FilePath = outputFile
	golog.Info(fmt.Sprintf("%s[RoomID: %s] 本次转码的文件有: %s, 最终生成: %s", l.Uname, l.UID, strings.Join(inputFile, " "), outputFile))
	var middleLst []string
	for k, f := range inputFile {
		inputFile[k], _ = filepath.Abs(f)
		middleLst = append(middleLst, strings.Replace(inputFile[k], ".flv", ".ts", -1))
	}

	if tools.Exists(outputFile) {
		golog.Info(fmt.Sprintf("%s[RoomID: %s] 输出文件已存在，合并到新视频中", l.Uname, l.UID))
		middleLst = append(middleLst, strings.Replace(outputFile, ".mp4", ".ts", -1))
		cmd := exec.Command("ffmpeg", "-i", outputFile, "-vcodec", "copy", "-acodec", "copy", "-vbsf", "h264_mp4toannexb", "-y", strings.Replace(outputFile, ".mp4", ".ts", -1))
		fmt.Println(cmd.String())
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		}
	}

	var middleToFileLst []string

	for _, f := range middleLst {
		// absPath, _ := filepath.Abs(f)
		middleToFileLst = append(middleToFileLst, fmt.Sprintf("file '%s'", f))
	}

	concatFilePath, _ := filepath.Abs(fmt.Sprintf("./recording/%s/tmp/concat.txt", l.Uname))
	if tools.Exists(concatFilePath) {
		os.Remove(concatFilePath)
	}
	concatFile, _ := os.OpenFile(concatFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	writeString := strings.Join(middleToFileLst, "\n")
	io.WriteString(concatFile, writeString)

	for k := range inputFile {
		cmd := exec.Command("ffmpeg", "-fflags", "+discardcorrupt", "-i", inputFile[k], "-c", "copy", "-y", middleLst[k])
		fmt.Println(cmd.String())
		// stdout, _ := cmd.StdoutPipe()
		// cmd.Stderr = cmd.Stdout
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		}
		// tools.LiveOutput(stdout)
	}

	reg, _ := regexp.Compile(`bitrate: (\d+) kb/s`)
	flag := false

	for _, f := range middleLst {
		getBitRateCmd := exec.Command("ffprobe", f)
		out, _ := getBitRateCmd.CombinedOutput()
		bitRate := reg.FindAllStringSubmatch(string(out), -1)
		if bitRate[0][1] > "3000" {
			flag = true
		}
	}

	if flag && l.Mp4Compress {
		cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", concatFilePath, "-vcodec", "hevc_nvenc", "-c:a", "copy", "-crf", "17", "-maxrate", "3M", "-bufsize", "3M", "-preset", "fast", "-y", outputFile)
		fmt.Println(cmd.String())
		// stdout, _ := cmd.StdoutPipe()
		// cmd.Stderr = cmd.Stdout
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		}
		// tools.LiveOutput(stdout)
	} else {
		cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", concatFilePath, "-c", "copy", "-y", outputFile)
		fmt.Println(cmd.String())
		// stdout, _ := cmd.StdoutPipe()
		// cmd.Stderr = cmd.Stdout
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		}
		// tools.LiveOutput(stdout)
	}

	if l.NeedM4a {
		cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", concatFilePath, "-acodec", "copy", "-vn", "-y", strings.Replace(outputFile, ".mp4", ".m4a", -1))
		fmt.Println(cmd.String())
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		}
		// stdout, _ := cmd.StdoutPipe()
		// cmd.Stderr = cmd.Stdout
		// tools.LiveOutput(stdout)
	}
	
	for _, f := range middleLst {
		err := os.Remove(f)
		if err != nil {
			golog.Error(err)
		} else {
			golog.Info(f, " has been removed")
		}
	}

	golog.Info(fmt.Sprintf("%s[RoomID: %s] 转码完成", l.Uname, l.UID))
}
