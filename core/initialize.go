package core

import (
	"MixFound/global"
	"MixFound/searcher"
	"MixFound/searcher/words"
	"fmt"
)

func NewContainer(tokenizer *words.Tokenizer) *searcher.Container {
	container := &searcher.Container{
		Dir:       global.CONFIG.Data,
		Debug:     global.CONFIG.Debug,
		Shard:     global.CONFIG.Shard,
		Timeout:   global.CONFIG.Timeout,
		BufferNum: global.CONFIG.BufferNum,
		Tokenizer: tokenizer,
	}
	if err := container.Init(); err != nil {
		panic(err)
	}

	return container
}

func NewTokenizer(dictionaryPath string) *words.Tokenizer {
	return words.NewTokenizer(dictionaryPath)
}

func Initialize() {
	//加载配置
	global.CONFIG = Parse()

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("panic:%s\n", r)
		}
	}()

	//加载词典目录，初始化分词器
	tokenizer := NewTokenizer(global.CONFIG.Dictionary)
	//初始化容器
	global.Container = NewContainer(tokenizer)

	//TODO初始化业务逻辑
	//TODO注册路由
	//TODO启动服务
	//TODO优雅关机
}
