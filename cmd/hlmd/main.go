package main

import (
	"fmt"
	"huangdong2012/filecoin-agent/hlmd"
	"os"
)

func init() {
	hlmd.Init(func(o *hlmd.Option) {
		o.Verbose = true
	})
}

func main() {
	if len(os.Args) >= 3 {
		if err := hlmd.Execute(os.Args[1:]); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("finish")
		}
	}
}
