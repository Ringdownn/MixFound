package controller

import "MixFound/services/search-engine/internal/web/service"

var srv *Services

type Services struct {
	Base     *service.Base
	Database *service.Database
	Index    *service.Index
}

func NewServices() {
	srv = &Services{
		Base:     service.NewBase(),
		Database: service.NewDatabase(),
		Index:    service.NewIndex(),
	}
}
