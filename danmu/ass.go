package danmu

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/kataras/golog"
)

type Ass struct {
	uname	string
	File   string
	startT time.Time
	assHeight int
	assWeight int
	assFontSize int
	assShowTime int
	assLoc int
}

func init() {
	rand.Seed(time.Now().Unix())
}

func NewAss(uname string, file string) *Ass {
	ass := &Ass{
		uname: uname,
		File: file,
		startT: time.Now(),
		assHeight: 720,
		assWeight: 1280,
		assFontSize: 50,
		assShowTime: 5,
		assLoc: 7,
	}
	if file == "" {
		return ass
	}
	filePath := fmt.Sprintf("./recording/%s/%s.ass", uname, file)
	assFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		golog.Error(fmt.Sprintf("os.OpenFile when Init Ass File: %s", err))
	}
	defer assFile.Close()
	header := `[Script Info]
	Title: Default Ass file
	ScriptType: v4.00+
	WrapStyle: 0
	ScaledBorderAndShadow: yes
	PlayResX: ` + strconv.Itoa(ass.assHeight) + `
	PlayResY: ` + strconv.Itoa(ass.assWeight) + `
	
	[V4+ Styles]
	Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
	Style: Default,,` + strconv.Itoa(ass.assFontSize) + `,&H40FFFFFF,&H000017FF,&H80000000,&H40000000,0,0,0,0,100,100,0,0,1,4,4,` + strconv.Itoa(ass.assLoc) + `,20,20,50,1
	
	[Events]
	Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
	`
	assFile.WriteString(header)
	return ass
}

func (a *Ass) WriteDanmu(msg string) {
	if a.File == "" {
		return
	}
	if msg == "" {
		return
	}

	st := time.Since(a.startT) + time.Duration(rand.Intn(2000)) * time.Millisecond
	et := st + time.Duration(a.assShowTime) * time.Second

	var b string
	b += `Dialogue: 0,`
	b += dtos(st) + `,` + dtos(et)
	b += `,Default,,0,0,0,,{\fad(200,500)\blur3}` + msg + "\n"

	assFile, err := os.OpenFile(fmt.Sprintf("./recording/%s/%s.ass", a.uname, a.File), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		golog.Error(fmt.Sprintf("os.OpenFile when WriteDanmu: %s", err))
	}
	defer assFile.Close()
	assFile.WriteString(b)
}