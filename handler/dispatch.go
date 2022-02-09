package handler

import "huangdong2012/filecoin-agent/model"

var (
	Dispatch = &dispatchHandler{}
)

type dispatchHandler struct {
}

func (h *dispatchHandler) Handle(msg *model.CommandRequest) (*model.CommandResponse, error) {
	//todo...
	return nil, nil
}
