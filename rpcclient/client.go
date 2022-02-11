package rpcclient

type RPCClient struct {
	serverurl string
	verbose   bool

	client Client
}

func NewRPCClient(serverurl, user, passwd string, verbose bool) *RPCClient {
	client := NewHTTPClient(serverurl)
	client.SetAuth(user, passwd)
	return &RPCClient{serverurl: serverurl, verbose: verbose, client: client}
}

func (r *RPCClient) call(srvName string, in, ret interface{}) error {
	return r.client.Call(srvName, in, ret)
}

func (r *RPCClient) Url() string {
	return r.serverurl + RPCPath
}
