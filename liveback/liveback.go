package liveback

import (
	"bilibili-recording-go/tools"
	"fmt"
	"sort"
	"strings"
)

type Playback struct {
	Timestamp int64  `json:"timestamp"`
	Title     string `json:"title"`
	FileName  string `json:"file_name"`
}

type PlaybackList []Playback



func GetAnchorLivebackList(anchorName string) PlaybackList {
	brgFiles := tools.ListDirWithoutDirPath(fmt.Sprintf("./recording/%s/brg", anchorName))
	lst := make(PlaybackList, 0)
	for _, brgFile := range brgFiles {
		lst = append(lst, Playback{
			Timestamp: tools.GetFileCreateTime(fmt.Sprintf("./recording/%s/brg/%s", anchorName, brgFile)),
			Title:     strings.Join(strings.Split(strings.TrimSuffix(brgFile, ".brg"), "_")[2:], "_"),
			FileName:  brgFile,
		})
	}
	sort.Slice(lst, func(i, j int) bool {
		return lst[i].Timestamp < lst[j].Timestamp
	})
	return lst
}
