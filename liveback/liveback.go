package liveback

import (
	"bilibili-recording-go/tools"
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Playback struct {
	Timestamp int64 `json:"timestamp"`
	Title    string `json:"title"`
	FileName string `json:"file_name"`
}

type PlaybackList []Playback

type LivebackStatistics struct {
	MedalStatistics []MedalInfo `json:"medal_statistics"`
	RevenueStatistics RevenueStatistics `json:"revenue_statistics"`
}

type MedalInfo struct {
	MedalName string `json:"medal_name"`
	AnchorName string `json:"anchor_name"`
	Num 	int    `json:"num"`
	LevelDetails []MedalLevelInfo `json:"level_details"`
}

type MedalLevelInfo struct {
	Level string `json:"level"`
	Num   int    `json:"num"`
}

type RevenueStatistics struct {
	SendUserNum int `json:"send_user_num"`
	TotalRevenue int64 `json:"total_revenue"`
	RichRanking []RankInfo `json:"rich_ranking"`
}

type RankInfo struct {
	UID string `json:"uid"`
	Uname string `json:"uname"`
	SendGifts []SendGiftInfo `json:"send_gifts"`
}

type SendGiftInfo struct {
	GiftName string `json:"gift_name"`
	Num int `json:"num"`
	Price int `json:"price"`
	Details []SendDetail `json:"details"`
}

type SendDetail struct {
	Num int `json:"num"`
	Timestamp int64 `json:"timestamp"`
}

func GetAnchorLivebackList(anchorName string) PlaybackList {
	brgFiles := tools.ListDirWithoutDirPath(fmt.Sprintf("./recording/%s/brg", anchorName))
	lst := make(PlaybackList, 0)
	for _, brgFile := range brgFiles {
		lst = append(lst, Playback{
			Timestamp: tools.GetFileCreateTime(fmt.Sprintf("./recording/%s/brg/%s", anchorName, brgFile)),
			Title:    strings.Join(strings.Split(strings.TrimSuffix(brgFile, ".brg"), "_")[2:], "_"),
			FileName: brgFile,
		})
	}
	sort.Slice(lst, func(i, j int) bool {
		return lst[i].Timestamp < lst[j].Timestamp
	})
	return lst
}

func GetLivebackStatistics(anchorName string, brgFileName string) (LivebackStatistics, bool) {
	filePath := fmt.Sprintf("./recording/%s/brg/%s", anchorName, brgFileName)
	brgFile, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return LivebackStatistics{}, false
	}
	defer brgFile.Close()
	return LivebackStatistics{
		MedalStatistics: GetMedalStatistics(brgFile),
		RevenueStatistics: GetRevenueStatics(brgFile),
	}, true
}

func GetMedalStatistics(brgFile *os.File) []MedalInfo {
	uidMap := make(map[string]bool)
	medalMap := make(map[string]*MedalInfo)
	levelMap := make(map[string]map[string]int)
	scanner := bufio.NewScanner(brgFile)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "DANMU_MSG") {
			lst := strings.Split(text, ",")
			uid := strings.TrimSpace(lst[1])
			if _, ok := uidMap[uid]; ok {
				continue
			}
			uidMap[uid] = true
			medalName := strings.TrimSpace(lst[4])
			if medalName == "" {
				continue
			}
			anchorName := strings.TrimSpace(lst[6])
			medalLevel := strings.TrimSpace(lst[5])
			if _, ok := medalMap[medalName]; !ok {
				medalMap[medalName] = &MedalInfo{
					MedalName: medalName,
					AnchorName: anchorName,
					Num: 0,
				}
				levelMap[medalName] = make(map[string]int)
			}
			medalMap[medalName].Num++
			if _, ok := levelMap[medalName][medalLevel]; !ok {
				levelMap[medalName][medalLevel] = 0
			} 
			levelMap[medalName][medalLevel]++
		} else if strings.HasPrefix(text, "SEND_GIFT") {
			lst := strings.Split(text, ",")
			uid := strings.TrimSpace(lst[1])
			if _, ok := uidMap[uid]; ok {
				continue
			}
			uidMap[uid] = true
			medalName := strings.TrimSpace(lst[3])
			if medalName == "" {
				continue
			}
			anchorName := strings.TrimSpace(lst[5])
			medalLevel := strings.TrimSpace(lst[4])
			if _, ok := medalMap[medalName]; !ok {
				medalMap[medalName] = &MedalInfo{
					MedalName: medalName,
					AnchorName: anchorName,
					Num: 0,
				}
				levelMap[medalName] = make(map[string]int)
			}
			medalMap[medalName].Num++
			if _, ok := levelMap[medalName][medalLevel]; !ok {
				levelMap[medalName][medalLevel] = 0
			} 
			levelMap[medalName][medalLevel]++
		} else if strings.HasPrefix(text, "SUPER_CHAT_MESSAGE") {
			lst := strings.Split(text, ",")
			uid := strings.TrimSpace(lst[1])
			if _, ok := uidMap[uid]; ok {
				continue
			}
			uidMap[uid] = true
			medalName := strings.TrimSpace(lst[4])
			if medalName == "" {
				continue
			}
			anchorName := strings.TrimSpace(lst[6])
			medalLevel := strings.TrimSpace(lst[5])
			if _, ok := medalMap[medalName]; !ok {
				medalMap[medalName] = &MedalInfo{
					MedalName: medalName,
					AnchorName: anchorName,
					Num: 0,
				}
				levelMap[medalName] = make(map[string]int)
			}
			medalMap[medalName].Num++
			if _, ok := levelMap[medalName][medalLevel]; !ok {
				levelMap[medalName][medalLevel] = 0
			} 
			levelMap[medalName][medalLevel]++
		}
	}
	medalLst := make([]MedalInfo, 0)
	for mname, minfo := range medalMap {
		m := MedalInfo {
			MedalName: mname,
			AnchorName: minfo.AnchorName,
			Num: minfo.Num,
		}
		levelDetails := make([]MedalLevelInfo, 0)
		for lname, lnum := range levelMap[mname] {
			levelDetails = append(levelDetails, MedalLevelInfo{
				Level: lname,
				Num: lnum,
			})
		}
		m.LevelDetails = levelDetails
		medalLst = append(medalLst, m)
	}
	return medalLst
}

func GetRevenueStatics(brgFile *os.File) RevenueStatistics {
	// uidMap := make(map[string]string)
	// scanner := bufio.NewScanner(brgFile)
	// for scanner.Scan() {
	// 	text := scanner.Text()
	// }
	return RevenueStatistics{}
}