package moses

import (
	"bytes"
	"testing"
)

func TestPrivateReadWriter(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	rw := NewPrivateReadWriter(buf, [4]byte{0x12, 0x34, 0x56, 0x78})
	msg := "hello, world!"
	if n, err := rw.Write([]byte(msg)); err != nil {
		t.Fatal(err)
	} else if n != len(msg) {
		t.Fatalf("invalid size: %d", n)
	}
	if str := string(buf.Bytes()); str == msg {
		t.Fatal("uncoded")
	}
	data := make([]byte, len(msg))
	if err := read(rw, data); err != nil {
		t.Fatal(err)
	}
	str := string(data)
	if str != msg {
		t.Fatal(str)
	}
}