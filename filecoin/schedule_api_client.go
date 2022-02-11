package filecoin


import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type scheduleClientRequest struct {
	Id      int64         `json:"id"`
	Version string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

func (r *scheduleClientRequest) Bytes() []byte {
	b, _ := json.Marshal(r)
	return b
}

type scheduleClientResponse struct {
	Id      uint64           `json:"id"`
	Version string           `json:"jsonrpc"`
	Result  *json.RawMessage `json:"result"`
	Error   interface{}      `json:"error,omitempty"`
}

func (c *scheduleClientResponse) ReadFromResult(x interface{}) error {
	if x == nil {
		return nil
	}
	return json.Unmarshal(*c.Result, x)
}

type ScheduleClient struct {
	addr  string
	token string
	id    int64
}

func NewScheduleClient(addr string, token string) *ScheduleClient {
	return NewSchedule(addr).SetToken(token)
}

func NewSchedule(addr string) *ScheduleClient {
	return &ScheduleClient{addr: addr}
}

// SetToken set Authorization token
func (c *ScheduleClient) SetToken(token string) *ScheduleClient {
	c.token = token
	return c
}

// Namespace Filecoin
func (c *ScheduleClient) FilecoinMethod(method string) string {
	return fmt.Sprintf("Filecoin.%s", method)
}

// Request call RPC method
func (c *ScheduleClient) Request(ctx context.Context, method string, result interface{}, params ...interface{}) error {
	request := &clientRequest{
		Id:      1,
		Version: "2.0",
		Method:  method,
		Params:  params,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.addr, bytes.NewReader(request.Bytes()))
	if err != nil {
		return err
	}
	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}
	//todo 跳过https
	http.DefaultClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}
	//check http status
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP STATUS NOT OK [%v][%v]", rsp.Status, string(body))
	}

	response := &clientResponse{}
	if err := json.Unmarshal(body, response); err != nil {
		return err
	}
	if response.Error != nil {
		return fmt.Errorf("jsonrpc call: %v", response.Error)
	}
	if response.Result == nil {
		return nil
	}

	return response.ReadFromResult(result)
}



func (c *ScheduleClient) GetWorkerBusyTask(ctx context.Context, wid string) (out int, err error) {
	err = c.Request(ctx, c.FilecoinMethod("GetWorkerBusyTask"), &out, wid)
	return
}

//获取worker详情
func (c *ScheduleClient) RequestDisableWorker(ctx context.Context, wid string) (err error) {
	err = c.Request(ctx, c.FilecoinMethod("RequestDisableWorker"), nil, wid)
	return err
}
