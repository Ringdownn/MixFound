package service

import (
	"MixFound/global"
	"MixFound/searcher"
	"MixFound/searcher/model"
)

type Index struct {
	Container *searcher.Container
}

func NewIndex() *Index {
	return &Index{
		Container: global.Container,
	}
}

func (index *Index) AddIndex(dbName string, doc *model.IndexDoc) error {
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
		if err := c.IndexDocument(doc); err != nil {
			return err
		}
	}
	return nil
}
