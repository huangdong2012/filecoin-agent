package handler

import (
	"encoding/json"
	"fmt"
	"grandhelmsman/filecoin-agent/infras"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"
)

var (
	Offset = &offsetHandler{
		offsets: &sync.Map{},
	}
)

type offsetHandler struct {
	offsets *sync.Map
	file    string
}

func (h *offsetHandler) init() {
	//1.load
	h.file = filepath.Join(opt.ProjectRoot, "upgrade", "offset.json")
	if data, err := ioutil.ReadFile(h.file); err == nil {
		dict := make(map[int32]int64)
		if err = json.Unmarshal(data, &dict); err == nil {
			for k, v := range dict {
				h.offsets.Store(k, v)
			}
		}
	}

	//2.write loop
	go h.loop()
}

func (h *offsetHandler) get(partition int32) int64 {
	offset := int64(0)
	h.offsets.Range(func(k, v interface{}) bool {
		if k.(int32) == partition {
			offset = v.(int64)
			return false
		}
		return true
	})
	return offset
}

func (h *offsetHandler) set(partition int32, offset int64) {
	h.offsets.Store(partition, offset)
}

func (h *offsetHandler) loop() {
	for range time.Tick(time.Second * 10) {
		dict := make(map[int32]int64)
		h.offsets.Range(func(k, v interface{}) bool {
			dict[k.(int32)] = v.(int64)
			return true
		})

		if err := ioutil.WriteFile(h.file, []byte(infras.ToJson(dict)), 0644); err != nil {
			fmt.Println("write offset.json error:", err)
		}
	}
}
