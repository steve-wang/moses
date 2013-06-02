package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"github.com/steve-wang/moses"
	"os"
	"path/filepath"
)

type Param struct {
	User         string `json:"user"`
	Password     string `json:"password"`
	ProxyAddress string `json:"proxy"`
	Port         int    `json:"port"`
}

func md5sum(str string) string {
	h := md5.New()
	io.WriteString(h, str)
	return fmt.Sprintf("%X", h.Sum(nil))
}

func main() {
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
	connector := &moses.PrivateConnector{
		User:         param.User,
		Password:     md5sum(param.Password),
		ProxyAddress: param.ProxyAddress,
	}
	acceptor := moses.NewSOCKS5Acceptor(connector)
	srv := moses.NewServer(acceptor)
	if err := srv.Start(uint16(param.Port)); err != nil {
		return err
	}
	srv.Serve()
	return nil
}
