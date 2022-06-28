package liveback

import (
	"bilibili-recording-go/tools"
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Playback struct {
	Timestamp int64  `json:"timestamp"`
	Title     string `json:"title"`
	FileName  string `json:"file_name"`
}

type PlaybackList []Playback

type LivebackStatistics struct {
	MedalStatistics   []MedalInfo       `json:"medal_statistics"`
	RevenueStatistics RevenueStatistics `json:"revenue_statistics"`
}

type MedalInfo struct {
	MedalName    string           `json:"medal_name"`
	AnchorName   string           `json:"anchor_name"`
	Num          int              `json:"num"`
	LevelDetails []MedalLevelInfo `json:"level_details"`
}

type MedalLevelInfo struct {
	Level string `json:"level"`
	Num   int    `json:"num"`
}

type RevenueStatistics struct {
	SendUserNum  int        `json:"send_user_num"`
	TotalRevenue int64      `json:"total_revenue"`
	RichRanking  []RankInfo `json:"rich_ranking"`
}

type RankInfo struct {
	UID            string         `json:"uid"`
	Uname          string         `json:"uname"`
	TotalSendPrice int64          `json:"total_send_price"`
	SendGifts      []SendGiftInfo `json:"send_gifts"`
}

type SendGiftInfo struct {
	GiftName  string `json:"gift_name"`
	Num       int    `json:"num"`
	Price     int    `json:"price"`
	Timestamp int64  `json:"timestamp"`
}

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

func GetLivebackStatistics(anchorName string, brgFileName string) (LivebackStatistics, bool) {
	filePath := fmt.Sprintf("./recording/%s/brg/%s", anchorName, brgFileName)
	brgFile, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return LivebackStatistics{}, false
	}
	defer brgFile.Close()
	return LivebackStatistics{
		MedalStatistics:   GetMedalStatistics(brgFile),
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
					MedalName:  medalName,
					AnchorName: anchorName,
					Num:        0,
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
					MedalName:  medalName,
					AnchorName: anchorName,
					Num:        0,
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
					MedalName:  medalName,
					AnchorName: anchorName,
					Num:        0,
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
		m := MedalInfo{
			MedalName:  mname,
			AnchorName: minfo.AnchorName,
			Num:        minfo.Num,
		}
		levelDetails := make([]MedalLevelInfo, 0)
		for lname, lnum := range levelMap[mname] {
			levelDetails = append(levelDetails, MedalLevelInfo{
				Level: lname,
				Num:   lnum,
			})
		}
		m.LevelDetails = levelDetails
		medalLst = append(medalLst, m)
	}
	return medalLst
}

func GetRevenueStatics(brgFile *os.File) RevenueStatistics {
	uidMap := make(map[string]*RankInfo)
	rs := new(RevenueStatistics)
	brgFile.Seek(0, 0)
	scanner := bufio.NewScanner(brgFile)
	for scanner.Scan() {
		text := scanner.Text()
		var uid, uname, gift_name string
		var gift_price, gift_num int
		var timestamp int64
		isGift := false
		if strings.HasPrefix(text, "GUARD_BUY") {
			lst := strings.Split(text, ",")
			uid = strings.TrimSpace(lst[1])
			uname = strings.TrimSpace(lst[2])
			gift_name = strings.TrimSpace(lst[4])
			gift_price, _ = strconv.Atoi(strings.TrimSpace(lst[5]))
			gift_num, _ = strconv.Atoi(strings.TrimSpace(lst[6]))
			timestamp, _ = strconv.ParseInt(strings.TrimSpace(lst[7]), 10, 64)
			isGift = true
		} else if strings.HasPrefix(text, "SEND_GIFT") {
			lst := strings.Split(text, ",")
			uid = strings.TrimSpace(lst[1])
			uname = strings.TrimSpace(lst[2])
			gift_name = strings.TrimSpace(lst[6])
			gift_price, _ = strconv.Atoi(strings.TrimSpace(lst[7]))
			gift_num, _ = strconv.Atoi(strings.TrimSpace(lst[9]))
			coin_type := strings.TrimSpace(lst[8])
			if coin_type == "silver" {
				continue
			}
			timestamp, _ = strconv.ParseInt(strings.TrimSpace(lst[10]), 10, 64)
			isGift = true
		} else if strings.HasPrefix(text, "SUPER_CHAT_MESSAGE") {
			lst := strings.Split(text, ",")
			uid = strings.TrimSpace(lst[1])
			uname = strings.TrimSpace(lst[2])
			gift_name = "SUPER CHAT MESSAGE"
			gift_price, _ = strconv.Atoi(strings.TrimSpace(lst[7]))
			gift_num = 1
			timestamp, _ = strconv.ParseInt(strings.TrimSpace(lst[10]), 10, 64)
			isGift = true
		}
		if _, ok := uidMap[uid]; !ok && isGift {
			rs.SendUserNum++
			uidMap[uid] = new(RankInfo)
			uidMap[uid].UID = uid
			uidMap[uid].Uname = uname
		}
		if isGift && uid != "" {
			uidMap[uid].TotalSendPrice += int64(gift_price) * int64(gift_num)
			rs.TotalRevenue += int64(gift_price) * int64(gift_num)
			uidMap[uid].SendGifts = append(uidMap[uid].SendGifts, SendGiftInfo{
				GiftName:  gift_name,
				Price:     gift_price,
				Num:       gift_num,
				Timestamp: timestamp,
			})
		}
	}
	rs.RichRanking = make([]RankInfo, 0)
	for _, rinfo := range uidMap {
		rs.RichRanking = append(rs.RichRanking, *rinfo)
	}
	sort.Slice(rs.RichRanking, func(i, j int) bool {
		return rs.RichRanking[i].TotalSendPrice > rs.RichRanking[j].TotalSendPrice
	})
	return *rs
}
