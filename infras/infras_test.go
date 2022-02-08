package infras

import "testing"

func TestCheckSum(t *testing.T) {
	hash, err := CheckSum("/Users/huangdong/Deploy/proxy/openresty/html/download/lotus.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hash)
}
