package live

import (
	"os/exec"
	"sync"

	"bilibili-recording-go/config"
)

// Live 主类
type Live struct {
	rooms map[string]config.RoomConfigInfo

	stop          chan struct{}
	recordChannel chan config.RoomConfigInfo
	unliveChannel chan string
	decodeChannel chan string
	uploadChannel chan string
	downloadCmd   *exec.Cmd
	state         sync.Map
	lock          *sync.Mutex
}

const (
	iinit   uint32 = iota
	start          // 开始监听
	running        // 正在录制
	waiting        // 在unlive中从running转移到waiting，如果不在录制时间段内就跳到waiting
	decoding
	decodeEnd
	updateWait
	updating
	updateEnd
	stop
	// 转码上传完成后，从waiting回到start
)

// Init Live
func (l *Live) Init() {
	l.rooms = make(map[string]config.RoomConfigInfo)
	l.stop = make(chan struct{})
	l.recordChannel = make(chan config.RoomConfigInfo)
	l.unliveChannel = make(chan string)
	l.decodeChannel = make(chan string)
	l.uploadChannel = make(chan string)
	l.lock = new(sync.Mutex)

	config, _ := config.LoadConfig()
	for _, v := range config.Live {
		l.state.Store(v.RoomID, iinit)
	}

	go l.recordWorker()
	go l.decodeWorker()

}

// AddRoom ADD
func (l *Live) AddRoom(info config.RoomConfigInfo) {
	l.recordChannel <- info
}
