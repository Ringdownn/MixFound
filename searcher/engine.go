package searcher

import (
	"MixFound/searcher/model"
	"MixFound/searcher/storage"
	"MixFound/searcher/utils"
	"MixFound/searcher/words"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

type Engine struct {
	IndexPath string
	Option    *Option

	invertedIndexStorages []*storage.LevelDBStorage
	positiveIndexStorages []*storage.LevelDBStorage
	docStorages           []*storage.LevelDBStorage

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

//初始化引擎

// InitOption 初始化默认配置
func (e *Engine) InitOption(option *Option) {
	if option == nil {
		option = e.GetOptions()
	}
	if e.Shard <= 0 {
		e.Shard = 10
	}
	if e.BufferNum <= 0 {
		e.BufferNum = 1000
	}

	//初始化引擎
	e.Init()
}

func (e *Engine) Init() {
	//加等待
	e.Add(1)
	defer e.Done()

	if e.Option == nil {
		e.Option = e.GetOptions()
	}
	if e.Timeout == 0 {
		e.Timeout = 30
	}

	e.documentCount = -1
	log.Print("chain num:", e.BufferNum*e.Shard)

	e.addDocumentWorkerChan = make([]chan *model.IndexDoc, e.Shard)

	for shard := 0; shard < e.Shard; shard++ {
		//初始化chan
		worker := make(chan *model.IndexDoc, e.BufferNum)
		e.addDocumentWorkerChan[shard] = worker

		go e.DocumentWorkerExec(worker)

		//初始化文档储存
		s, err := storage.NewLevelDBStorage(e.getFilePath(fmt.Sprintf("%s_%d", e.Option.DocIndexName, shard)), e.Timeout)
		if err != nil {
			panic(err)
		}
		e.docStorages = append(e.docStorages, s)

		//初始化key储存
		ks, kerr := storage.NewLevelDBStorage(e.getFilePath(fmt.Sprintf("%s_%d", e.Option.InvertedIndexName, shard)), e.Timeout)
		if kerr != nil {
			panic(kerr)
		}
		e.invertedIndexStorages = append(e.invertedIndexStorages, ks)

		//id和key映射
		iks, ikerr := storage.NewLevelDBStorage(e.getFilePath(fmt.Sprintf("%s_%d", e.Option.PositiveIndexName, shard)), e.Timeout)
		if ikerr != nil {
			panic(ikerr)
		}
		e.positiveIndexStorages = append(e.positiveIndexStorages, iks)
	}
	go e.automaticGC()
}

// 每十秒进行一次手动垃圾回收
func (e *Engine) automaticGC() {
	ticker := time.NewTicker(time.Second * time.Duration(10))
	for {
		<-ticker.C
		runtime.GC()
	}
}

func (e *Engine) GetOptions() *Option {
	return &Option{
		InvertedIndexName: "inverted_index",
		PositiveIndexName: "positive_index",
		DocIndexName:      "docs",
	}
}

func (e *Engine) getFilePath(filename string) string {
	return e.IndexPath + string(os.PathSeparator) + filename
}

//计算分片逻辑

// GetShard 计算分片
func (e *Engine) GetShard(id uint32) int {
	return int(id % uint32(e.Shard))
}

// GetShardByWord 通过string计算分片
func (e *Engine) GetShardByWord(word string) int {
	return int(utils.StringToInt(word) % uint32(e.Shard))
}

//文件处理流程

// DocumentWorkerExec 添加文档队列
func (e *Engine) DocumentWorkerExec(worker chan *model.IndexDoc) {
	for {
		doc := <-worker
		e.AddDocument(doc)
	}
}

// AddDocument 分词索引
func (e *Engine) AddDocument(index *model.IndexDoc) {
	e.Wait()
	text := index.Text

	splitWords := e.Tokenizer.Cut(text)

	id := index.Id
	//检测是否需要更新倒排索引，id不存在/索引不存在
	inserts, needUpdateInverted := e.optimizeIndex(id, splitWords)
	if needUpdateInverted {
		for _, word := range inserts {
			e.addInvertedIndex(word, id)
		}
	}

	e.addPositiveIndex(index, splitWords)
}

//添加正排索引
func (e *Engine) addPositiveIndex(word *model.IndexDoc, splitWords []string) {

}

//添加倒排索引
func (e *Engine) addInvertedIndex(word string, id uint32) {

}

//检测是否需要更新
func (e *Engine) optimizeIndex(id uint32, splitWords []string) ([]string, bool) {

}

// Close 关闭引擎
func (e *Engine) Close() {
	e.Lock()
	defer e.Unlock()

	for i := 0; i < e.Shard; i++ {
		_ = e.docStorages[i].Close()
		_ = e.invertedIndexStorages[i].Close()
		_ = e.positiveIndexStorages[i].Close()
	}
}

// Drop 删除文件
func (e *Engine) Drop() error {
	e.Lock()
	defer e.Unlock()

	if err := os.RemoveAll(e.IndexPath); err != nil {
		return err
	}

	e.docStorages = make([]*storage.LevelDBStorage, 0)
	e.invertedIndexStorages = make([]*storage.LevelDBStorage, 0)
	e.positiveIndexStorages = make([]*storage.LevelDBStorage, 0)
	return nil
}
