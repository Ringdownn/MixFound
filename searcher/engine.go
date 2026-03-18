package searcher

import (
	"MixFound/searcher/model"
	"MixFound/searcher/storage"
	"MixFound/searcher/words"
	"log"
	"sync"
)

type Engine struct {
	IndexPath string
	Option    *Option

	InvertedIndexStorages []*storage.LevelDBStorage
	PositiveIndexStorages []*storage.LevelDBStorage
	DocIndexStorages      []*storage.LevelDBStorage

	sync.Mutex
	sync.WaitGroup
	addDocumentWorkerChan []chan *model.IndexDoc //添加索引的通道
	IsDebug               bool
	Tokenizer             *words.Tokenizer //分词器
	DatabaseName          string           //数据库名

	Shard     int   //分片数
	Timeout   int64 //超时时间
	BufferNum int   //分片缓冲数

	documentCount int64
}

type Option struct {
	InvertedIndexName string //倒排索引路径
	PositiveIndexName string //正排索引路径
	DocIndexName      string //文档路径
}

func (e *Engine) Init() {
	//加等待
	e.Add(1)
	defer e.Done()

	if e.Option == nil {
		e.Option = e.GetOpions()
	}
	if e.Timeout == 0 {
		e.Timeout = 30
	}

	e.documentCount = -1
	log.Print("chain num:")
}

func (e *Engine) GetOpions() *Option {
	return &Option{
		InvertedIndexName: "inverted_index",
		PositiveIndexName: "positive_index",
		DocIndexName:      "docs",
	}
}
