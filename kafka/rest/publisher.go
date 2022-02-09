package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"huangdong2012/filecoin-agent/infras"
	"io/ioutil"
	"net/http"
	"strings"
)

func NewPublisher(url string) *Publisher {
	return &Publisher{Url: url}
}

type Publisher struct {
	Url string
}

type restPubMsgReq struct {
	Records []restPubMsgRecord `json:"records"`
}

type restPubMsgRecord struct {
	Value interface{} `json:"value"`
}

type restPubMsgResp struct {
	Offsets []*struct {
		Partition int32  `json:"partition"`
		Offset    int64  `json:"offset"`
		Error     string `json:"error"`
		ErrorCode int64  `json:"error_code"`
	} `json:"offsets"`
}

func (r *Publisher) Publish(topic string, value string) (int32, int64, error) {
	url := fmt.Sprintf("%s/topics/%s", r.Url, topic)
	msg := restPubMsgReq{
		Records: []restPubMsgRecord{
			{Value: value},
		},
	}
	resp, err := http.Post(url, contentType, strings.NewReader(infras.ToJson(msg)))
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}
	pubResp := &restPubMsgResp{}
	if err = json.Unmarshal(data, pubResp); err != nil {
		return 0, 0, err
	}
	if len(pubResp.Offsets) > 0 {
		if len(pubResp.Offsets[0].Error) > 0 {
			return 0, 0, errors.New(pubResp.Offsets[0].Error)
		}
		return pubResp.Offsets[0].Partition, pubResp.Offsets[0].Offset, nil
	}
	return 0, 0, errors.New("pub response invalid: " + string(data))
}
