package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func NewConsumer(url, id, topic string) (*Consumer, error) {
	consumer := &Consumer{
		Url:   url,
		ID:    id,
		Topic: topic,
	}
	if err := consumer.createConsumerGroup(); err != nil {
		return nil, err
	}
	if err := consumer.subscribeTopic(); err != nil {
		return nil, err
	}
	return consumer, nil
}

type Consumer struct {
	Url   string
	ID    string
	Topic string
}

type Message struct {
	Topic     string `json:"topic"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Partition int32  `json:"partition"`
	Offset    int64  `json:"offset"`
}

type consumeGroupPayload struct {
	Name                     string `json:"name"`
	AutoOffsetReset          string `json:"auto.offset.reset"`
	Format                   string `json:"format"`
	EnableAutoCommit         bool   `json:"auto.commit.enable"`
	FetchMinBytes            int    `json:"fetch.min.bytes"`
	ConsumerRequestTimeoutMs int    `json:"consumer.request.timeout.ms"`
}

type subscriptionPayload struct {
	Topics []string `json:"topics"`
}

func (p *Consumer) createConsumerGroup() error {
	var (
		err     error
		data    []byte
		path    = fmt.Sprintf("/consumers/%s-consumer-group-%s", p.Topic, p.ID)
		req     *http.Request
		resp    *http.Response
		payload = consumeGroupPayload{
			Name:                     p.Topic + "-consumer",
			AutoOffsetReset:          "earliest",
			Format:                   "json",
			EnableAutoCommit:         true,
			FetchMinBytes:            512,
			ConsumerRequestTimeoutMs: 30000,
		}
	)
	if data, err = json.Marshal(payload); err != nil {
		return err
	}
	if req, err = http.NewRequest("POST", p.Url+path, bytes.NewReader(data)); err != nil {
		return err
	} else {
		req.Header.Set("Content-Type", contentType)
	}
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return err
	}
	if resp.StatusCode != 200 && resp.StatusCode != 409 {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			logrus.Infof("%+v", string(bytes))
		}
		return errors.New("Error creating consumer")
	}
	defer resp.Body.Close()

	data = []byte{}
	if data, err = ioutil.ReadAll(resp.Body); err != nil {
		return err
	}
	out := &postResponse{}
	if err = json.Unmarshal(data, out); err != nil {
		return err
	} else if out.ErrorCode > 0 && out.ErrorCode != 40902 {
		return errors.New(out.Message)
	}
	return nil
}

func (p *Consumer) subscribeTopic() error {
	var (
		err     error
		data    []byte
		path    = fmt.Sprintf("/consumers/%s-consumer-group-%s/instances/%s-consumer/subscription", p.Topic, p.ID, p.Topic)
		req     *http.Request
		resp    *http.Response
		payload = subscriptionPayload{
			Topics: []string{p.Topic},
		}
	)
	if data, err = json.Marshal(payload); err != nil {
		return err
	}
	if req, err = http.NewRequest("POST", p.Url+path, bytes.NewReader(data)); err != nil {
		return err
	} else {
		req.Header.Set("Content-Type", contentType)
	}
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return errors.New("Error subscribing to the topic")
	}
	defer resp.Body.Close()

	return nil
}

func (p *Consumer) Consume() ([]*Message, error) {
	var (
		err  error
		data []byte
		path = fmt.Sprintf("/consumers/%s-consumer-group-%s/instances/%s-consumer/records", p.Topic, p.ID, p.Topic)
		req  *http.Request
		resp *http.Response
		out  = make([]*Message, 0, 0)
	)
	if req, err = http.NewRequest("GET", p.Url+path, nil); err != nil {
		return nil, err
	} else {
		req.Header.Set("Accept", contentType)
	}
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("Error consuming records from topic")
	}
	defer resp.Body.Close()

	if data, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}
