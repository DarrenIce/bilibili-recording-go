package danmu

import (
	"bilibili-recording-go/tools"
	"fmt"
	"os"

	"github.com/kataras/golog"
)

type Brg struct {
	uname string
	File string
}

func NewBrg(uname string, file string) *Brg {
	brg := &Brg{
		uname: uname,
		File: file,
	}
	if file == "" {
		return brg
	}
	filePath := fmt.Sprintf("./recording/%s/brg/%s.brg", uname, file)
	tools.Mkdir(fmt.Sprintf("./recording/%s/brg", uname))
	brgFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		golog.Error(fmt.Sprintf("os.OpenFile when Init Brg File: %s", err))
	}
	defer brgFile.Close()
	return brg
}

func (b *Brg) WriteMsg(msg string) {
	if b.File == "" {
		return
	}
	if msg == "" {
		return
	}
	brgFile, err := os.OpenFile(fmt.Sprintf("./recording/%s/brg/%s.brg", b.uname, b.File), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		golog.Error(fmt.Sprintf("os.OpenFile when WriteMsg: %s", err))
	}
	defer brgFile.Close()
	_, err = brgFile.WriteString(msg)
	if err != nil {
		golog.Error(fmt.Sprintf("WriteString when WriteMsg: %s", err))
	}
}