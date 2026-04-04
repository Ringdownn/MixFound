package service

import (
	"MixFound/services/search-engine/internal/global"
	"MixFound/services/search-engine/internal/searcher"
	"MixFound/services/search-engine/internal/searcher/model"
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
