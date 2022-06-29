package liveback

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/go-ego/gse"
)

type Word struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func CreateWordClod(anchorName string, brgFileName string) ([]Word, bool) {
	filePath := fmt.Sprintf("./recording/%s/brg/%s", anchorName, brgFileName)
	brgFile, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return nil, false
	}
	defer brgFile.Close()
	var seg gse.Segmenter
	seg.LoadDict("zh_s")
	seg.LoadStop("./liveback/stop_word.txt")
	wordMap := make(map[string]int)
	scanner := bufio.NewScanner(brgFile)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "DANMU_MSG") {
			lst := strings.Split(text, ",")
			if len(lst) < 9 {
				continue
			}
			jiebSlice := seg.Slice(lst[7], false)
			for _, word := range jiebSlice {
				wordMap[word]++
			}
		}

	}
	var words []Word
	for word, value := range wordMap {
		words = append(words, Word{word, value})
	}
	return words, true
}
