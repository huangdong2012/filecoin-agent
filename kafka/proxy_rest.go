package kafka

import "huangdong2012/filecoin-agent/model"

var (
	Rest = &restProxy{}
)

type restProxy struct {
}

func (p *restProxy) Publish(topic string, value string) (int32, int64, error) {

	return 0, 0, nil
}

func (p *restProxy) Consume(id, topic string, stopC <-chan bool, offsetOpt *OffsetOption) (<-chan *model.CommandRequest, error) {

	return nil, nil
}
