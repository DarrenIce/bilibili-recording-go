package decode

import (
	"bytes"
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

	"bilibili-recording-go/live"
	"bilibili-recording-go/tools"
)

func init() {
	go decodeWorker()
}

func decodeWorker() {
	golog.Info("Goroutine DecodeWorker start")
	for {
		live, ok := <-live.DecodeChan
		if ok {
			fmt.Println(live)
		}
		golog.Info(fmt.Sprintf("%s[RoomID: %s] 开始转码", live.Uname, live.RoomID))
		Decode(live)
		golog.Info(fmt.Sprintf("%s[RoomID: %s] 结束转码", live.Uname, live.RoomID))
	}
}

type fileInfo struct {
	fileName       string
	lastModifyTime int64
}

// Decode 转码
func Decode(l *live.LiveSnapshot) {
	var inputFile []string
	if l.TmpFilePath == "" {
		inputFile = GetLatestFiles(l, 0)
	} else {
		inputFile = []string{l.TmpFilePath}
	}
	uploadName, outputName := GenerateFileName(inputFile, l)
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
	ConvertFlv2Ts(middleLst, outputFile, inputFile, l)
	concatFilePath, _ := filepath.Abs(fmt.Sprintf("./recording/%s/tmp/concat.txt", l.Uname))
	ConvertTs2Mp4(middleLst, outputFile, l)

	if l.NeedM4a {
		cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", concatFilePath, "-acodec", "copy", "-vn", "-y", strings.Replace(outputFile, ".mp4", ".m4a", -1))
		fmt.Println(cmd.String())
		golog.Info(cmd.String())
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

func GetLatestFiles(l *live.LiveSnapshot, timeStamp int) []string {
	var fileLst []fileInfo
	tmpDir := fmt.Sprintf("./recording/%s/tmp", l.Uname)
	for _, f := range tools.ListDir(tmpDir) {
		if ok := strings.HasSuffix(f, ".flv"); ok {
			fileLst = append(fileLst, fileInfo{fileName: f, lastModifyTime: tools.GetFileLastModifyTime(f)})
		}
	}
	sort.Slice(fileLst, func(i, j int) bool { return fileLst[i].lastModifyTime < fileLst[j].lastModifyTime })
	if timeStamp == 0 {
		return []string{fileLst[len(fileLst)-1].fileName}
	} else {
		var files []string
		latestTime := fileLst[len(fileLst)-1].lastModifyTime
		for k, v := range fileLst {
			if tools.GetTimeDeltaFromTimestamp(latestTime, v.lastModifyTime) < timeStamp {
				files = append(files, fileLst[k].fileName)
			}
		}
		return files
	}
}

func GenerateFileName(inputFile []string, l *live.LiveSnapshot) (string, string) {
	fileTime := tools.GetFileCreateTime(inputFile[0])
	filesplit := strings.Split(inputFile[0], "/")
	titleWithTsp := strings.TrimSuffix(filesplit[len(filesplit)-1], ".flv")
	titleSplits := strings.Split(titleWithTsp, "_")
	areaName := titleSplits[0]
	title := strings.Join(titleSplits[1:len(titleSplits)-1], "_")
	ftime := fmt.Sprintf("%s场_%s_%s", time.Unix(fileTime, 0).Format("2006-01-02 15时04分"), areaName, title)
	uploadName := fmt.Sprintf("%s%s", l.Uname, ftime)
	outputName := fmt.Sprintf("%s_%s", l.Uname, ftime)
	return uploadName, outputName
}

func ConvertFlv2Ts(middleLst []string, outputFile string, inputFile []string, l *live.LiveSnapshot) {
	if tools.Exists(outputFile) && l.DivideByTitle {
		golog.Info(fmt.Sprintf("%s[RoomID: %s] 输出文件已存在，合并到新视频中", l.Uname, l.UID))
		middleLst = append(middleLst, strings.Replace(outputFile, ".mp4", ".ts", -1))
		cmd := exec.Command("ffmpeg", "-i", outputFile, "-vcodec", "copy", "-acodec", "copy", "-vbsf", "hevc_mp4toannexb", "-y", strings.Replace(outputFile, ".mp4", ".ts", -1))
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
		golog.Info(cmd.String())
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
}

func ConvertTs2Mp4(middleLst []string, outputFile string, l *live.LiveSnapshot) {
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
	concatFilePath, _ := filepath.Abs(fmt.Sprintf("./recording/%s/tmp/concat.txt", l.Uname))
	if flag && l.Mp4Compress {
		cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", concatFilePath, "-vcodec", "hevc_nvenc", "-c:a", "copy", "-crf", "17", "-maxrate", "3M", "-bufsize", "3M", "-preset", "fast", "-y", outputFile)
		fmt.Println(cmd.String())
		golog.Info(cmd.String())
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
		cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", concatFilePath, "-vcodec", "hevc_nvenc", "-c:a", "copy", "-y", outputFile)
		fmt.Println(cmd.String())
		golog.Info(cmd.String())
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
}
