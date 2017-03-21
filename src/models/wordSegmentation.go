package models

import (
	"strings"
	"github.com/yanyiwu/gojieba"
)

type WordSegmentation struct {
	Gojieba		*gojieba.Jieba
}

//新建实例
func NewWordSegmentation() *WordSegmentation {
	wordSegmentation := &WordSegmentation{}
	wordSegmentation.Gojieba = gojieba.NewJieba()
	return wordSegmentation
}

/**
全模式
 */
func (wc *WordSegmentation) CutAll(s string) string {

	var words []string
	words = wc.Gojieba.CutAll(s)
	return strings.Join(words, "/")
}

/**
精确模式
 */
func (wc *WordSegmentation) Cut(s string) string {

	var words []string
	use_hmm := true
	words = wc.Gojieba.Cut(s, use_hmm)
	return strings.Join(words, "/")
}

/**
搜索引擎模式
 */
func (wc *WordSegmentation) CutForSearch(s string) string {

	var words []string
	use_hmm := true
	words = wc.Gojieba.CutForSearch(s, use_hmm)
	return strings.Join(words, "/")
}

func (wc *WordSegmentation) Tag(s string) string {

	var words []string
	words = wc.Gojieba.Tag(s)
	return strings.Join(words, ",")
}

//func (wc *WordSegmentation) Tokenize(s string) string {
//
//	use_hmm := true
//	wordinfos := wc.Gojieba.Tokenize(s, gojieba.SearchMode, !use_hmm)
//	return json.Marshal(wordinfos)
//}
//
//func (wc *WordSegmentation) Extract(s string, len int32) string {
//
//	keywords := wc.Gojieba.ExtractWithWeight(s, len)
//	return json.Marshal(keywords)
//}