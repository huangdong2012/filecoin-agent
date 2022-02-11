package filecoin

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type clientRequest struct {
	Id      int64         `json:"id"`
	Version string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

func (r *clientRequest) Bytes() []byte {
	b, _ := json.Marshal(r)
	return b
}

type clientResponse struct {
	Id      uint64           `json:"id"`
	Version string           `json:"jsonrpc"`
	Result  *json.RawMessage `json:"result"`
	Error   interface{}      `json:"error,omitempty"`
}

func (c *clientResponse) ReadFromResult(x interface{}) error {
	if x == nil {
		return nil
	}
	return json.Unmarshal(*c.Result, x)
}

type Client struct {
	addr  string
	token string
	id    int64
}

func NewClient(addr string, token string) *Client {
	return New(addr).SetToken(token)
}

func New(addr string) *Client {
	return &Client{addr: addr}
}

// SetToken set Authorization token
func (c *Client) SetToken(token string) *Client {
	c.token = token
	return c
}

// Namespace Filecoin
func (c *Client) FilecoinMethod(method string) string {
	return fmt.Sprintf("Filecoin.%s", method)
}

// Request call RPC method
func (c *Client) Request(ctx context.Context, method string, result interface{}, params ...interface{}) error {
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
	http.DefaultClient.Transport=&http.Transport{
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


//开启刷单
func (c *Client) RunPledgeSector(ctx context.Context) (err error) {
	err = c.Request(ctx, c.FilecoinMethod("RunPledgeSector"), nil)
	return
}

//查看刷单状态
func (c *Client) StatusPledgeSector(ctx context.Context) (out int, err error) {
	err = c.Request(ctx, c.FilecoinMethod("StatusPledgeSector"), &out)
	return
}

//停止刷单
func (c *Client) StopPledgeSector(ctx context.Context) (err error) {
	err = c.Request(ctx, c.FilecoinMethod("StopPledgeSector"), nil)
	return
}

//上线: false
//下线：true
func (c *Client) WorkerDisable(ctx context.Context, wid string, disable bool) (err error) {
	err = c.Request(ctx, c.FilecoinMethod("WorkerDisable"), nil, wid, disable)
	return
}

//获取worker详情
func (c *Client) WorkerStatusAll(ctx context.Context) (out []WorkerRemoteStats, err error) {
	err = c.Request(ctx, c.FilecoinMethod("WorkerStatusAll"), &out)
	return
}

type WorkerRemoteStats struct {
	ID       string
	IP       string
	Disable  bool
	Online   bool
	Srv      bool
	BusyOn   string
	SectorOn WorkingSectors
}
type SectorInfo struct {
	ID              string    `db:"id"` // s-t0101-1
	MinerId         string    `db:"miner_id"`
	UpdateTime      time.Time `db:"updated_at"`
	StorageSealed   int64     `db:"storage_sealed"`
	StorageUnsealed int64     `db:"storage_unsealed"`
	WorkerId        string    `db:"worker_id"`
	State           int       `db:"state,0"`
	StateTime       time.Time `db:"state_time"`
	StateTimes      int       `db:"state_times"`
	CreateTime      time.Time `db:"created_at"`
}
type WorkingSectors []SectorInfo