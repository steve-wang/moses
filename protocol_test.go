package moses

import (
	"bytes"
	"testing"
)

func TestReqname(t *testing.T) {
	var req reqname
	req.Name = "news.163.com"
	req.Port = 3306
	buf := bytes.NewBuffer(nil)
	if err := req.Write(buf); err != nil {
		t.Fatal(err)
	}
	var req2 reqname
	if err := req2.Read(buf); err != nil {
		t.Fatal(err)
	}
	if req != req2 {
		t.Fatal("%v != %v", req, req2)
	}
	if addr := req.Address(); addr != "news.163.com:3306" {
		t.Fatalf("wrong address: %s", addr)
	}
}

func TestReqipv4(t *testing.T) {
	var req reqipv4
	req.IP = [4]byte{192, 168, 20, 1}
	req.Port = 1100
	buf := bytes.NewBuffer(nil)
	if err := req.Write(buf); err != nil {
		t.Fatal(err)
	}
	var req2 reqipv4
	if err := req2.Read(buf); err != nil {
		t.Fatal(err)
	}
	if req != req2 {
		t.Fatal("%v != %v", req, req2)
	}
	if addr := req.Address(); addr != "192.168.20.1:1100" {
		t.Fatalf("wrong address: %s", addr)
	}
}