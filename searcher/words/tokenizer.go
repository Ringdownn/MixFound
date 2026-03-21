package words

import (
	"MixFound/searcher/utils"
	"embed"
	"strings"

	"github.com/wangbin/jiebago"
)

var (
	//go:embed data/*.txt
	dictionaryFS embed.FS
)

type Tokenizer struct {
	seg jiebago.Segmenter
}

func NewTokenizer(dictionaryPath string) *Tokenizer {
	file, err := dictionaryFS.Open("data/dictionary.txt")
	if err != nil {
		panic(err)
	}
	utils.ReleaseAssets(file, dictionaryPath)
	tokenizer := new(Tokenizer)
	err = tokenizer.seg.LoadDictionary(dictionaryPath)
	if err != nil {
		panic(err)
	}
	return tokenizer
}

func (t *Tokenizer) Cut(text string) []string {
	//全部小写
	text = strings.ToLower(text)
	//去除空格
	text = utils.RemoveSpace(text)
	//去除标点符号
	text = utils.RemovePunctuation(text)

	var wordsMap = make(map[string]struct{})

	//开启分词，返回一个channel
	resultChan := t.seg.CutForSearch(text, true)
	var wordsSlice []string
	for {
		//从channel中取出词
		w, ok := <-resultChan
		if !ok {
			break
		}
		_, found := wordsMap[w]
		if !found {
			//标记已经存在并加入到结果
			wordsMap[w] = struct{}{}
			wordsSlice = append(wordsSlice, w)
		}

	}
	return wordsSlice
}
