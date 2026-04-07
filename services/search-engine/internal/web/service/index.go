package service

import (
	"MixFound/services/search-engine/internal/global"
	"MixFound/services/search-engine/internal/queue/producer"
	"MixFound/services/search-engine/internal/searcher"
	"MixFound/services/search-engine/internal/searcher/model"
	"fmt"
	"strconv"
)

type Index struct {
	Container       *searcher.Container
	taggingProducer *producer.TaggingProducer
}

// NewIndex 初始化 Index 服务，注入 taggingProducer
func NewIndex(taggingProducer *producer.TaggingProducer) *Index {
	return &Index{
		Container:       global.Container,
		taggingProducer: taggingProducer,
	}
}

func (index *Index) AddIndex(dbName string, doc *model.IndexDoc) error {
	if doc.ImageURL != "" {
		if err := index.taggingProducer.PublishTaggingTask(dbName, doc); err != nil {
			fmt.Print("发送到打标队列失败" + strconv.Itoa(int(doc.Id)))
			return index.Container.GetDataBase(dbName).IndexDocument(doc)
		}
		return nil
	}
	return index.Container.GetDataBase(dbName).IndexDocument(doc)
}

func (index *Index) RemoveIndex(dbName string, doc *model.RemoveIndexModel) error {
	db := index.Container.GetDataBase(dbName)
	if err := db.RemoveIndex(doc.Id); err != nil {
		return err
	}
	return nil
}

func (index *Index) BatchAddIndex(dbName string, docs []*model.IndexDoc) error {
	c := index.Container.GetDataBase(dbName)
	for _, doc := range docs {
		if doc.ImageURL != "" {
			if err := index.taggingProducer.PublishTaggingTask(dbName, doc); err != nil {
				fmt.Print("发送到打标队列失败" + strconv.Itoa(int(doc.Id)))
				if err := c.IndexDocument(doc); err != nil {
					return err
				}
			}
			continue
		}
		if err := c.IndexDocument(doc); err != nil {
			return err
		}
	}
	return nil
}
