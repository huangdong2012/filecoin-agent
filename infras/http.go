package infras

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadToDir(url, user, pwd, dir string) (string, error) {
	var (
		err    error
		req    *http.Request
		resp   *http.Response
		name   = filepath.Base(url)
		target = filepath.Join(dir, name)
	)
	if req, err = http.NewRequest("GET", url, bytes.NewBuffer([]byte{})); err != nil {
		return "", err
	}
	if len(user) > 0 || len(pwd) > 0 {
		req.SetBasicAuth(user, pwd)
	}
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Error Status of HTTP: %v", resp.StatusCode)
	}
	if err = CopyToFile(target, resp.Body); err != nil {
		return "", err
	}

	return target, nil
}

func CopyToFile(target string, body io.Reader) error {
	file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := io.Copy(file, body); err != nil {
		return err
	}
	return nil
}
