package decode

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	_ "regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/golog"
	"github.com/tidwall/gjson"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"

	"bilibili-recording-go/live"
	"bilibili-recording-go/tools"
)

var (
	decodeChan chan *decodeTuple
)

func init() {
	decodeChan = make(chan *decodeTuple, 100)
	go decodeWorker()
	go decode()
}

type decodeTuple struct {
	live       *live.LiveSnapshot
	convertFile string
	outputName string
}

func decodeWorker() {
	for {
		live := <-live.DecodeChan
		var inputFile []string
		if live.TmpFilePath == "" {
			inputFile = GetLatestFiles(live, 0)
		} else {
			inputFile = []string{live.TmpFilePath}
		}
		_, outputName := GenerateFileName(inputFile, live)
		decodeChan <- &decodeTuple{
			live:       live,
			convertFile: inputFile[0],
			outputName: outputName,
		}
	}
}

func decode() {
	for {
		dt := <-decodeChan
		golog.Info(fmt.Sprintf("%s[RoomID: %s] 开始转码", dt.live.Uname, dt.live.RoomID))
		Decode(dt.live, dt.convertFile, dt.outputName)
		golog.Info(fmt.Sprintf("%s[RoomID: %s] 结束转码", dt.live.Uname, dt.live.RoomID))
	}
}

type fileInfo struct {
	fileName       string
	lastModifyTime int64
}

// Decode 转码
func Decode(l *live.LiveSnapshot, convertFile string, outputName string) {
	pwd, _ := os.Getwd()
	inputFile := []string{convertFile}
	
	var middleLst []string
	isAppend := false;
	appendFile := ""
	appendFileName := ""

	ft := GetFileCreateTimeFromName(outputName)
	farea := strings.Split(outputName, "_")[2]
	ftitle := strings.Split(strings.TrimSuffix(outputName, ".mp4"), "_")[3]
	flst, _ := ioutil.ReadDir(fmt.Sprintf("./recording/%s/", l.Uname))
	if l.Platform == "douyin" {
		for _, f := range flst {
			if strings.HasSuffix(f.Name(), ".mp4") {
				t := GetFileCreateTimeFromName(f.Name())
				if int(ft.Sub(t).Seconds()) == 0 {
					golog.Info(fmt.Sprintf("%s[RoomID: %s] 文件已存在，进行覆盖", l.Uname, l.RoomID))
					break
				}
				tarea := strings.Split(f.Name(), "_")[2]
				ttitle := strings.Split(strings.TrimSuffix(f.Name(), ".mp4"), "_")[3]
				fmt.Println(int(ft.Sub(t).Seconds()))
				if ft.After(t) && int(ft.Sub(t).Seconds()) < 3600 * 3 && farea == tarea && ttitle == ftitle {
					isAppend = true
					appendFile, _ = filepath.Abs(fmt.Sprintf("./recording/%s/%s", l.Uname, f.Name()))
					appendFileName = f.Name()
					outputName = f.Name()
					break
				}
			}
		}
	}

	for k, f := range inputFile {
		inputFile[k], _ = filepath.Abs(f)
		middleLst = append(middleLst, strings.Replace(inputFile[k], ".flv", ".mts", -1))
	}
	
	var outputFile string
	if isAppend {
		fmt.Println(isAppend, appendFile)
		golog.Info(fmt.Sprintf("%s[RoomID: %s] %s 与已有文件创建时间差小于3小时，追加到已有文件中", l.Uname, l.RoomID, convertFile))
		outputFile = appendFile
		inputFile = append([]string{appendFile}, inputFile...)
		tmpFile, _ := filepath.Abs(fmt.Sprintf("./recording/%s/tmp/%s", l.Uname, appendFileName))
		middleLst = append([]string{strings.Replace(tmpFile, ".mp4", ".mts", -1)}, middleLst...)
	} else {
		outputFile = filepath.Join(pwd, "recording", l.Uname, fmt.Sprintf("%s.mp4", outputName))
		// l.UploadName = uploadName
		l.FilePath = outputFile
		golog.Info(fmt.Sprintf("%s[RoomID: %s] 本次转码的文件为: %s, 最终生成: %s", l.Uname, l.RoomID, strings.Join(inputFile, " "), outputFile))
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
		// ffmpeg_go.Input(concatFilePath, ffmpeg_go.KwArgs{"f": "concat", "safe": "0"}).Output(
		// 	strings.Replace(outputFile, ".mp4", ".m4a", -1), ffmpeg_go.KwArgs{"acodec": "copy", "vn": ""}).OverWriteOutput().ErrorToStdOut().Run()
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
	if !isAppend {
		outputName += ".mp4"
	}
	if l.Platform == "douyin" {
		err := os.Rename(strings.Replace(outputFile, ".mp4", ".mts", -1), fmt.Sprintf("./recording/%s/tmp/%s", l.Uname, strings.Replace(outputName, ".mp4", ".mts", -1)))
		if err != nil {
			golog.Error(err)
		}
	} else {
		err := os.Remove(strings.Replace(outputFile, ".mp4", ".mts", -1))
		if err != nil {
			golog.Error(err)
		} else {
			golog.Info(strings.Replace(outputFile, ".mp4", ".mts", -1), " has been removed")
		}
	}
	
}

func GetFileCreateTimeFromName(fileName string) time.Time {
	loc, _ := time.LoadLocation("PRC")
	t, _ := time.ParseInLocation("2006-01-02 15时04分05秒", fmt.Sprint(strings.TrimSuffix(strings.Split(fileName, "_")[1], "场"), "00秒"), loc)
	return t
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
	// fileTime := tools.GetFileCreateTime(inputFile[0])
	filesplit := strings.Split(inputFile[0], "/")
	titleWithTsp := strings.TrimSuffix(filesplit[len(filesplit)-1], ".flv")
	titleSplits := strings.Split(titleWithTsp, "_")
	areaName := titleSplits[0]
	title := strings.Join(titleSplits[1:len(titleSplits)-1], "_")
	fileTime := strings.Split(titleWithTsp, "_")[2]
	loc, _ := time.LoadLocation("PRC")
	t, _ := time.ParseInLocation("20060102150405", fileTime, loc)
	ftime := fmt.Sprintf("%s场_%s_%s", t.Format("2006-01-02 15时04分"), areaName, title)
	uploadName := fmt.Sprintf("%s%s", l.Uname, ftime)
	outputName := fmt.Sprintf("%s_%s", l.Uname, ftime)
	return uploadName, outputName
}

func ConvertFlv2Ts(middleLst []string, outputFile string, inputFile []string, l *live.LiveSnapshot) {
	var middleToFileLst []string

	for _, f := range middleLst {
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
		if _, err := os.Stat(middleLst[k]); errors.Is(err, os.ErrNotExist) {
			ffmpeg_go.Input(inputFile[k], ffmpeg_go.KwArgs{"fflags": "+discardcorrupt"}).Output(
				middleLst[k], ffmpeg_go.KwArgs{"q": "0", "vcodec": "hevc_nvenc", "c:a": "copy"}).OverWriteOutput().ErrorToStdOut().Run()
		}
	}

	ffmpeg_go.Input(concatFilePath, ffmpeg_go.KwArgs{"fflags": "+discardcorrupt", "f": "concat", "safe": "0"}).Output(
		strings.Replace(outputFile, ".mp4", ".mts", -1), ffmpeg_go.KwArgs{"c": "copy"}).OverWriteOutput().ErrorToStdOut().Run()
}

func ConvertTs2Mp4(middleLst []string, outputFile string, l *live.LiveSnapshot) {
	flag := false

	data, err := ffmpeg_go.Probe(strings.Replace(outputFile, ".mp4", ".mts", -1))
	if err == nil {
		bitRateStr := gjson.Get(data, "format.bit_rate").String()
		bitRate, _ := strconv.Atoi(bitRateStr)
		if bitRate > 3000000 {
			flag = true
		}
	}
	if flag && l.Mp4Compress {
		ffmpeg_go.Input(strings.Replace(outputFile, ".mp4", ".mts", -1), ffmpeg_go.KwArgs{"fflags": "+discardcorrupt"}).Output(
			outputFile, ffmpeg_go.KwArgs{"c": "copy", "crf": "17", "maxrate": "3M", "bufsize": "3M", "preset": "fast"}).OverWriteOutput().ErrorToStdOut().Run()
	} else {
		ffmpeg_go.Input(strings.Replace(outputFile, ".mp4", ".mts", -1), ffmpeg_go.KwArgs{"fflags": "+discardcorrupt"}).Output(
			outputFile, ffmpeg_go.KwArgs{"c": "copy"}).OverWriteOutput().ErrorToStdOut().Run()
	}
}
