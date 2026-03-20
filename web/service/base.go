package service

import (
	"MixFound/global"
	"MixFound/searcher"
	"MixFound/searcher/model"
)

type Base struct {
	Container *searcher.Container
}

func NewBase() *Base {
	return &Base{
		Container: global.Container,
	}
}

func (b *Base) Query(request *model.SearchRequest) (*model.SearchResult, error) {
	return b.Container.GetDataBase(request.Database).MultiSearch(request)
}
