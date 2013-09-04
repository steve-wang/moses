package main

import (
	"encoding/json"
	"fmt"
	"github.com/steve-wang/moses"
	"os"
	"path/filepath"
	"runtime"
)

type Param struct {
	User      string `json:"user"`
	Password  string `json:"password"`
	LocalPort uint16 `json:"port_local"`
	ProxyHost string `json:"host_proxy"`
	ProxyPort uint16 `json:"port_proxy"`
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	var param Param
	load := func(param *Param) error {
		f, err := os.Open(filepath.Join(wd, "moses_client.json"))
		if err != nil {
			return err
		}
		defer f.Close()
		return json.NewDecoder(f).Decode(param)
	}
	if err := load(&param); err != nil {
		fmt.Println(err)
		return
	}
	if err := run(&param); err != nil {
		fmt.Println(err)
		return
	}
}

func run(param *Param) error {
	connector := &moses.SOCKS5Connector{
		User:     param.User,
		Password: param.Password,
		Host:     param.ProxyHost,
		Port:     param.ProxyPort,
	}
	//acceptor := moses.NewSOCK5Acceptor(nil)
	acceptor := moses.SOCKS4Acceptor{}
	srv := moses.NewServer(acceptor, connector)
	if err := srv.Start(param.LocalPort); err != nil {
		return err
	}
	srv.Serve()
	return nil
}
