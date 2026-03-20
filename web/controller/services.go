package controller

import "MixFound/web/service"

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
